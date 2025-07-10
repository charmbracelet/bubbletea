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
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/input"
	"github.com/charmbracelet/x/term"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/sync/errgroup"
)

// ErrProgramPanic is returned by [Program.Run] when the program recovers from a panic.
var ErrProgramPanic = errors.New("program experienced a panic")

// ErrProgramKilled is returned by [Program.Run] when the program gets killed.
var ErrProgramKilled = errors.New("program was killed")

// ErrInterrupted is returned by [Program.Run] when the program get a SIGINT
// signal, or when it receives a [InterruptMsg].
var ErrInterrupted = errors.New("program was interrupted")

// Msg contain data from the result of a IO operation. Msgs trigger the update
// function and, henceforth, the UI.
type Msg any

// Model contains the program's state as well as its core functions.
type Model interface {
	// Init is the first function that will be called. It returns an optional
	// initial command. To not perform an initial command return nil.
	Init() Cmd

	// Update is called when a message is received. Use it to inspect messages
	// and, in response, update the model and/or send a command.
	Update(Msg) (Model, Cmd)
}

// ViewModel is an optional interface that can be implemented by the main model
// to provide a view. If the main model does not implement a view interface,
// the program won't render anything.
type ViewModel interface {
	// View renders the program's UI, which is just a string. The view is
	// rendered after every Update.
	View() string
}

// Cursor represents a cursor on the terminal screen.
type Cursor struct {
	// Position is a [Position] that determines the cursor's position on the
	// screen relative to the top left corner of the frame.
	Position

	// Color is a [color.Color] that determines the cursor's color.
	Color color.Color

	// Shape is a [CursorShape] that determines the cursor's shape.
	Shape CursorShape

	// Blink is a boolean that determines whether the cursor should blink.
	Blink bool
}

// NewCursor returns a new cursor with the default settings and the given
// position.
func NewCursor(x, y int) *Cursor {
	return &Cursor{
		Position: Position{X: x, Y: y},
		Color:    nil,
		Shape:    CursorBlock,
		Blink:    true,
	}
}

// CursorModel is an optional interface that can be implemented by the main
// model to provide a view that manages the cursor. If the main model does not
// implement a view interface, the program won't render anything.
type CursorModel interface {
	// View renders the program's UI, which is just a string. The view is
	// rendered after every Update. The cursor is optional, if it's nil the
	// cursor will be hidden.
	// Use [NewCursor] to quickly create a cursor for a given position with
	// default styles.
	View() (string, *Cursor)
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
	withColorProfile
	withGraphemeClustering
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

	// externalCtx is a context that was passed in via WithContext, otherwise defaulting
	// to ctx.Background() (in case it was not), the internal context is derived from it.
	externalCtx context.Context

	// ctx is the programs's internal context for signalling internal teardown.
	// It is built and derived from the externalCtx in NewProgram().
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
	environ environ

	// where to read inputs from, this will usually be os.Stdin.
	input io.Reader
	// ttyInput is null if input is not a TTY.
	ttyInput              term.File
	previousTtyInputState *term.State
	inputReader           *input.Reader
	traceInput            bool // true if input should be traced
	readLoopDone          chan struct{}
	mouseMode             bool // indicates whether we should enable mouse on Windows

	// modes keeps track of terminal modes that have been enabled or disabled.
	modes         ansi.Modes
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

	// stores the requested keyboard enhancements.
	requestedEnhancements KeyboardEnhancements
	// activeEnhancements stores the active keyboard enhancements read from the
	// terminal.
	activeEnhancements KeyboardEnhancements

	// When a program is suspended, the terminal state is saved and the program
	// is paused. This saves the terminal colors state so they can be restored
	// when the program is resumed.
	setBg, setFg, setCc color.Color

	// Initial window size. Mainly used for testing.
	width, height int

	// whether to use hard tabs to optimize cursor movements
	useHardTabs bool
	// whether to use backspace to optimize cursor movements
	useBackspace bool
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
func NewProgram(model Model, opts ...ProgramOption) *Program {
	p := &Program{
		initialModel: model,
		msgs:         make(chan Msg),
		rendererDone: make(chan struct{}),
		modes:        ansi.Modes{},
	}

	// Apply all options to the program.
	for _, opt := range opts {
		opt(p)
	}

	// A context can be provided with a ProgramOption, but if none was provided
	// we'll use the default background context.
	if p.externalCtx == nil {
		p.externalCtx = context.Background()
	}
	// Initialize context and teardown channel.
	p.ctx, p.cancel = context.WithCancel(p.externalCtx)

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
	enableTracing := func() {
		// Enable different types of tracing.
		if output, _ := strconv.ParseBool(os.Getenv("TEA_TRACE_OUTPUT")); output {
			p.output.trace = true
		}
		if input, _ := strconv.ParseBool(os.Getenv("TEA_TRACE_INPUT")); input {
			p.traceInput = true
		}
	}
	tracePath, traceOk := os.LookupEnv("TEA_TRACE")
	traceEnabled, err := strconv.ParseBool(os.Getenv("TEA_TRACE"))
	switch {
	case err != nil && traceOk && len(tracePath) > 0:
		// We have a trace filepath.
		if _, logErr := LogToFile(tracePath, "bubbletea"); logErr == nil {
			enableTracing()
		}
	case err == nil && traceEnabled:
		// Use the default [log] output.
		enableTracing()
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

			case s := <-sig:
				if atomic.LoadUint32(&p.ignoreSignals) == 0 {
					switch s {
					case syscall.SIGINT:
						p.msgs <- InterruptMsg{}
					default:
						p.msgs <- QuitMsg{}
					}
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
						defer func() {
							if r := recover(); r != nil {
								p.recoverFromPanic(r)
							}
						}()
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

			case InterruptMsg:
				return model, ErrInterrupted

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

			case enableMouseCellMotionMsg:
				p.enableMouse(false)

			case enableMouseAllMotionMsg:
				p.enableMouse(true)

			case disableMouseMotionMsg:
				p.disableMouse()

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
					c, ok := colorful.MakeColor(msg.Color)
					if ok {
						p.execute(ansi.SetBackgroundColor(c.Hex()))
					}
				} else {
					p.execute(ansi.ResetBackgroundColor)
				}
				p.setBg = msg.Color

			case setForegroundColorMsg:
				if msg.Color != nil {
					c, ok := colorful.MakeColor(msg.Color)
					if ok {
						p.execute(ansi.SetForegroundColor(c.Hex()))
					}
				} else {
					p.execute(ansi.ResetForegroundColor)
				}
				p.setFg = msg.Color

			case setCursorColorMsg:
				if msg.Color != nil {
					c, ok := colorful.MakeColor(msg.Color)
					if ok {
						p.execute(ansi.SetCursorColor(c.Hex()))
					}
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

			case enableKeyboardEnhancementsMsg:
				if isWindows() {
					// We use the Windows Console API which supports keyboard
					// enhancements.
					// Send an empty message to tell the user we support
					// keyboard enhancements on Windows.
					go p.Send(KeyboardEnhancementsMsg{})
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

			case disableKeyboardEnhancementsMsg:
				if isWindows() {
					// We use the Windows Console API which supports keyboard
					// enhancements.
					break
				}

				if p.activeEnhancements.modifyOtherKeys > 0 {
					p.execute(ansi.ResetModifyOtherKeys)
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
				p.execute(ansi.RequestNameVersion)

			case requestCapabilityMsg:
				p.execute(ansi.RequestTermcap(string(msg)))

			case BatchMsg:
				for _, cmd := range msg {
					select {
					case <-p.ctx.Done():
						return model, nil
					case cmds <- cmd:
					}
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

							_ = g.Wait() // wait for all commands from batch msg to finish
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
				p.execute(ansi.RequestCursorPositionReport)

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
			model, cmd = model.Update(msg) // run update

			select {
			case <-p.ctx.Done():
				return model, nil
			case cmds <- cmd: // process command (if any)
			}

			p.render(model) // render view
		}
	}
}

// hasView returns true if the model has a view.
func hasView(model Model) (ok bool) {
	switch model.(type) {
	case ViewModel, CursorModel:
		ok = true
	}
	return
}

// render renders the given view to the renderer.
func (p *Program) render(model Model) {
	var view string
	var cur *Cursor
	switch model := model.(type) {
	case ViewModel:
		view = model.View()
	case CursorModel:
		view, cur = model.View()
	}

	// Ensure we reset the cursor color on exit.
	if cur != nil {
		p.setCc = cur.Color
	}

	p.renderer.render(view, cur) // send view to renderer
}

// Run initializes the program and runs its event loops, blocking until it gets
// terminated by either [Program.Quit], [Program.Kill], or its signal handler.
// Returns the final model.
func (p *Program) Run() (returnModel Model, returnErr error) {
	p.handlers = channelHandlers{}
	cmds := make(chan Cmd)
	p.errs = make(chan error, 1)

	p.finished = make(chan struct{})
	defer func() {
		close(p.finished)
	}()

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
		defer func() {
			if r := recover(); r != nil {
				returnErr = fmt.Errorf("%w: %w", ErrProgramKilled, ErrProgramPanic)
				p.recoverFromPanic(r)
			}
		}()
	}

	// Check if output is a TTY before entering raw mode, hiding the cursor and
	// so on.
	if err := p.initTerminal(); err != nil {
		return p.initialModel, err
	}

	// Get the initial window size.
	resizeMsg := WindowSizeMsg{Width: p.width, Height: p.height}
	if p.ttyOutput != nil {
		// Set the initial size of the terminal.
		w, h, err := term.GetSize(p.ttyOutput.Fd())
		if err != nil {
			return p.initialModel, fmt.Errorf("bubbletea: error getting terminal size: %w", err)
		}

		resizeMsg.Width, resizeMsg.Height = w, h
	}

	if p.renderer == nil { //nolint:nestif
		if hasView(p.initialModel) {
			stdr, ok := os.LookupEnv("TEA_STANDARD_RENDERER")
			if has, _ := strconv.ParseBool(stdr); ok && has {
				p.renderer = newRenderer(p.output)
			} else {
				// If no renderer is set use the cursed one.
				p.renderer = newCursedRenderer(
					p.output,
					p.getenv("TERM"),
					resizeMsg.Width,
					resizeMsg.Height,
					p.useHardTabs,
					p.useBackspace,
					p.ttyInput == nil,
				)
			}
		} else {
			// If the model has no view we don't need a renderer.
			p.renderer = &nilRenderer{}
		}
	}

	// Get the color profile and send it to the program.
	if !p.startupOptions.has(withColorProfile) {
		p.profile = colorprofile.Detect(p.output.Writer(), p.environ)
	}

	// Set the color profile on the renderer and send it to the program.
	p.renderer.setColorProfile(p.profile)
	go p.Send(ColorProfileMsg{p.profile})

	// Send the initial size to the program.
	go p.Send(resizeMsg)
	p.renderer.resize(resizeMsg.Width, resizeMsg.Height)

	// Send the environment variables used by the program.
	go p.Send(EnvMsg(p.environ))

	// Init the input reader and initial model.
	model := p.initialModel
	if p.input != nil {
		if err := p.initInputReader(false); err != nil {
			return model, err
		}
	}

	// Hide the cursor before starting the renderer. This is handled by the
	// renderer so we don't need to write the sequence here.
	p.modes.Reset(ansi.TextCursorEnableMode)
	p.renderer.hideCursor()

	// Honor program startup options.
	if p.startupTitle != "" {
		p.execute(ansi.SetWindowTitle(p.startupTitle))
	}
	if p.startupOptions&withAltScreen != 0 {
		// Enter alternate screen mode. This is handled by the renderer so we
		// don't need to write the sequence here.
		p.modes.Set(ansi.AltScreenSaveCursorMode)
		p.renderer.enterAltScreen()
	}
	if p.startupOptions&withoutBracketedPaste == 0 {
		p.execute(ansi.SetBracketedPasteMode)
		p.modes.Set(ansi.BracketedPasteMode)
	}
	if p.startupOptions&withGraphemeClustering != 0 {
		p.execute(ansi.SetGraphemeClusteringMode)
		p.execute(ansi.RequestGraphemeClusteringMode)
		// We store the state of grapheme clustering after we query it and get
		// a response in the eventLoop.
	}

	// Enable mouse mode.
	cellMotion := p.startupOptions&withMouseCellMotion != 0
	allMotion := p.startupOptions&withMouseAllMotion != 0
	if cellMotion || allMotion {
		p.enableMouse(allMotion)
	}

	if p.startupOptions&withReportFocus != 0 {
		p.execute(ansi.SetFocusEventMode)
		p.modes.Set(ansi.FocusEventMode)
	}

	if !isWindows() {
		// Enable unambiguous keys using whichever protocol the terminal prefer.
		p.requestedEnhancements.kittyFlags |= ansi.KittyDisambiguateEscapeCodes
		if p.requestedEnhancements.modifyOtherKeys == 0 {
			p.requestedEnhancements.modifyOtherKeys = 1 // mode 1
		}
		// We use the Windows Console API which supports keyboard
		// enhancements.
		p.requestKeyboardEnhancements()
	} else {
		// Send an empty message to tell the user we support
		// keyboard enhancements on Windows.
		go p.Send(KeyboardEnhancementsMsg{})
	}

	// Start the renderer.
	p.startRenderer()

	// Initialize the program.
	initCmd := model.Init()
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
	p.render(model)

	// Handle resize events.
	p.handlers.add(p.handleResize())

	// Process commands.
	p.handlers.add(p.handleCommands(cmds))

	// Run event loop, handle updates and draw.
	model, err := p.eventLoop(model, cmds)

	if err == nil && len(p.errs) > 0 {
		err = <-p.errs // Drain a leftover error in case eventLoop crashed.
	}

	killed := p.externalCtx.Err() != nil || p.ctx.Err() != nil || err != nil
	if killed {
		if err == nil && p.externalCtx.Err() != nil {
			// Return also as context error the cancellation of an external context.
			// This is the context the user knows about and should be able to act on.
			err = fmt.Errorf("%w: %w", ErrProgramKilled, p.externalCtx.Err())
		} else if err == nil && p.ctx.Err() != nil {
			// Return only that the program was killed (not the internal mechanism).
			// The user does not know or need to care about the internal program context.
			err = ErrProgramKilled
		} else {
			// Return that the program was killed and also the error that caused it.
			err = fmt.Errorf("%w: %w", ErrProgramKilled, err)
		}
	} else {
		// Graceful shutdown of the program (not killed):
		// Ensure we rendered the final state of the model.
		p.render(model)
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
	_, _ = io.WriteString(p.output, seq)
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
	})
}

// recoverFromPanic recovers from a panic, prints the stack trace, and restores
// the terminal to a usable state.
func (p *Program) recoverFromPanic(r any) {
	select {
	case p.errs <- ErrProgramPanic:
	default:
	}
	p.cancel() // Just in case a previous shutdown has failed.
	p.shutdown(true)
	// We use "\r\n" to ensure the output is formatted even when restoring the
	// terminal does not work or when raw mode is still active.
	rec := strings.ReplaceAll(fmt.Sprintf("%s", r), "\n", "\r\n")
	fmt.Fprintf(os.Stderr, "Caught panic:\r\n\r\n%s\r\n\r\nRestoring terminal...\r\n\r\n", rec)
	stack := strings.ReplaceAll(fmt.Sprintf("%s\n", debug.Stack()), "\n", "\r\n")
	fmt.Fprint(os.Stderr, stack)
	if v, err := strconv.ParseBool(os.Getenv("TEA_DEBUG")); err == nil && v {
		f, err := os.Create(fmt.Sprintf("bubbletea-panic-%d.log", time.Now().Unix()))
		if err == nil {
			defer f.Close()        //nolint:errcheck
			fmt.Fprintln(f, rec)   //nolint:errcheck
			fmt.Fprintln(f)        //nolint:errcheck
			fmt.Fprintln(f, stack) //nolint:errcheck
		}
	}
}

// ReleaseTerminal restores the original terminal state and cancels the input
// reader. You can return control to the Program with RestoreTerminal.
func (p *Program) ReleaseTerminal() error {
	return p.releaseTerminal(false)
}

func (p *Program) releaseTerminal(reset bool) error {
	atomic.StoreUint32(&p.ignoreSignals, 1)
	if p.inputReader != nil {
		p.inputReader.Cancel()
	}

	p.waitForReadLoop()

	if p.renderer != nil {
		p.stopRenderer(false)
		if reset {
			p.renderer.reset()
		}
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
	if err := p.initInputReader(false); err != nil {
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
		p.execute(ansi.KeyModifierOptions(4, p.activeEnhancements.modifyOtherKeys)) //nolint:mnd
	}
	if p.activeEnhancements.kittyFlags != 0 {
		p.execute(ansi.PushKittyKeyboard(p.activeEnhancements.kittyFlags))
	}
	if p.modes.IsSet(ansi.FocusEventMode) {
		p.execute(ansi.SetFocusEventMode)
	}
	if p.modes.IsSet(ansi.ButtonEventMouseMode) || p.modes.IsSet(ansi.AnyEventMouseMode) {
		if p.startupOptions&withMouseCellMotion != 0 {
			p.execute(ansi.SetButtonEventMouseMode)
			p.execute(ansi.SetSgrExtMouseMode)
		} else if p.startupOptions&withMouseAllMotion != 0 {
			p.execute(ansi.SetAnyEventMouseMode)
			p.execute(ansi.SetSgrExtMouseMode)
		}
	}
	if p.modes.IsSet(ansi.GraphemeClusteringMode) {
		p.execute(ansi.SetGraphemeClusteringMode)
	}

	// Restore terminal colors.
	if p.setBg != nil {
		c, ok := colorful.MakeColor(p.setBg)
		if ok {
			p.execute(ansi.SetBackgroundColor(c.Hex()))
		}
	}
	if p.setFg != nil {
		c, ok := colorful.MakeColor(p.setFg)
		if ok {
			p.execute(ansi.SetForegroundColor(c.Hex()))
		}
	}
	if p.setCc != nil {
		c, ok := colorful.MakeColor(p.setCc)
		if ok {
			p.execute(ansi.SetCursorColor(c.Hex()))
		}
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
func (p *Program) Println(args ...any) {
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
func (p *Program) Printf(template string, args ...any) {
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
				_ = p.renderer.flush()
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
		_ = p.renderer.flush()
	}

	_ = p.renderer.close()
}

// requestKeyboardEnhancements tries to enable keyboard enhancements and read
// the active keyboard enhancements from the terminal.
func (p *Program) requestKeyboardEnhancements() {
	if p.requestedEnhancements.modifyOtherKeys > 0 {
		p.execute(ansi.KeyModifierOptions(4, p.requestedEnhancements.modifyOtherKeys)) //nolint:mnd
		p.execute(ansi.QueryModifyOtherKeys)
	}
	if p.requestedEnhancements.kittyFlags > 0 {
		p.execute(ansi.PushKittyKeyboard(p.requestedEnhancements.kittyFlags))
		p.execute(ansi.RequestKittyKeyboard)
	}
}

// enableMouse enables mouse events on the terminal. When all is true, it will
// enable [ansi.AnyEventMouseMode], otherwise, it will use
// [ansi.ButtonEventMouseMode].
// Note this has no effect on Windows since we use the Windows Console API.
func (p *Program) enableMouse(all bool) {
	if isWindows() {
		// XXX: This is used to enable mouse mode on Windows. We need
		// to reinitialize the cancel reader to get the mouse events to
		// work.
		if !p.mouseMode {
			p.mouseMode = true
			if p.inputReader != nil {
				// Only reinitialize if the input reader has been initialized.
				_ = p.initInputReader(true)
			}
		}
	}

	if all {
		p.execute(ansi.SetAnyEventMouseMode + ansi.SetSgrExtMouseMode)
		p.modes.Set(ansi.AnyEventMouseMode, ansi.SgrExtMouseMode)
	} else {
		p.execute(ansi.SetButtonEventMouseMode + ansi.SetSgrExtMouseMode)
		p.modes.Set(ansi.ButtonEventMouseMode, ansi.SgrExtMouseMode)
	}
}

// disableMouse disables mouse events on the terminal.
// Note this has no effect on Windows since we use the Windows Console API.
func (p *Program) disableMouse() {
	if isWindows() {
		// XXX: On Windows, mouse mode is enabled on the input reader
		// level. We need to instruct the input reader to stop reading
		// mouse events.
		if p.mouseMode {
			p.mouseMode = false
			if p.inputReader != nil {
				// Only reinitialize if the input reader has been initialized.
				_ = p.initInputReader(true)
			}
		}
	}

	var modes []ansi.Mode
	if p.modes.IsSet(ansi.AnyEventMouseMode) {
		modes = append(modes, ansi.AnyEventMouseMode)
	}
	if p.modes.IsSet(ansi.ButtonEventMouseMode) {
		modes = append(modes, ansi.ButtonEventMouseMode)
	}
	if len(modes) > 0 {
		modes = append(modes, ansi.SgrExtMouseMode)
		for _, m := range modes {
			// We could combine all of these modes into one single sequence,
			// but we're being cautious here for terminals that might not support
			// that format i.e. `CSI ? 10003 ; 1006 l`.
			p.execute(ansi.ResetMode(m))
			p.modes.Reset(m)
		}
	}
}
