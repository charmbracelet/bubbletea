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
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/input"
	"github.com/charmbracelet/x/term"
	"golang.org/x/sync/errgroup"
)

// ErrProgramKilled is returned by [Program.Run] when the program gets killed.
var ErrProgramKilled = errors.New("program was killed")

// ErrInterrupted is returned by [Program.Run] when the program get a SIGINT
// signal, or when it receives a [InterruptMsg].
var ErrInterrupted = errors.New("program was interrupted")

// Msg contain data from the result of a IO operation. Msgs trigger the update
// function and, henceforth, the UI.
type Msg interface{}

// Model contains the program's state as well as its core functions.
type Model[T any] interface {
	// Init is the first function that will be called. It returns an optional
	// initial command. To not perform an initial command return nil.
	Init() (T, Cmd)

	// Update is called when a message is received. Use it to inspect messages
	// and, in response, update the model and/or send a command.
	Update(Msg) (T, Cmd)

	// View renders the program's UI, which is just a [fmt.Stringer]. The view
	// is rendered after every Update.
	// The main model can return a [Frame] to set the cursor position and
	// style.
	View() fmt.Stringer
}

// Cmd is an IO operation that returns a message when it's complete. If it's
// nil it's considered a no-op. Use it for things like HTTP requests, timers,
// saving and loading from disk, and so on.
//
// Note that there's almost never a reason to use a command to send a message
// to another part of your program. That can almost always be done in the
// update function.
type Cmd func() Msg

// Profile returns a pointer to a color profile. Use it to set the color
// profile of the terminal.
func Profile(p colorprofile.Profile) *colorprofile.Profile {
	return &p
}

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
type Program[T any] struct {
	Input  io.Reader
	Output io.Writer
	Env    []string

	Init   func() (T, Cmd)
	Filter func(T, Msg) Msg
	Update func(T, Msg) (T, Cmd)
	View   func(T) fmt.Stringer

	Model T

	// handlers is a list of channels that need to be waited on before the
	// program can exit.
	handlers channelHandlers

	// Title is the title that will be set on the terminal when the
	// program starts.
	Title string

	Context context.Context
	cancel  context.CancelFunc

	msgs         chan Msg
	errs         chan error
	finished     chan struct{}
	shutdownOnce sync.Once

	// DontCatchPanics is a flag that determines whether or not the program should
	// catch panics.
	DontCatchPanics bool

	// IgnoreSignals is a flag that determines whether or not the program should
	// ignore signals.
	IgnoreSignals bool

	// Profile is the color profile of the terminal. Use [Profile] to set it.
	Profile *colorprofile.Profile // the terminal color profile

	// ttyOutput is null if output is not a TTY.
	ttyOutput           term.File
	previousOutputState *term.State
	renderer            renderer

	// the environment variables for the program, defaults to os.Environ().
	environ environ

	// ttyInput is null if input is not a TTY.
	ttyInput              term.File
	previousTtyInputState *term.State
	inputReader           *input.Reader
	traceInput            bool // true if input should be traced
	readLoopDone          chan struct{}

	// IsInputTTY is true if the input is a TTY. Use this to tell the program
	// that the input is a TTY and that it should be treated as such.
	IsInputTTY bool

	// modes keeps track of terminal modes that have been enabled or disabled.
	modes         ansi.Modes
	ignoreSignals uint32

	// FPS is the frames per second we should set on the renderer, if
	// applicable,
	FPS int

	// ticker is the ticker that will be used to write to the renderer.
	ticker *time.Ticker

	// once is used to stop the renderer.
	once sync.Once

	// rendererDone is used to stop the renderer.
	rendererDone chan struct{}

	// stores the requested keyboard enhancements.
	requestedEnhancements KeyboardEnhancements
	// activeEnhancements stores the active keyboard enhancements read from the
	// terminal.
	activeEnhancements KeyboardEnhancements

	// keyboardc is used to signal that the keyboard enhancements have been
	// read from the terminal.
	keyboardc chan struct{}

	// When a program is suspended, the terminal state is saved and the program
	// is paused. This saves the terminal colors state so they can be restored
	// when the program is resumed.
	setBg, setFg, setCc color.Color

	// Initial window size. Mainly used for testing.
	width, height int

	// whether to use hard tabs to optimize cursor movements
	useHardTabs bool
}

// Quit is a special command that tells the Bubble Tea program to exit.
func Quit() Msg {
	return QuitMsg{}
}

// QuitMsg signals that the program should quit. You can send a [QuitMsg] with
// [Quit].
type QuitMsg struct{}

// Suspend is a special command that tells the Bubble Tea program to suspend.
func Suspend() Msg {
	return SuspendMsg{}
}

// SuspendMsg signals the program should suspend.
// This usually happens when ctrl+z is pressed on common programs, but since
// bubbletea puts the terminal in raw mode, we need to handle it in a
// per-program basis.
//
// You can send this message with [Suspend()].
type SuspendMsg struct{}

// ResumeMsg can be listen to to do something once a program is resumed back
// from a suspend state.
type ResumeMsg struct{}

// InterruptMsg signals the program should suspend.
// This usually happens when ctrl+c is pressed on common programs, but since
// bubbletea puts the terminal in raw mode, we need to handle it in a
// per-program basis.
//
// You can send this message with [Interrupt()].
type InterruptMsg struct{}

// Interrupt is a special command that tells the Bubble Tea program to
// interrupt.
func Interrupt() Msg {
	return InterruptMsg{}
}

// NewProgram creates a new Program.
func NewProgram[T any](model Model[T]) *Program[T] {
	p := new(Program[T])
	p.Init = model.Init
	p.Update = func(t T, msg Msg) (T, Cmd) { return any(t).(Model[T]).Update(msg) }
	p.View = func(t T) fmt.Stringer { return any(t).(Model[T]).View() }
	return p
}

func (p *Program[T]) init() {
	p.msgs = make(chan Msg)
	p.rendererDone = make(chan struct{})
	p.keyboardc = make(chan struct{})
	p.modes = ansi.Modes{}

	// A context can be provided with a ProgramOption, but if none was provided
	// we'll use the default background context.
	if p.Context == nil {
		p.Context = context.Background()
	}
	// Initialize context and teardown channel.
	p.Context, p.cancel = context.WithCancel(p.Context)

	// if no output was set, set it to stdout
	if p.Output == nil {
		p.Output = os.Stdout
	}

	// if no environment was set, set it to os.Environ()
	if p.environ == nil {
		p.environ = os.Environ()
	}

	if p.FPS < 1 {
		p.FPS = defaultFPS
	} else if p.FPS > maxFPS {
		p.FPS = maxFPS
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
				// p.Output.(*safeWriter).trace = true
			}
			if input, _ := strconv.ParseBool(os.Getenv("TEA_TRACE_INPUT")); input {
				p.traceInput = true
			}
		}
	}
}

func (p *Program[T]) handleSignals() chan struct{} {
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
			case <-p.Context.Done():
				return

			case s := <-sig:
				switch s {
				case syscall.SIGINT:
					p.msgs <- InterruptMsg{}
				default:
					p.msgs <- QuitMsg{}
				}
				return
			}
		}
	}()

	return ch
}

// handleResize handles terminal resize events.
func (p *Program[T]) handleResize() chan struct{} {
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
func (p *Program[T]) handleCommands(cmds chan Cmd) chan struct{} {
	ch := make(chan struct{})

	go func() {
		defer close(ch)

		for {
			select {
			case <-p.Context.Done():
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
					if p.DontCatchPanics {
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
func (p *Program[T]) eventLoop(cmds chan Cmd) {
	for {
		select {
		case <-p.Context.Done():
			return

		case msg := <-p.msgs:
			// Filter messages.
			if p.Filter != nil {
				msg = p.Filter(p.Model, msg)
			}
			if msg == nil {
				continue
			}

			// Handle special internal messages.
			switch msg := msg.(type) {
			case QuitMsg:
				return

			case InterruptMsg:
				go func() { p.errs <- ErrInterrupted }()
				return

			case SuspendMsg:
				if suspendSupported {
					p.suspend()
				}

			case CapabilityMsg:
				switch msg {
				case "RGB", "Tc":
					if p.Profile != nil && *p.Profile != colorprofile.TrueColor {
						*p.Profile = colorprofile.TrueColor
						go p.Send(ColorProfileMsg{*p.Profile})
					}
				}

			case modeReportMsg:
				switch msg.Mode {
				case ansi.GraphemeClusteringMode:
					// 1 means mode is set (see DECRPM).
					p.modes[ansi.GraphemeClusteringMode] = msg.Value
				}

			case enableModeMsg:
				mode := p.modes.Get(msg.Mode)
				if mode.IsSet() {
					break
				}

				p.modes.Set(msg.Mode)

				switch msg.Mode {
				case ansi.AltScreenSaveCursorMode:
					p.renderer.enterAltScreen()
				case ansi.TextCursorEnableMode:
					p.renderer.showCursor()
				case ansi.GraphemeClusteringMode:
					// We store the state of grapheme clustering after we enable it
					// and get a response in the eventLoop.
					p.execute(ansi.SetGraphemeClusteringMode + ansi.RequestGraphemeClusteringMode)
				default:
					p.execute(ansi.SetMode(msg.Mode))
				}

			case disableModeMsg:
				mode := p.modes.Get(msg.Mode)
				if mode.IsReset() {
					break
				}

				p.modes.Reset(msg.Mode)

				switch msg.Mode {
				case ansi.AltScreenSaveCursorMode:
					p.renderer.exitAltScreen()
				case ansi.TextCursorEnableMode:
					p.renderer.hideCursor()
				default:
					p.execute(ansi.ResetMode(msg.Mode))
				}

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
				} else {
					p.execute(ansi.ResetBackgroundColor)
				}
				p.setBg = msg.Color

			case setForegroundColorMsg:
				if msg.Color != nil {
					p.execute(ansi.SetForegroundColor(msg.Color))
				} else {
					p.execute(ansi.ResetForegroundColor)
				}
				p.setFg = msg.Color

			case setCursorColorMsg:
				if msg.Color != nil {
					p.execute(ansi.SetCursorColor(msg.Color))
				} else {
					p.execute(ansi.ResetCursorColor)
				}
				p.setCc = msg.Color

			case backgroundColorMsg:
				p.execute(ansi.RequestBackgroundColor)

			case foregroundColorMsg:
				p.execute(ansi.RequestForegroundColor)

			case cursorColorMsg:
				p.execute(ansi.RequestCursorColor)

			case KeyboardEnhancementsMsg:
				p.activeEnhancements.kittyFlags = msg.kittyFlags
				p.activeEnhancements.modifyOtherKeys = msg.modifyOtherKeys

				go func() {
					// Signal that we've read the keyboard enhancements.
					p.keyboardc <- struct{}{}
				}()

			case enableKeyboardEnhancementsMsg:
				if runtime.GOOS == "windows" {
					// We use the Windows Console API which supports keyboard
					// enhancements.
					break
				}

				var ke KeyboardEnhancements
				for _, e := range msg {
					e(&ke)
				}

				p.requestedEnhancements.kittyFlags |= ke.kittyFlags
				if ke.modifyOtherKeys > p.requestedEnhancements.modifyOtherKeys {
					p.requestedEnhancements.modifyOtherKeys = ke.modifyOtherKeys
				}

				p.requestKeyboardEnhancements()

				// Ensure we send a message so that terminals that don't support the
				// requested features can disable them.
				go p.sendKeyboardEnhancementsMsg()

			case disableKeyboardEnhancementsMsg:
				if runtime.GOOS == "windows" {
					// We use the Windows Console API which supports keyboard
					// enhancements.
					break
				}

				if p.activeEnhancements.modifyOtherKeys > 0 {
					p.execute(ansi.DisableModifyOtherKeys)
					p.activeEnhancements.modifyOtherKeys = 0
					p.requestedEnhancements.modifyOtherKeys = 0
				}
				if p.activeEnhancements.kittyFlags > 0 {
					p.execute(ansi.DisableKittyKeyboard)
					p.activeEnhancements.kittyFlags = 0
					p.requestedEnhancements.kittyFlags = 0
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
							g, _ := errgroup.WithContext(p.Context)
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

			case WindowSizeMsg:
				p.renderer.resize(msg.Width, msg.Height)

			case windowSizeMsg:
				go p.checkResize()

			case requestCursorPosMsg:
				p.execute(ansi.RequestCursorPosition)

			case RawMsg:
				p.execute(fmt.Sprint(msg.Msg))

			case printLineMessage:
				p.renderer.insertAbove(msg.messageBody)

			case repaintMsg:
				p.renderer.repaint()

			case clearScreenMsg:
				p.renderer.clearScreen()

			case ColorProfileMsg:
				p.renderer.setColorProfile(msg.Profile)
			}

			var cmd Cmd
			p.Model, cmd = p.Update(p.Model, msg) // run update
			cmds <- cmd                           // process command (if any)

			view := p.View(p.Model)
			switch view := view.(type) {
			case Frame:
				// Ensure we reset the cursor color on exit.
				if view.Cursor != nil {
					p.setCc = view.Cursor.Color
				}
			}

			p.renderer.render(view) //nolint:errcheck // send view to renderer
		}
	}
}

// Run initializes the program and runs its event loops, blocking until it gets
// terminated by either [Program.Quit], [Program.Kill], or its signal handler.
// Returns the final model.
func (p *Program[T]) Run() error {
	if err := p.Start(); err != nil {
		return err
	}
	return p.Wait()
}

func (p *Program[T]) Start() error {
	p.init()

	p.handlers = channelHandlers{}
	cmds := make(chan Cmd)
	p.errs = make(chan error)
	p.finished = make(chan struct{}, 1)

	if p.Input == nil {
		p.Input = os.Stdin
	}

	// The user has not set a custom input, so we need to check whether or
	// not standard input is a terminal. If it's not, we open a new TTY for
	// input. This will allow things to "just work" in cases where data was
	// piped in or redirected to the application.
	if f, ok := p.Input.(term.File); !p.IsInputTTY && (!ok || !term.IsTerminal(f.Fd())) {
		f, err := openInputTTY()
		if err != nil {
			return err
		}
		p.Input = f
		p.IsInputTTY = true
	}

	// Handle signals.
	if !p.IgnoreSignals {
		p.handlers.add(p.handleSignals())
	}

	// Recover from panics.
	if !p.DontCatchPanics {
		defer p.recoverFromPanic()
	}

	// Check if output is a TTY before entering raw mode, hiding the cursor and
	// so on.
	if err := p.initTerminal(); err != nil {
		return err
	}
	if p.renderer == nil {
		// If no renderer is set use the ferocious one.
		p.renderer = newCursedRenderer(p.Output, p.getenv("TERM"), p.useHardTabs)
	}

	// Get the color profile and send it to the program.
	if p.Profile == nil {
		p.Profile = Profile(colorprofile.Detect(p.Output, p.environ))
	}

	// Set the color profile on the renderer and send it to the program.
	p.renderer.setColorProfile(*p.Profile)
	go p.Send(ColorProfileMsg{*p.Profile})

	// Get the initial window size.
	resizeMsg := WindowSizeMsg{Width: p.width, Height: p.height}
	if p.ttyOutput != nil {
		// Set the initial size of the terminal.
		w, h, err := term.GetSize(p.ttyOutput.Fd())
		if err != nil {
			return err
		}

		resizeMsg.Width, resizeMsg.Height = w, h
	}

	// Send the initial size to the program.
	go p.Send(resizeMsg)
	p.renderer.resize(resizeMsg.Width, resizeMsg.Height)

	// Send the environment variables used by the program.
	go p.Send(EnvMsg(p.environ))

	// Init the input reader and initial model.
	if p.Input != nil {
		if err := p.initInputReader(); err != nil {
			return err
		}
	}

	// Hide the cursor before starting the renderer. This is handled by the
	// renderer so we don't need to write the sequence here.
	p.modes.Reset(ansi.TextCursorEnableMode)
	p.renderer.hideCursor()

	// Honor program startup options.
	if p.Title != "" {
		p.execute(ansi.SetWindowTitle(p.Title))
	}

	p.execute(ansi.SetBracketedPasteMode)
	p.modes.Set(ansi.BracketedPasteMode)

	// Start the renderer.
	p.startRenderer()

	// Initialize the program.
	var initCmd Cmd
	p.Model, initCmd = p.Init()
	if initCmd != nil {
		ch := make(chan struct{})
		p.handlers.add(ch)

		go func() {
			defer close(ch)

			select {
			case cmds <- initCmd:
			case <-p.Context.Done():
			}
		}()
	}

	// Render the initial view.
	p.renderer.render(p.View(p.Model)) //nolint:errcheck

	// Handle resize events.
	p.handlers.add(p.handleResize())

	// Process commands.
	p.handlers.add(p.handleCommands(cmds))

	go func() {
		// Run event loop, handle updates and draw.
		p.eventLoop(cmds)
		killed := p.Context.Err() != nil
		// Ensure we rendered the final state of the model.
		p.renderer.render(p.View(p.Model)) //nolint:errcheck
		// Restore terminal state.
		p.shutdown(killed)
	}()

	return nil
}

// Send sends a message to the main update function, effectively allowing
// messages to be injected from outside the program for interoperability
// purposes.
//
// If the program hasn't started yet this will be a blocking operation.
// If the program has already been terminated this will be a no-op, so it's safe
// to send messages after the program has exited.
func (p *Program[T]) Send(msg Msg) {
	select {
	case <-p.Context.Done():
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
func (p *Program[T]) Quit() {
	p.Send(Quit())
}

// Kill stops the program immediately and restores the former terminal state.
// The final render that you would normally see when quitting will be skipped.
// [program.Run] returns a [ErrProgramKilled] error.
func (p *Program[T]) Kill() {
	p.shutdown(true)
}

// Wait waits/blocks until the underlying Program finished shutting down.
func (p *Program[T]) Wait() error {
	defer func() {
		killed := p.Context.Err() != nil
		// Restore terminal state.
		p.shutdown(killed)
	}()

	select {
	case err := <-p.errs:
		return err
	case <-p.Context.Done():
		return nil
	case <-p.finished:
		return nil
	}
}

// execute writes the given sequence to the program output.
func (p *Program[T]) execute(seq string) {
	io.WriteString(p.Output, seq) //nolint:errcheck
}

// shutdown performs operations to free up resources and restore the terminal
// to its original state.
func (p *Program[T]) shutdown(kill bool) {
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
func (p *Program[T]) recoverFromPanic() {
	if r := recover(); r != nil {
		p.shutdown(true)
		fmt.Printf("Caught panic:\n\n%s\n\nRestoring terminal...\n\n", r)
		debug.PrintStack()
	}
}

// ReleaseTerminal restores the original terminal state and cancels the input
// reader. You can return control to the Program with RestoreTerminal.
func (p *Program[T]) ReleaseTerminal() error {
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
func (p *Program[T]) RestoreTerminal() error {
	atomic.StoreUint32(&p.ignoreSignals, 0)

	if err := p.initTerminal(); err != nil {
		return err
	}
	if err := p.initInputReader(); err != nil {
		return err
	}
	if p.modes.IsReset(ansi.AltScreenSaveCursorMode) {
		// entering alt screen already causes a repaint.
		go p.Send(repaintMsg{})
	}

	p.startRenderer()
	if p.modes.IsSet(ansi.BracketedPasteMode) {
		p.execute(ansi.SetBracketedPasteMode)
	}
	if p.activeEnhancements.modifyOtherKeys != 0 {
		p.execute(ansi.ModifyOtherKeys(p.activeEnhancements.modifyOtherKeys))
	}
	if p.activeEnhancements.kittyFlags != 0 {
		p.execute(ansi.PushKittyKeyboard(p.activeEnhancements.kittyFlags))
	}
	if p.modes.IsSet(ansi.FocusEventMode) {
		p.execute(ansi.SetFocusEventMode)
	}
	if p.modes.IsSet(ansi.ButtonEventMouseMode) || p.modes.IsSet(ansi.AnyEventMouseMode) {
		p.execute(ansi.SetButtonEventMouseMode)
		p.execute(ansi.SetSgrExtMouseMode)
	} else if p.modes.IsSet(ansi.AnyEventMouseMode) {
		p.execute(ansi.SetAnyEventMouseMode)
		p.execute(ansi.SetSgrExtMouseMode)
	}
	if p.modes.IsSet(ansi.GraphemeClusteringMode) {
		p.execute(ansi.SetGraphemeClusteringMode)
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
func (p *Program[T]) Println(args ...interface{}) {
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
func (p *Program[T]) Printf(template string, args ...interface{}) {
	p.msgs <- printLineMessage{
		messageBody: fmt.Sprintf(template, args...),
	}
}

// startRenderer starts the renderer.
func (p *Program[T]) startRenderer() {
	framerate := time.Second / time.Duration(p.FPS)
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
func (p *Program[T]) stopRenderer(kill bool) {
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

// sendKeyboardEnhancementsMsg sends a message with the active keyboard
// enhancements to the program after a short timeout, or immediately if the
// keyboard enhancements have been read from the terminal.
func (p *Program[T]) sendKeyboardEnhancementsMsg() {
	if runtime.GOOS == "windows" {
		// We use the Windows Console API which supports keyboard enhancements.
		p.Send(KeyboardEnhancementsMsg{})
		return
	}

	// Initial keyboard enhancements message. Ensure we send a message so that
	// terminals that don't support the requested features can disable them.
	const timeout = 100 * time.Millisecond
	select {
	case <-time.After(timeout):
		p.Send(KeyboardEnhancementsMsg{})
	case <-p.keyboardc:
	}
}

// requestKeyboardEnhancements tries to enable keyboard enhancements and read
// the active keyboard enhancements from the terminal.
func (p *Program[T]) requestKeyboardEnhancements() {
	if p.requestedEnhancements.modifyOtherKeys > 0 {
		p.execute(ansi.ModifyOtherKeys(p.requestedEnhancements.modifyOtherKeys))
		p.execute(ansi.RequestModifyOtherKeys)
	}
	if p.requestedEnhancements.kittyFlags > 0 {
		p.execute(ansi.PushKittyKeyboard(p.requestedEnhancements.kittyFlags))
		p.execute(ansi.RequestKittyKeyboard)
	}
}
