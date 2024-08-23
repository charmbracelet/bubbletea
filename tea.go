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
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/ansi/compressor"
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
	Init() Cmd

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
	withANSICompressor
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
)

// channelHandlers manages the series of channels returned by various processes.
// It allows us to wait for those processes to terminate before exiting the
// program.
type channelHandlers []chan struct{}

// Adds a channel to the list of handlers. We wait for all handlers to terminate
// gracefully on shutdown.
func (h *channelHandlers) add(ch chan struct{}) {
	*h = append(*h, ch)
}

// shutdown waits for all handlers to terminate.
func (h channelHandlers) shutdown() {
	var wg sync.WaitGroup
	for _, ch := range h {
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

	// Configuration options that will set as the program is initializing,
	// treated as bits. These options can be set via various ProgramOptions.
	startupOptions startupOptions

	// startupTitle is the title that will be set on the terminal when the
	// program starts.
	startupTitle string

	inputType inputType

	ctx    context.Context
	cancel context.CancelFunc

	msgs     chan Msg
	errs     chan error
	finished chan struct{}

	// where to send output, this will usually be os.Stdout.
	output io.Writer
	// ttyOutput is null if output is not a TTY.
	ttyOutput           term.File
	previousOutputState *term.State
	renderer            Renderer

	// the environment variables for the program, defaults to os.Environ().
	environ []string

	// where to read inputs from, this will usually be os.Stdin.
	input io.Reader
	// ttyInput is null if input is not a TTY.
	ttyInput              term.File
	previousTtyInputState *term.State
	inputReader           *driver
	readLoopDone          chan struct{}

	// was the altscreen active before releasing the terminal?
	altScreenWasActive bool
	ignoreSignals      uint32

	bpActive bool // was the bracketed paste mode active before releasing the terminal?

	graphemeClustering bool // whether grapheme clustering is enabled

	cursorHidden bool // the cursor visibility state

	mouseEnabled bool // whether mouse reporting is enabled

	reportFocus bool // whether focus reporting is enabled

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

	// kittyFlags stores kitty keyboard protocol progressive enhancement flags.
	kittyFlags int

	// modifyOtherKeys stores the XTerm modifyOtherKeys mode.
	modifyOtherKeys int

	// win32Input indicates whether the program has win32-input-mode enabled.
	win32Input bool
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
		p.output = os.Stdout
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
					msg := cmd() // this can be long.
					p.Send(msg)
				}()
			}
		}
	}()

	return ch
}

func (p *Program) disableMouse() {
	p.execute(ansi.DisableMouseCellMotion)
	p.execute(ansi.DisableMouseAllMotion)
	p.execute(ansi.DisableMouseSgrExt)
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

			case ReportModeMsg:
				switch msg.Mode {
				case graphemeClustering:
					// 1 means mode is set (see DECRPM).
					p.graphemeClustering = msg.Value == 1
					if p.graphemeClustering {
						p.renderer.SetMode(graphemeClustering, true)
					}
				}

			case clearScreenMsg:
				p.renderer.ClearScreen()

			case enterAltScreenMsg:
				p.renderer.SetMode(altScreenMode, true)

			case exitAltScreenMsg:
				p.renderer.SetMode(altScreenMode, false)

			case enableMouseCellMotionMsg, enableMouseAllMotionMsg:
				switch msg.(type) {
				case enableMouseCellMotionMsg:
					p.execute(ansi.EnableMouseCellMotion)
				case enableMouseAllMotionMsg:
					p.execute(ansi.EnableMouseAllMotion)
				}
				// mouse mode (1006) is a no-op if the terminal doesn't support it.
				p.execute(ansi.EnableMouseSgrExt)
				p.mouseEnabled = true

			case disableMouseMsg:
				p.disableMouse()
				p.mouseEnabled = false

			case showCursorMsg:
				p.renderer.SetMode(hideCursor, false)

			case hideCursorMsg:
				p.renderer.SetMode(hideCursor, true)

			case enableBracketedPasteMsg:
				p.execute(ansi.EnableBracketedPaste)
				p.bpActive = true

			case disableBracketedPasteMsg:
				p.execute(ansi.DisableBracketedPaste)
				p.bpActive = false

			case enableGraphemeClusteringMsg:
				p.execute(ansi.EnableGraphemeClustering)
				p.execute(ansi.RequestGraphemeClustering)
				// We store the state of grapheme clustering after we enable it
				// and get a response in the eventLoop.

			case disableGraphemeClusteringMsg:
				if p.graphemeClustering {
					// We only disable grapheme clustering if it was enabled.
					p.execute(ansi.DisableGraphemeClustering)
					p.renderer.SetMode(graphemeClustering, false)
				}

			case enableReportFocusMsg:
				p.execute(ansi.EnableReportFocus)
				p.reportFocus = true

			case disableReportFocusMsg:
				p.execute(ansi.DisableReportFocus)
				p.reportFocus = false

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
				}

			case setForegroundColorMsg:
				if msg.Color != nil {
					p.execute(ansi.SetForegroundColor(msg.Color))
				}

			case setCursorColorMsg:
				if msg.Color != nil {
					p.execute(ansi.SetCursorColor(msg.Color))
				}

			case backgroundColorMsg:
				p.execute(ansi.RequestBackgroundColor)

			case foregroundColorMsg:
				p.execute(ansi.RequestForegroundColor)

			case cursorColorMsg:
				p.execute(ansi.RequestCursorColor)

			case _KittyKeyboardMsg:
				// Store the kitty flags whenever they are queried.
				p.kittyFlags = int(msg)

			case setKittyKeyboardFlagsMsg:
				p.kittyFlags = int(msg)
				p.execute(ansi.PushKittyKeyboard(p.kittyFlags))

			case kittyKeyboardMsg:
				p.execute(ansi.RequestKittyKeyboard)

			case modifyOtherKeys:
				p.execute(ansi.RequestModifyOtherKeys)

			case setModifyOtherKeysMsg:
				p.modifyOtherKeys = int(msg)
				p.execute(ansi.ModifyOtherKeys(p.modifyOtherKeys))

			case setEnhancedKeyboardMsg:
				if bool(msg) {
					p.kittyFlags = 3
					p.modifyOtherKeys = 1
				} else {
					p.kittyFlags = 0
					p.modifyOtherKeys = 0
				}
				p.execute(ansi.ModifyOtherKeys(p.modifyOtherKeys))
				p.execute(ansi.PushKittyKeyboard(p.kittyFlags))

			case enableWin32InputMsg:
				p.execute(ansi.EnableWin32Input)
				p.win32Input = true

			case disableWin32InputMsg:
				p.execute(ansi.DisableWin32Input)
				p.win32Input = false

			case execMsg:
				// NB: this blocks.
				p.exec(msg.cmd, msg.fn)

			case terminalVersion:
				p.execute(ansi.RequestXTVersion)

			case primaryDeviceAttrsMsg:
				p.execute(ansi.RequestPrimaryDeviceAttributes)

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

						msg := cmd()
						if batchMsg, ok := msg.(BatchMsg); ok {
							g, _ := errgroup.WithContext(p.ctx)
							for _, cmd := range batchMsg {
								cmd := cmd
								g.Go(func() error {
									p.Send(cmd())
									return nil
								})
							}

							//nolint:errcheck
							g.Wait() // wait for all commands from batch msg to finish
							continue
						}

						p.Send(msg)
					}
				}()

			case setWindowTitleMsg:
				p.SetWindowTitle(string(msg))

			case windowSizeMsg:
				go p.checkResize()

			case WindowSizeMsg:
				p.renderer.Resize(msg.Width, msg.Height)

			case printLineMessage:
				p.renderer.InsertAbove(msg.messageBody) //nolint:errcheck
			}

			// Process internal messages for the renderer.
			if r, ok := p.renderer.(*standardRenderer); ok {
				r.handleMessages(msg)
			}

			var cmd Cmd
			model, cmd = model.Update(msg)  // run update
			cmds <- cmd                     // process command (if any)
			p.renderer.Render(model.View()) //nolint:errcheck // send view to renderer
		}
	}
}

// Run initializes the program and runs its event loops, blocking until it gets
// terminated by either [Program.Quit], [Program.Kill], or its signal handler.
// Returns the final model.
func (p *Program) Run() (Model, error) {
	handlers := channelHandlers{}
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
		handlers.add(p.handleSignals())
	}

	// Recover from panics.
	if !p.startupOptions.has(withoutCatchPanics) {
		defer func() {
			if r := recover(); r != nil {
				p.shutdown(true)
				fmt.Printf("Caught panic:\n\n%s\n\nRestoring terminal...\n\n", r)
				debug.PrintStack()
				return
			}
		}()
	}

	// Check if output is a TTY before entering raw mode, hiding the cursor and
	// so on.
	if err := p.initTerminal(); err != nil {
		return p.initialModel, err
	}

	// If no renderer is set use the standard one.
	output := p.output
	if p.renderer == nil {
		// TODO(v2): remove the ANSI compressor
		if p.startupOptions.has(withANSICompressor) {
			output = &compressor.Writer{Forward: output}
		}
		p.renderer = NewStandardRenderer()
	}

	// Set the renderer output.
	p.renderer.SetOutput(output)
	if p.ttyOutput != nil {
		// Set the initial size of the terminal.
		w, h, err := term.GetSize(p.ttyOutput.Fd())
		if err != nil {
			return p.initialModel, err
		}

		p.renderer.Resize(w, h)

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
	p.renderer.SetMode(hideCursor, true)

	// Honor program startup options.
	if p.startupTitle != "" {
		p.execute(ansi.SetWindowTitle(p.startupTitle))
	}
	if p.startupOptions&withAltScreen != 0 {
		p.renderer.SetMode(altScreenMode, true)
	}
	if p.startupOptions&withoutBracketedPaste == 0 {
		p.execute(ansi.EnableBracketedPaste)
		p.bpActive = true
	}
	if p.startupOptions&withoutGraphemeClustering == 0 {
		p.execute(ansi.EnableGraphemeClustering)
		p.execute(ansi.RequestGraphemeClustering)
		// We store the state of grapheme clustering after we query it and get
		// a response in the eventLoop.
	}
	if p.startupOptions&withMouseCellMotion != 0 {
		p.execute(ansi.EnableMouseCellMotion)
		p.execute(ansi.EnableMouseSgrExt)
		p.mouseEnabled = true
	} else if p.startupOptions&withMouseAllMotion != 0 {
		p.execute(ansi.EnableMouseAllMotion)
		p.execute(ansi.EnableMouseSgrExt)
		p.mouseEnabled = true
	}
	if p.startupOptions&withModifyOtherKeys != 0 {
		p.execute(ansi.ModifyOtherKeys(p.modifyOtherKeys))
	}
	if p.startupOptions&withKittyKeyboard != 0 {
		p.execute(ansi.PushKittyKeyboard(p.kittyFlags))
	}

	if p.startupOptions&withReportFocus != 0 {
		p.execute(ansi.EnableReportFocus)
		p.reportFocus = true
	}
	if p.startupOptions&withWindowsInputMode != 0 {
		p.execute(ansi.EnableWin32Input)
		p.win32Input = true
	}

	// Start the renderer.
	p.startRenderer()

	// Initialize the program.
	if initCmd := model.Init(); initCmd != nil {
		ch := make(chan struct{})
		handlers.add(ch)

		go func() {
			defer close(ch)

			select {
			case cmds <- initCmd:
			case <-p.ctx.Done():
			}
		}()
	}

	// Render the initial view.
	p.renderer.Render(model.View()) //nolint:errcheck

	// Handle resize events.
	handlers.add(p.handleResize())

	// Process commands.
	handlers.add(p.handleCommands(cmds))

	// Run event loop, handle updates and draw.
	model, err := p.eventLoop(model, cmds)
	killed := p.ctx.Err() != nil
	if killed {
		err = fmt.Errorf("%w: %s", ErrProgramKilled, p.ctx.Err())
	} else {
		// Ensure we rendered the final state of the model.
		p.renderer.Render(model.View()) //nolint:errcheck
	}

	// Tear down.
	p.cancel()

	// Check if the cancel reader has been setup before waiting and closing.
	if p.inputReader != nil {
		// Wait for input loop to finish.
		if p.inputReader.Cancel() {
			p.waitForReadLoop()
		}
		_ = p.inputReader.Close()
	}

	// Wait for all handlers to finish.
	handlers.shutdown()

	// Restore terminal state.
	p.shutdown(killed)

	return model, err
}

// StartReturningModel initializes the program and runs its event loops,
// blocking until it gets terminated by either [Program.Quit], [Program.Kill],
// or its signal handler. Returns the final model.
//
// Deprecated: please use [Program.Run] instead.
func (p *Program) StartReturningModel() (Model, error) {
	return p.Run()
}

// Start initializes the program and runs its event loops, blocking until it
// gets terminated by either [Program.Quit], [Program.Kill], or its signal
// handler.
//
// Deprecated: please use [Program.Run] instead.
func (p *Program) Start() error {
	_, err := p.Run()
	return err
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
	p.cancel()
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
	if p.renderer != nil {
		p.stopRenderer(kill)
	}

	_ = p.restoreTerminalState()
	p.finished <- struct{}{}
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
		// TODO: store these values when they're set in the eventLoop and [Run].
		p.altScreenWasActive = p.renderer.Mode(altScreenMode)
		p.cursorHidden = p.renderer.Mode(hideCursor)
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
	if p.altScreenWasActive {
		p.renderer.SetMode(altScreenMode, true)
	} else {
		// entering alt screen already causes a repaint.
		go p.Send(repaintMsg{})
	}
	if p.renderer != nil {
		p.startRenderer()
		if p.cursorHidden {
			p.renderer.SetMode(hideCursor, true)
		} else {
			p.renderer.SetMode(hideCursor, false)
		}
	}
	if p.cursorHidden {
		p.execute(ansi.HideCursor)
	}
	if p.bpActive {
		p.execute(ansi.EnableBracketedPaste)
	}
	if p.modifyOtherKeys != 0 {
		p.execute(ansi.ModifyOtherKeys(p.modifyOtherKeys))
	}
	if p.kittyFlags != 0 {
		p.execute(ansi.PushKittyKeyboard(p.kittyFlags))
	}
	if p.reportFocus {
		p.execute(ansi.EnableReportFocus)
	}
	if p.mouseEnabled {
		if p.startupOptions&withMouseCellMotion != 0 {
			p.execute(ansi.EnableMouseCellMotion)
			p.execute(ansi.EnableMouseSgrExt)
		} else if p.startupOptions&withMouseAllMotion != 0 {
			p.execute(ansi.EnableMouseAllMotion)
			p.execute(ansi.EnableMouseSgrExt)
		}
	}
	if p.graphemeClustering {
		p.execute(ansi.EnableGraphemeClustering)
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
	go func() {
		for {
			select {
			case <-p.rendererDone:
				p.ticker.Stop()
				return

			case <-p.ticker.C:
				p.renderer.Flush() //nolint:errcheck
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
		p.renderer.Flush() //nolint:errcheck
	}

	p.renderer.Close() //nolint:errcheck

	if !kill && p.startupOptions.has(withANSICompressor) {
		if w, ok := p.output.(io.WriteCloser); ok {
			_ = w.Close()
		}
	}
}
