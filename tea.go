// Package tea provides a framework for building rich terminal user interfaces
// based on the paradigms of The Elm Architecture. It's well-suited for simple
// and complex terminal applications, either inline, full-window, or a mix of
// both. It's been battle-tested in several large projects and is
// production-ready.
//
// A tutorial is available at https://github.com/charmbracelet/bubbletea/tree/master/tutorials
//
// Example programs can be found at https://github.com/charmbracelet/bubbletea/tree/master/examples
package tea

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"golang.org/x/sync/errgroup"
)

// ErrProgramKilled is returned by [Program.Run] when the program got killed.
var ErrProgramKilled = errors.New("program was killed")

// Msg contain data from the result of a IO operation. Msgs trigger the update
// function and, henceforth, the UI.
type Msg interface{}

// Model contains the program's state as well as its core functions.
type Model interface {
	// Init is the first function that will be called. It returns an optional
	// initial command. To not perform an initial command return nil.
	Init() (Model, Cmd)

	// Update is called when a message is received. Use it to inspect messages
	// and, in response, update the model and/or send a command.
	Update(Msg) (Model, Cmd)

	// View renders the program's UI, which is just a string. The view is
	// rendered after every Update.
	View() string
}

// Cmd is an IO operation that returns a message when it's complete. If it's
// nil it's considered a no-op. Use it for things like HTTP requests, timers,
// saving and loading from disk, and so on.
//
// Note that there's almost never a reason to use a command to send a message
// to another part of your program. That can almost always be done in the
// update function.
type Cmd func() Msg

type inputType int

const (
	defaultInput inputType = iota
	ttyInput
	customInput
)

// String implements the stringer interface for [inputType]. It is inteded to
// be used in testing.
func (i inputType) String() string {
	return [...]string{
		"default input",
		"tty input",
		"custom input",
	}[i]
}

// Options to customize the program during its initialization. These are
// generally set with ProgramOptions.
//
// The options here are treated as bits.
type startupOptions int16

func (s startupOptions) has(option startupOptions) bool {
	return s&option != 0
}

const (
	withAltScreen startupOptions = 1 << iota
	withMouseCellMotion
	withMouseAllMotion
	withoutSignalHandler
	// Catching panics is incredibly useful for restoring the terminal to a
	// usable state after a panic occurs. When this is set, Bubble Tea will
	// recover from panics, print the stack trace, and disable raw mode. This
	// feature is on by default.
	withoutCatchPanics
	withoutBracketedPaste
	withReportFocus
	withKittyKeyboard
	withModifyOtherKeys
	withWindowsInputMode
	withoutGraphemeClustering
	withColorProfile
	withKeyboardEnhancements
	withGraphemeClustering

	withFerociousRenderer
)

// channelHandlers manages the series of channels returned by various processes.
// It allows us to wait for those processes to terminate before exiting the
// program.
type channelHandlers struct {
	handlers []chan struct{}
	mu       sync.RWMutex
}

// Adds a channel to the list of handlers. We wait for all handlers to terminate
// gracefully on shutdown.
func (h *channelHandlers) add(ch chan struct{}) {
	h.mu.Lock()
	h.handlers = append(h.handlers, ch)
	h.mu.Unlock()
}

// shutdown waits for all handlers to terminate.
func (h *channelHandlers) shutdown() {
	var wg sync.WaitGroup

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, ch := range h.handlers {
		wg.Add(1)
		go func(ch chan struct{}) {
			<-ch
			wg.Done()
		}(ch)
	}
	wg.Wait()
}

// Program is a terminal user interface.
type Program struct {
	initialModel Model

	// handlers is a list of channels that need to be waited on before the
	// program can exit.
	handlers channelHandlers

	// Configuration options that will set as the program is initializing,
	// treated as bits. These options can be set via various ProgramOptions.
	startupOptions startupOptions

	// startupTitle is the title that will be set on the terminal when the
	// program starts.
	startupTitle string

	inputType inputType

	ctx    context.Context
	cancel context.CancelFunc

	msgs         chan Msg
	errs         chan error
	finished     chan struct{}
	shutdownOnce sync.Once

	profile colorprofile.Profile // the terminal color profile

	// where to send output, this will usually be os.Stdout.
	output *safeWriter

	// ttyOutput is null if output is not a TTY.
	ttyOutput           term.File
	previousOutputState *term.State
	renderer            renderer

	// the environment variables for the program, defaults to os.Environ().
	environ []string

	// where to read inputs from, this will usually be os.Stdin.
	input io.Reader
	// ttyInput is null if input is not a TTY.
	ttyInput              term.File
	previousTtyInputState *term.State
	inputReader           *driver
	traceInput            bool // true if input should be traced
	readLoopDone          chan struct{}

	// modes keeps track of terminal modes that have been enabled or disabled.
	modes         map[string]bool
	ignoreSignals uint32

	filter func(Model, Msg) Msg

	// fps is the frames per second we should set on the renderer, if
	// applicable,
	fps int

	// ticker is the ticker that will be used to write to the renderer.
	ticker *time.Ticker

	// once is used to stop the renderer.
	once sync.Once

	// rendererDone is used to stop the renderer.
	rendererDone chan struct{}

	keyboard keyboardEnhancements

	// When a program is suspended, the terminal state is saved and the program
	// is paused. This saves the terminal colors state so they can be restored
	// when the program is resumed.
	setBg, setFg, setCc color.Color

	// exp stores program experimental features.
	exp experimentalOptions
}

// Quit is a special command that tells the Bubble Tea program to exit.
func Quit() Msg {
	return QuitMsg{}
}

// QuitMsg signals that the program should quit. You can send a QuitMsg with
// Quit.
type QuitMsg struct{}

// Suspend is a special command that tells the Bubble Tea program to suspend.
func Suspend() Msg {
	return SuspendMsg{}
}

// SuspendMsg signals the program should suspend.
// This usually happens when ctrl+z is pressed on common programs, but since
// bubbletea puts the terminal in raw mode, we need to handle it in a
// per-program basis.
// You can send this message with Suspend.
type SuspendMsg struct{}

// ResumeMsg can be listen to to do something once a program is resumed back
// from a suspend state.
type ResumeMsg struct{}

// NewProgram creates a new Program.
func NewProgram(model Model, opts ...ProgramOption) *Program {
	p := &Program{
		initialModel: model,
		msgs:         make(chan Msg),
		rendererDone: make(chan struct{}),
		modes:        make(map[string]bool),
		exp:          experimentalOptions{},
	}

	// Apply all options to the program.
	for _, opt := range opts {
		opt(p)
	}

	// A context can be provided with a ProgramOption, but if none was provided
	// we'll use the default background context.
	if p.ctx == nil {
		p.ctx = context.Background()
	}
	// Initialize context and teardown channel.
	p.ctx, p.cancel = context.WithCancel(p.ctx)

	// if no output was set, set it to stdout
	if p.output == nil {
		p.output = newSafeWriter(os.Stdout)
	}

	// if no environment was set, set it to os.Environ()
	if p.environ == nil {
		p.environ = os.Environ()
	}

	if p.fps < 1 {
		p.fps = defaultFPS
	} else if p.fps > maxFPS {
		p.fps = maxFPS
	}

	// Detect if tracing is enabled.
	if tracePath := os.Getenv("TEA_TRACE"); tracePath != "" {
		switch tracePath {
		case "0", "false", "off":
			break
		}

		if _, err := LogToFile(tracePath, "bubbletea"); err == nil {
			// Enable different types of tracing.
			if output, _ := strconv.ParseBool(os.Getenv("TEA_TRACE_OUTPUT")); output {
				p.output.trace = true
			}
			if input, _ := strconv.ParseBool(os.Getenv("TEA_TRACE_INPUT")); input {
				p.traceInput = true
			}
		}
	}

	// Experimental features. Right now, we only have one experimental feature
	// to use the new cell buffer as a default renderer.
	if exp := p.getenv("TEA_EXPERIMENTAL"); exp != "" {
		p.exp = strings.Split(exp, ",")
	}

	return p
}

func (p *Program) handleSignals() chan struct{} {
	ch := make(chan struct{})

	// Listen for SIGINT and SIGTERM.
	//
	// In most cases ^C will not send an interrupt because the terminal will be
	// in raw mode and ^C will be captured as a keystroke and sent along to
	// Program.Update as a KeyMsg. When input is not a TTY, however, ^C will be
	// caught here.
	//
	// SIGTERM is sent by unix utilities (like kill) to terminate a process.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		defer func() {
			signal.Stop(sig)
			close(ch)
		}()

		for {
			select {
			case <-p.ctx.Done():
				return

			case <-sig:
				if atomic.LoadUint32(&p.ignoreSignals) == 0 {
					p.msgs <- QuitMsg{}
					return
				}
			}
		}
	}()

	return ch
}

// handleResize handles terminal resize events.
func (p *Program) handleResize() chan struct{} {
	ch := make(chan struct{})

	if p.ttyOutput != nil {
		// Listen for window resizes.
		go p.listenForResize(ch)
	} else {
		close(ch)
	}

	return ch
}

// handleCommands runs commands in a goroutine and sends the result to the
// program's message channel.
func (p *Program) handleCommands(cmds chan Cmd) chan struct{} {
	ch := make(chan struct{})

	go func() {
		defer close(ch)

		for {
			select {
			case <-p.ctx.Done():
				return

			case cmd := <-cmds:
				if cmd == nil {
					continue
				}

				// Don't wait on these goroutines, otherwise the shutdown
				// latency would get too large as a Cmd can run for some time
				// (e.g. tick commands that sleep for half a second). It's not
				// possible to cancel them so we'll have to leak the goroutine
				// until Cmd returns.
				go func() {
					// Recover from panics.
					if !p.startupOptions.has(withoutCatchPanics) {
						defer p.recoverFromPanic()
					}

					msg := cmd() // this can be long.
					p.Send(msg)
				}()
			}
		}
	}()

	return ch
}

// eventLoop is the central message loop. It receives and handles the default
// Bubble Tea messages, update the model and triggers redraws.
func (p *Program) eventLoop(model Model, cmds chan Cmd) (Model, error) {
	for {
		select {
		case <-p.ctx.Done():
			return model, nil

		case err := <-p.errs:
			return model, err

		case msg := <-p.msgs:
			// Filter messages.
			if p.filter != nil {
				msg = p.filter(model, msg)
			}
			if msg == nil {
				continue
			}

			// Handle special internal messages.
			switch msg := msg.(type) {
			case QuitMsg:
				return model, nil

			case SuspendMsg:
				if suspendSupported {
					p.suspend()
				}

			case CapabilityMsg:
				switch msg {
				case "RGB", "Tc":
					if p.profile != colorprofile.TrueColor {
						p.profile = colorprofile.TrueColor
						go p.Send(ColorProfileMsg{p.profile})
					}
				}

			case setCursorStyle:
				p.execute(ansi.SetCursorStyle(int(msg)))

			case modeReportMsg:
				switch msg.Mode {
				case int(ansi.GraphemeClusteringMode):
					// 1 means mode is set (see DECRPM).
					p.modes[ansi.GraphemeClusteringMode.String()] = msg.Value == 1 || msg.Value == 3
				}

			case enableModeMsg:
				if on, ok := p.modes[string(msg)]; ok && on {
					break
				}

				p.execute(fmt.Sprintf("\x1b[%sh", string(msg)))
				p.modes[string(msg)] = true
				switch string(msg) {
				case ansi.GraphemeClusteringMode.String():
					// We store the state of grapheme clustering after we enable it
					// and get a response in the eventLoop.
					p.execute(ansi.RequestGraphemeClustering)
				}

			case disableModeMsg:
				if on, ok := p.modes[string(msg)]; ok && !on {
					break
				}

				p.execute(fmt.Sprintf("\x1b[%sl", string(msg)))
				p.modes[string(msg)] = false

			case readClipboardMsg:
				p.execute(ansi.RequestSystemClipboard)

			case setClipboardMsg:
				p.execute(ansi.SetSystemClipboard(string(msg)))

			case readPrimaryClipboardMsg:
				p.execute(ansi.RequestPrimaryClipboard)

			case setPrimaryClipboardMsg:
				p.execute(ansi.SetPrimaryClipboard(string(msg)))

			case setBackgroundColorMsg:
				if msg.Color != nil {
					p.execute(ansi.SetBackgroundColor(msg.Color))
					p.setBg = msg.Color
				}

			case setForegroundColorMsg:
				if msg.Color != nil {
					p.execute(ansi.SetForegroundColor(msg.Color))
					p.setFg = msg.Color
				}

			case setCursorColorMsg:
				if msg.Color != nil {
					p.execute(ansi.SetCursorColor(msg.Color))
					p.setCc = msg.Color
				}

			case backgroundColorMsg:
				p.execute(ansi.RequestBackgroundColor)

			case foregroundColorMsg:
				p.execute(ansi.RequestForegroundColor)

			case cursorColorMsg:
				p.execute(ansi.RequestCursorColor)

			case KeyboardEnhancementsMsg:
				if p.keyboard.kittyFlags != msg.kittyFlags {
					p.keyboard.kittyFlags |= msg.kittyFlags
				}
				if p.keyboard.modifyOtherKeys == 0 || msg.modifyOtherKeys > p.keyboard.modifyOtherKeys {
					p.keyboard.modifyOtherKeys = msg.modifyOtherKeys
				}

			case enableKeyboardEnhancementsMsg:
				if runtime.GOOS == "windows" {
					// We use the Windows Console API which supports keyboard
					// enhancements.
					break
				}

				var ke keyboardEnhancements
				for _, e := range msg {
					e(&ke)
				}

				p.keyboard.kittyFlags |= ke.kittyFlags
				if ke.modifyOtherKeys > p.keyboard.modifyOtherKeys {
					p.keyboard.modifyOtherKeys = ke.modifyOtherKeys
				}

				if p.keyboard.modifyOtherKeys > 0 {
					p.execute(ansi.ModifyOtherKeys(p.keyboard.modifyOtherKeys))
				}
				if p.keyboard.kittyFlags > 0 {
					p.execute(ansi.PushKittyKeyboard(p.keyboard.kittyFlags))
				}

			case disableKeyboardEnhancementsMsg:
				if runtime.GOOS == "windows" {
					// We use the Windows Console API which supports keyboard
					// enhancements.
					break
				}

				if p.keyboard.modifyOtherKeys > 0 {
					p.execute(ansi.DisableModifyOtherKeys)
					p.keyboard.modifyOtherKeys = 0
				}
				if p.keyboard.kittyFlags > 0 {
					p.execute(ansi.DisableKittyKeyboard)
					p.keyboard.kittyFlags = 0
				}

			case execMsg:
				// NB: this blocks.
				p.exec(msg.cmd, msg.fn)

			case terminalVersion:
				p.execute(ansi.RequestXTVersion)

			case requestCapabilityMsg:
				p.execute(ansi.RequestTermcap(string(msg)))

			case BatchMsg:
				for _, cmd := range msg {
					cmds <- cmd
				}
				continue

			case sequenceMsg:
				go func() {
					// Execute commands one at a time, in order.
					for _, cmd := range msg {
						if cmd == nil {
							continue
						}

						switch msg := cmd().(type) {
						case BatchMsg:
							g, _ := errgroup.WithContext(p.ctx)
							for _, cmd := range msg {
								cmd := cmd
								g.Go(func() error {
									p.Send(cmd())
									return nil
								})
							}

							//nolint:errcheck
							g.Wait() // wait for all commands from batch msg to finish
							continue
						case sequenceMsg:
							for _, cmd := range msg {
								p.Send(cmd())
							}
						default:
							p.Send(msg)
						}
					}
				}()

			case setWindowTitleMsg:
				p.execute(ansi.SetWindowTitle(string(msg)))

			case windowSizeMsg:
				go p.checkResize()

			case requestCursorPosMsg:
				p.execute(ansi.RequestCursorPosition)
			}

			// Process internal messages for the renderer.
			p.renderer.update(msg)

			var cmd Cmd
			model, cmd = model.Update(msg)  // run update
			cmds <- cmd                     // process command (if any)
			p.renderer.render(model.View()) //nolint:errcheck // send view to renderer
		}
	}
}

// Run initializes the program and runs its event loops, blocking until it gets
// terminated by either [Program.Quit], [Program.Kill], or its signal handler.
// Returns the final model.
func (p *Program) Run() (Model, error) {
	p.handlers = channelHandlers{}
	cmds := make(chan Cmd)
	p.errs = make(chan error)
	p.finished = make(chan struct{}, 1)

	defer p.cancel()

	switch p.inputType {
	case defaultInput:
		p.input = os.Stdin

		// The user has not set a custom input, so we need to check whether or
		// not standard input is a terminal. If it's not, we open a new TTY for
		// input. This will allow things to "just work" in cases where data was
		// piped in or redirected to the application.
		//
		// To disable input entirely pass nil to the [WithInput] program option.
		f, isFile := p.input.(term.File)
		if !isFile {
			break
		}
		if term.IsTerminal(f.Fd()) {
			break
		}

		f, err := openInputTTY()
		if err != nil {
			return p.initialModel, err
		}
		defer f.Close() //nolint:errcheck
		p.input = f

	case ttyInput:
		// Open a new TTY, by request
		f, err := openInputTTY()
		if err != nil {
			return p.initialModel, err
		}
		defer f.Close() //nolint:errcheck
		p.input = f

	case customInput:
		// (There is nothing extra to do.)
	}

	// Handle signals.
	if !p.startupOptions.has(withoutSignalHandler) {
		p.handlers.add(p.handleSignals())
	}

	// Recover from panics.
	if !p.startupOptions.has(withoutCatchPanics) {
		defer p.recoverFromPanic()
	}

	// Check if output is a TTY before entering raw mode, hiding the cursor and
	// so on.
	if err := p.initTerminal(); err != nil {
		return p.initialModel, err
	}

	// Get the color profile and send it to the program.
	if !p.startupOptions.has(withColorProfile) {
		p.profile = colorprofile.Detect(p.output.Writer(), p.environ)
	}

	go p.Send(ColorProfileMsg{p.profile})
	if p.renderer == nil {
		// If no renderer is set use the ferocious one.
		if p.startupOptions&withFerociousRenderer != 0 || p.exp.has(experimentalFerocious) {
			p.renderer = newFerociousRenderer(p.profile)
		} else {
			p.renderer = newStandardRenderer(p.profile)
		}
	}

	// Set the renderer output.
	p.renderer.update(rendererWriter{p.output})
	if p.ttyOutput != nil {
		// Set the initial size of the terminal.
		w, h, err := term.GetSize(p.ttyOutput.Fd())
		if err != nil {
			return p.initialModel, err
		}

		// Send the initial size to the program.
		go p.Send(WindowSizeMsg{
			Width:  w,
			Height: h,
		})
	}

	// Init the input reader and initial model.
	model := p.initialModel
	if p.input != nil {
		if err := p.initInputReader(); err != nil {
			return model, err
		}
	}

	// Hide the cursor before starting the renderer.
	p.modes[ansi.CursorEnableMode.String()] = false
	p.execute(ansi.HideCursor)
	p.renderer.update(disableMode(ansi.CursorEnableMode.String()))

	// Honor program startup options.
	if p.startupTitle != "" {
		p.execute(ansi.SetWindowTitle(p.startupTitle))
	}
	if p.startupOptions&withAltScreen != 0 {
		p.execute(ansi.EnableAltScreenBuffer)
		p.modes[ansi.AltScreenBufferMode.String()] = true
		p.renderer.update(enableMode(ansi.AltScreenBufferMode.String()))
	}
	if p.startupOptions&withoutBracketedPaste == 0 {
		p.execute(ansi.EnableBracketedPaste)
		p.modes[ansi.BracketedPasteMode.String()] = true
	}
	if p.startupOptions&withGraphemeClustering != 0 {
		p.execute(ansi.EnableGraphemeClustering)
		p.execute(ansi.RequestGraphemeClustering)
		// We store the state of grapheme clustering after we query it and get
		// a response in the eventLoop.
	}
	if p.startupOptions&withMouseCellMotion != 0 {
		p.execute(ansi.EnableMouseCellMotion)
		p.execute(ansi.EnableMouseSgrExt)
		p.modes[ansi.MouseCellMotionMode.String()] = true
		p.modes[ansi.MouseSgrExtMode.String()] = true
	} else if p.startupOptions&withMouseAllMotion != 0 {
		p.execute(ansi.EnableMouseAllMotion)
		p.execute(ansi.EnableMouseSgrExt)
		p.modes[ansi.MouseAllMotionMode.String()] = true
		p.modes[ansi.MouseSgrExtMode.String()] = true
	}

	if p.startupOptions&withReportFocus != 0 {
		p.execute(ansi.EnableReportFocus)
		p.modes[ansi.ReportFocusMode.String()] = true
	}
	if p.startupOptions&withKeyboardEnhancements != 0 && runtime.GOOS != "windows" {
		// We use the Windows Console API which supports keyboard
		// enhancements.

		if p.keyboard.modifyOtherKeys > 0 {
			p.execute(ansi.ModifyOtherKeys(p.keyboard.modifyOtherKeys))
			p.execute(ansi.RequestModifyOtherKeys)
		}
		if p.keyboard.kittyFlags > 0 {
			p.execute(ansi.PushKittyKeyboard(p.keyboard.kittyFlags))
			p.execute(ansi.RequestKittyKeyboard)
		}
	}

	// Start the renderer.
	p.startRenderer()

	// Initialize the program.
	var initCmd Cmd
	model, initCmd = model.Init()
	if initCmd != nil {
		ch := make(chan struct{})
		p.handlers.add(ch)

		go func() {
			defer close(ch)

			select {
			case cmds <- initCmd:
			case <-p.ctx.Done():
			}
		}()
	}

	// Render the initial view.
	p.renderer.render(model.View()) //nolint:errcheck

	// Handle resize events.
	p.handlers.add(p.handleResize())

	// Process commands.
	p.handlers.add(p.handleCommands(cmds))

	// Run event loop, handle updates and draw.
	model, err := p.eventLoop(model, cmds)
	killed := p.ctx.Err() != nil
	if killed {
		err = fmt.Errorf("%w: %s", ErrProgramKilled, p.ctx.Err())
	} else {
		// Ensure we rendered the final state of the model.
		p.renderer.render(model.View()) //nolint:errcheck
	}

	// Restore terminal state.
	p.shutdown(killed)

	return model, err
}

// Send sends a message to the main update function, effectively allowing
// messages to be injected from outside the program for interoperability
// purposes.
//
// If the program hasn't started yet this will be a blocking operation.
// If the program has already been terminated this will be a no-op, so it's safe
// to send messages after the program has exited.
func (p *Program) Send(msg Msg) {
	select {
	case <-p.ctx.Done():
	case p.msgs <- msg:
	}
}

// Quit is a convenience function for quitting Bubble Tea programs. Use it
// when you need to shut down a Bubble Tea program from the outside.
//
// If you wish to quit from within a Bubble Tea program use the Quit command.
//
// If the program is not running this will be a no-op, so it's safe to call
// if the program is unstarted or has already exited.
func (p *Program) Quit() {
	p.Send(Quit())
}

// Kill stops the program immediately and restores the former terminal state.
// The final render that you would normally see when quitting will be skipped.
// [program.Run] returns a [ErrProgramKilled] error.
func (p *Program) Kill() {
	p.shutdown(true)
}

// Wait waits/blocks until the underlying Program finished shutting down.
func (p *Program) Wait() {
	<-p.finished
}

// execute writes the given sequence to the program output.
func (p *Program) execute(seq string) {
	io.WriteString(p.output, seq) //nolint:errcheck
}

// shutdown performs operations to free up resources and restore the terminal
// to its original state.
func (p *Program) shutdown(kill bool) {
	p.shutdownOnce.Do(func() {
		p.cancel()

		// Wait for all handlers to finish.
		p.handlers.shutdown()

		// Check if the cancel reader has been setup before waiting and closing.
		if p.inputReader != nil {
			// Wait for input loop to finish.
			if p.inputReader.Cancel() {
				if !kill {
					p.waitForReadLoop()
				}
			}
			_ = p.inputReader.Close()
		}

		if p.renderer != nil {
			p.stopRenderer(kill)
		}

		_ = p.restoreTerminalState()
		if !kill {
			p.finished <- struct{}{}
		}
	})
}

// recoverFromPanic recovers from a panic, prints the stack trace, and restores
// the terminal to a usable state.
func (p *Program) recoverFromPanic() {
	if r := recover(); r != nil {
		p.shutdown(true)
		fmt.Printf("Caught panic:\n\n%s\n\nRestoring terminal...\n\n", r)
		debug.PrintStack()
	}
}

// ReleaseTerminal restores the original terminal state and cancels the input
// reader. You can return control to the Program with RestoreTerminal.
func (p *Program) ReleaseTerminal() error {
	atomic.StoreUint32(&p.ignoreSignals, 1)
	if p.inputReader != nil {
		p.inputReader.Cancel()
	}

	p.waitForReadLoop()

	if p.renderer != nil {
		p.stopRenderer(false)
	}

	return p.restoreTerminalState()
}

// RestoreTerminal reinitializes the Program's input reader, restores the
// terminal to the former state when the program was running, and repaints.
// Use it to reinitialize a Program after running ReleaseTerminal.
func (p *Program) RestoreTerminal() error {
	atomic.StoreUint32(&p.ignoreSignals, 0)

	if err := p.initTerminal(); err != nil {
		return err
	}
	if err := p.initInputReader(); err != nil {
		return err
	}
	if p.modes[ansi.AltScreenBufferMode.String()] {
		p.execute(ansi.EnableAltScreenBuffer)
	} else {
		// entering alt screen already causes a repaint.
		go p.Send(repaintMsg{})
	}

	p.startRenderer()
	if !p.modes[ansi.CursorEnableMode.String()] {
		p.execute(ansi.HideCursor)
	} else {
		p.execute(ansi.ShowCursor)
	}
	if p.modes[ansi.BracketedPasteMode.String()] {
		p.execute(ansi.EnableBracketedPaste)
	}
	if p.keyboard.modifyOtherKeys != 0 {
		p.execute(ansi.ModifyOtherKeys(p.keyboard.modifyOtherKeys))
	}
	if p.keyboard.kittyFlags != 0 {
		p.execute(ansi.PushKittyKeyboard(p.keyboard.kittyFlags))
	}
	if p.modes[ansi.ReportFocusMode.String()] {
		p.execute(ansi.EnableReportFocus)
	}
	if p.modes[ansi.MouseCellMotionMode.String()] || p.modes[ansi.MouseAllMotionMode.String()] {
		if p.startupOptions&withMouseCellMotion != 0 {
			p.execute(ansi.EnableMouseCellMotion)
			p.execute(ansi.EnableMouseSgrExt)
		} else if p.startupOptions&withMouseAllMotion != 0 {
			p.execute(ansi.EnableMouseAllMotion)
			p.execute(ansi.EnableMouseSgrExt)
		}
	}
	if p.modes[ansi.GraphemeClusteringMode.String()] {
		p.execute(ansi.EnableGraphemeClustering)
	}

	// Restore terminal colors.
	if p.setBg != nil {
		p.execute(ansi.SetBackgroundColor(p.setBg))
	}
	if p.setFg != nil {
		p.execute(ansi.SetForegroundColor(p.setFg))
	}
	if p.setCc != nil {
		p.execute(ansi.SetCursorColor(p.setCc))
	}

	// If the output is a terminal, it may have been resized while another
	// process was at the foreground, in which case we may not have received
	// SIGWINCH. Detect any size change now and propagate the new size as
	// needed.
	go p.checkResize()

	return nil
}

// Println prints above the Program. This output is unmanaged by the program
// and will persist across renders by the Program.
//
// If the altscreen is active no output will be printed.
func (p *Program) Println(args ...interface{}) {
	p.msgs <- printLineMessage{
		messageBody: fmt.Sprint(args...),
	}
}

// Printf prints above the Program. It takes a format template followed by
// values similar to fmt.Printf. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Printf (but similar to log.Printf) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func (p *Program) Printf(template string, args ...interface{}) {
	p.msgs <- printLineMessage{
		messageBody: fmt.Sprintf(template, args...),
	}
}

// startRenderer starts the renderer.
func (p *Program) startRenderer() {
	framerate := time.Second / time.Duration(p.fps)
	if p.ticker == nil {
		p.ticker = time.NewTicker(framerate)
	} else {
		// If the ticker already exists, it has been stopped and we need to
		// reset it.
		p.ticker.Reset(framerate)
	}

	// Since the renderer can be restarted after a stop, we need to reset
	// the done channel and its corresponding sync.Once.
	p.once = sync.Once{}

	// Start the renderer.
	if p.renderer != nil {
		p.renderer.reset()
	}
	go func() {
		for {
			select {
			case <-p.rendererDone:
				p.ticker.Stop()
				return

			case <-p.ticker.C:
				p.renderer.flush() //nolint:errcheck
			}
		}
	}()
}

// stopRenderer stops the renderer.
// If kill is true, the renderer will be stopped immediately without flushing
// the last frame.
func (p *Program) stopRenderer(kill bool) {
	// Stop the renderer before acquiring the mutex to avoid a deadlock.
	p.once.Do(func() {
		p.rendererDone <- struct{}{}
	})

	if !kill {
		// flush locks the mutex
		p.renderer.flush() //nolint:errcheck
	}

	p.renderer.close() //nolint:errcheck
}
