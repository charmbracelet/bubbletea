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
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"
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
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/cancelreader"
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
type Msg = uv.Event

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

// ViewableModel is an optional interface that can be implemented by the main
// model to provide a view that can be composed of multiple layers. If the
// main model does not implement a view interface, the program won't render
// anything.
type ViewableModel interface {
	// View returns a [View] that contains the layers to be rendered. The
	// layers are rendered based on their z-index, with the lowest z-index
	// rendered first and the highest z-index rendered last. If some layers
	// have the same z-index, they are rendered in the order they were added to
	// the view.
	// The cursor is optional, if it's nil the cursor will be hidden.
	View() View
}

// Buffer represents a terminal cell buffer that defines the current state of
// the terminal screen.
type Buffer = uv.Buffer

// Screen represents a read writable canvas that can be used to render
// components on the terminal screen.
type Screen = uv.Screen

// Rectangle represents a rectangular area with two points: the top left corner
// and the bottom right corner. It is used to define the area where components
// will be rendered on the terminal screen.
type Rectangle = uv.Rectangle

// Layer represents a drawable component on a [Screen].
type Layer interface {
	// Draw renders the component on the given [Screen] within the specified
	// [Rectangle]. The component should draw itself within the bounds of the
	// rectangle, which is defined by the top left corner (x0, y0) and the
	// bottom right corner (x1, y1).
	Draw(s Screen, r Rectangle)
}

// Hittable is an interface that can be implemented by a [Layer] to test
// whether a layer was hit by a mouse event.
type Hittable interface {
	// Hit tests the layer against the given position. If the position is
	// inside the layer, it returns the layer ID that was hit. If no
	// layer was hit, it returns an empty string.
	Hit(x, y int) string
}

// NewView is a helper function to create a new [View] with the given string or
// [Layer].
func NewView(s any) View {
	var view View
	view.SetContent(s)
	return view
}

// View represents a terminal view that can be composed of multiple layers.
// It can also contain a cursor that will be rendered on top of the layers.
type View struct {
	Layer           Layer
	Cursor          *Cursor
	BackgroundColor color.Color
	ForegroundColor color.Color
	WindowTitle     string
	ProgressBar     *ProgressBar

	// AltScreen puts the program in the alternate screen buffer
	// (i.e. the program goes into full window mode). Note that the altscreen will
	// be automatically exited when the program quits.
	//
	// Example:
	//
	//	func (m model) View() tea.View {
	//	    v := tea.NewView("Hello, World!")
	//	    v.AltScreen = true
	//	    return v
	//	}
	//
	AltScreen bool

	// ReportFocus enables reporting when the terminal gains and loses focus.
	// When this is enabled [FocusMsg] and [BlurMsg] messages will be sent to
	// your Update method.
	//
	// Note that while most terminals and multiplexers support focus reporting,
	// some do not. Also note that tmux needs to be configured to report focus
	// events.
	ReportFocus bool

	// DisableBracketedPasteMode disables bracketed paste mode for this view.
	DisableBracketedPasteMode bool

	// MouseMode sets the mouse mode for this view. It can be one of
	// [MouseModeNone], [MouseModeCellMotion], or [MouseModeAllMotion].
	MouseMode MouseMode

	// DisableKeyEnhancements disables all key enhancements for this view.
	DisableKeyEnhancements bool

	// KeyReleases enables support for reporting key release events. This is
	// useful for terminals that support the Kitty keyboard protocol "Report
	// event types" progressive enhancement feature.
	KeyReleases bool

	// UniformKeyLayout enables support for reporting key events as though they
	// were on a PC-101 layout. This is useful for uniform key event reporting
	// across different keyboard layouts. This is equivalent to the Kitty
	// keyboard protocol "Report alternate keys" and "Report all keys as escape
	// codes" progressive enhancement features.
	UniformKeyLayout bool
}

// SetContent sets the content of the view to the value.
func (v *View) SetContent(s any) {
	switch vi := s.(type) {
	case string:
		v.Layer = uv.NewStyledString(vi)
	case fmt.Stringer:
		v.Layer = uv.NewStyledString(vi.String())
	case Layer:
		v.Layer = vi
	default:
		v.Layer = uv.NewStyledString(fmt.Sprintf("%v", vi))
	}
}

// MouseMode represents the mouse mode of a view.
type MouseMode int

const (
	// MouseModeNone disables mouse events.
	MouseModeNone MouseMode = iota

	// MouseModeCellMotion enables mouse click, release, and wheel events.
	// Mouse movement events are also captured if a mouse button is pressed
	// (i.e., drag events). Cell motion mode is better supported than all
	// motion mode.
	//
	// This will try to enable the mouse in extended mode (SGR), if that is not
	// supported by the terminal it will fall back to normal mode (X10).
	MouseModeCellMotion

	// MouseModeAllMotion enables all mouse events, including click, release,
	// wheel, and movement events. You will receive mouse movement events even
	// when no buttons are pressed.
	//
	// This will try to enable the mouse in extended mode (SGR), if that is not
	// supported by the terminal it will fall back to normal mode (X10).
	MouseModeAllMotion
)

// ProgressBarState represents the state of the progress bar.
type ProgressBarState int

// Progress bar states.
const (
	ProgressBarNone ProgressBarState = iota
	ProgressBarDefault
	ProgressBarError
	ProgressBarIndeterminate
	ProgressBarWarning
)

// String return a human-readable value for the given [ProgressBarState].
func (s ProgressBarState) String() string {
	return [...]string{
		"None",
		"Default",
		"Error",
		"Indeterminate",
		"Warning",
	}[s]
}

// ProgressBar represents the terminal progress bar.
//
// Support depends on the terminal.
//
// See https://learn.microsoft.com/en-us/windows/terminal/tutorials/progress-bar-sequences
type ProgressBar struct {
	// State is the current state of the progress bar. It can be one of
	// [ProgressBarNone], [ProgressBarDefault], [ProgressBarError],
	// [ProgressBarIndeterminate], and [ProgressBarWarn].
	State ProgressBarState
	// Value is the current value of the progress bar. It should be between
	// 0 and 100.
	Value int
}

// NewProgressBar returns a new progress bar with the given state and value.
// The value is ignored if the state is [ProgressBarNone] or
// [ProgressBarIndeterminate].
func NewProgressBar(state ProgressBarState, value int) *ProgressBar {
	return &ProgressBar{
		State: state,
		Value: min(max(value, 0), 100),
	}
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
	// Input sets the input which, by default, is stdin when nil. In most cases
	// you won't need to use this. To disable input entirely use
	// [Program.DisableInput].
	//
	//	p := NewProgram(model)
	//	p.Input = customInputFile
	Input io.Reader

	// Output sets the output which, by default, is stdout. In most cases you
	// won't need to use this.
	Output io.Writer

	// Env sets the environment variables that the program will use. This
	// useful when the program is running in a remote session (e.g. SSH) and
	// you want to pass the environment variables from the remote session to
	// the program.
	//
	// Example:
	//
	//	var sess ssh.Session // ssh.Session is a type from the github.com/charmbracelet/ssh package
	//	pty, _, _ := sess.Pty()
	//	environ := append(sess.Environ(), "TERM="+pty.Term)
	//	p := tea.NewProgram(model)
	//	p.Env = environ
	Env []string

	// DisableInput disables all input. This is useful for programs that
	// don't need input, like a progress bar or a spinner.
	DisableInput bool

	// DisableSignalHandler disables the signal handler that Bubble Tea sets up
	// for Programs. This is useful if you want to handle signals yourself.
	DisableSignalHandler bool

	// DisableCatchPanics disables the panic catching that Bubble Tea does by
	// default. If panic catching is disabled the terminal will be in a fairly
	// unusable state after a panic because Bubble Tea will not perform its usual
	// cleanup on exit.
	DisableCatchPanics bool

	// IgnoreSignals will ignore OS signals. This is mainly useful for testing.
	IgnoreSignals bool

	// Filter supplies an event filter that will be invoked before Bubble Tea
	// processes a tea.Msg. The event filter can return any tea.Msg which will
	// then get handled by Bubble Tea instead of the original event. If the
	// event filter returns nil, the event will be ignored and Bubble Tea will
	// not process it.
	//
	// As an example, this could be used to prevent a program from shutting
	// down if there are unsaved changes.
	//
	// Example:
	//
	//	func filter(m tea.Model, msg tea.Msg) tea.Msg {
	//		if _, ok := msg.(tea.QuitMsg); !ok {
	//			return msg
	//		}
	//
	//		model := m.(myModel)
	//		if model.hasChanges {
	//			return nil
	//		}
	//
	//		return msg
	//	}
	//
	//	p := tea.NewProgram(Model{});
	//	p.Filter = filter
	//
	//	if _,err := p.Run(context.Background()); err != nil {
	//		fmt.Println("Error running program:", err)
	//		os.Exit(1)
	//	}
	Filter func(Model, Msg) Msg

	// FPS sets a custom maximum FPS at which the renderer should run. If less
	// than 1, the default value of 60 will be used. If over 120, the FPS will
	// be capped at 120.
	FPS int

	// Set the initial size of the terminal window. This is useful when you
	// need to set the initial size of the terminal window, for example during
	// testing or when you want to run your program in a non-interactive
	// environment.
	InitialWidth, InitialHeight int

	// ColorProfile when not nil, sets the color profile that the program will
	// use. This is useful when you want to force a specific color profile. By
	// default, Bubble Tea will try to detect the terminal's color profile from
	// environment variables and terminfo capabilities. Use [Program.Env] to
	// set custom environment variables.
	ColorProfile *colorprofile.Profile

	// InitialModel is the initial model for the program and is the only
	// required field when creating a new program.
	InitialModel Model

	// handlers is a list of channels that need to be waited on before the
	// program can exit.
	handlers channelHandlers

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
	output    io.Writer
	outputBuf bytes.Buffer // buffer used to queue commands to be sent to the output

	// ttyOutput is null if output is not a TTY.
	ttyOutput           term.File
	previousOutputState *term.State
	renderer            renderer

	// the environment variables for the program, defaults to os.Environ().
	environ uv.Environ
	// the program's logger for debugging.
	logger uv.Logger

	// where to read inputs from, this will usually be os.Stdin.
	input io.Reader
	// ttyInput is null if input is not a TTY.
	ttyInput              term.File
	previousTtyInputState *term.State
	cancelReader          cancelreader.CancelReader
	inputScanner          *uv.TerminalReader
	readLoopDone          chan struct{}

	// modes keeps track of terminal modes that have been enabled or disabled.
	ignoreSignals uint32

	// initialized is true when the program has been initialized.
	initialized atomic.Bool

	// ticker is the ticker that will be used to write to the renderer.
	ticker *time.Ticker

	// once is used to stop the renderer.
	once sync.Once

	// rendererDone is used to stop the renderer.
	rendererDone chan struct{}

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

// ResumeMsg can be listen to do something once a program is resumed back
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

// NewProgram creates a new [Program].
func NewProgram(model Model) *Program {
	p := &Program{
		InitialModel: model,
	}

	tracePath, traceOk := os.LookupEnv("TEA_TRACE")
	if traceOk && len(tracePath) > 0 {
		// We have a trace filepath.
		if f, err := os.OpenFile(tracePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o600); err == nil {
			p.logger = log.New(f, "bubbletea: ", log.LstdFlags|log.Lshortfile)
		}
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
					if !p.DisableCatchPanics {
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
			msg = p.translateInputEvent(msg)

			// Filter messages.
			if p.Filter != nil {
				msg = p.Filter(model, msg)
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

			case MouseMsg:
				for _, m := range p.renderer.hit(msg) {
					go p.Send(m) // send hit messages
				}

			case readClipboardMsg:
				p.execute(ansi.RequestSystemClipboard)

			case setClipboardMsg:
				p.execute(ansi.SetSystemClipboard(string(msg)))

			case readPrimaryClipboardMsg:
				p.execute(ansi.RequestPrimaryClipboard)

			case setPrimaryClipboardMsg:
				p.execute(ansi.SetPrimaryClipboard(string(msg)))

			case backgroundColorMsg:
				p.execute(ansi.RequestBackgroundColor)

			case foregroundColorMsg:
				p.execute(ansi.RequestForegroundColor)

			case cursorColorMsg:
				p.execute(ansi.RequestCursorColor)

			case execMsg:
				// NB: this blocks.
				p.exec(msg.cmd, msg.fn)

			case terminalVersion:
				p.execute(ansi.RequestNameVersion)

			case requestCapabilityMsg:
				p.execute(ansi.RequestTermcap(string(msg)))

			case BatchMsg:
				go p.execBatchMsg(msg)
				continue

			case sequenceMsg:
				go p.execSequenceMsg(msg)
				continue

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
	case ViewModel, CursorModel, ViewableModel:
		ok = true
	}
	return
}

// render renders the given view to the renderer.
func (p *Program) render(model Model) {
	var view View
	switch model := model.(type) {
	case ViewModel, CursorModel:
		var frame string
		switch model := model.(type) {
		case ViewModel:
			frame = model.View()
		case CursorModel:
			frame, view.Cursor = model.View()
		}
		view.Layer = uv.NewStyledString(frame)
	case ViewableModel:
		view = model.View()
	}
	if p.renderer != nil {
		p.renderer.render(view) // send view to renderer
	}
}

func (p *Program) execSequenceMsg(msg sequenceMsg) {
	if !p.DisableCatchPanics {
		defer func() {
			if r := recover(); r != nil {
				p.recoverFromGoPanic(r)
			}
		}()
	}

	// Execute commands one at a time, in order.
	for _, cmd := range msg {
		if cmd == nil {
			continue
		}
		msg := cmd()
		switch msg := msg.(type) {
		case BatchMsg:
			p.execBatchMsg(msg)
		case sequenceMsg:
			p.execSequenceMsg(msg)
		default:
			p.Send(msg)
		}
	}
}

func (p *Program) execBatchMsg(msg BatchMsg) {
	if !p.DisableCatchPanics {
		defer func() {
			if r := recover(); r != nil {
				p.recoverFromGoPanic(r)
			}
		}()
	}

	// Execute commands one at a time.
	var wg sync.WaitGroup
	for _, cmd := range msg {
		if cmd == nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			if !p.DisableCatchPanics {
				defer func() {
					if r := recover(); r != nil {
						p.recoverFromGoPanic(r)
					}
				}()
			}

			msg := cmd()
			switch msg := msg.(type) {
			case BatchMsg:
				p.execBatchMsg(msg)
			case sequenceMsg:
				p.execSequenceMsg(msg)
			default:
				p.Send(msg)
			}
		}()
	}

	wg.Wait() // wait for all commands from batch msg to finish
}

func (p *Program) init(ctx context.Context) {
	if p.initialized.Load() {
		return
	}
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.msgs = make(chan Msg)
	p.errs = make(chan error, 1)
	p.rendererDone = make(chan struct{})
	p.initialized.Store(true)
}

// Run initializes the program and runs its event loops, blocking until it gets
// terminated by either [Program.Quit], [Program.Kill], or its signal handler.
// Returns the final model.
func (p *Program) Run(ctx context.Context) (returnModel Model, returnErr error) {
	if p.InitialModel == nil {
		return nil, errors.New("bubbletea: InitialModel cannot be nil")
	}

	// A context can be provided with a ProgramOption, but if none was provided
	// we'll use the default background context.
	if ctx == nil {
		ctx = context.Background()
	}

	p.init(ctx)

	// Initialize context and teardown channel.
	p.handlers = channelHandlers{}
	cmds := make(chan Cmd)

	if p.Input == nil && !p.DisableInput {
		p.Input = os.Stdin
	}
	if p.Output == nil {
		p.Output = os.Stdout
	}
	if p.Env == nil {
		p.Env = os.Environ()
	}
	if p.IgnoreSignals {
		atomic.StoreUint32(&p.ignoreSignals, 1)
	}
	if p.FPS < 1 {
		p.FPS = defaultFPS
	} else if p.FPS > maxFPS {
		p.FPS = maxFPS
	}

	p.input = p.Input
	p.output = p.Output
	p.environ = uv.Environ(p.Env)

	p.finished = make(chan struct{})
	defer func() {
		close(p.finished)
	}()

	defer p.cancel()

	// Handle signals.
	if !p.DisableSignalHandler {
		p.handlers.add(p.handleSignals())
	}

	// Recover from panics.
	if !p.DisableCatchPanics {
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
		return p.InitialModel, err
	}

	// Get the initial window size.
	width, height := p.InitialWidth, p.InitialHeight
	if p.ttyOutput != nil {
		// Set the initial size of the terminal.
		w, h, err := term.GetSize(p.ttyOutput.Fd())
		if err != nil {
			return p.InitialModel, fmt.Errorf("bubbletea: error getting terminal size: %w", err)
		}

		width, height = w, h
	}

	p.width, p.height = width, height
	resizeMsg := WindowSizeMsg{Width: p.width, Height: p.height}

	if p.renderer == nil {
		if hasView(p.InitialModel) {
			// If no renderer is set use the cursed one.
			r := newCursedRenderer(
				p.output,
				p.environ,
				p.width,
				p.height,
			)
			r.setLogger(p.logger)
			r.setOptimizations(p.useHardTabs, p.useBackspace, p.ttyInput == nil)
			p.renderer = r
		}
	}

	// Get the color profile and send it to the program.
	if p.ColorProfile == nil {
		cp := colorprofile.Detect(p.output, p.environ)
		p.ColorProfile = &cp
	}

	p.profile = *p.ColorProfile

	// Set the color profile on the renderer and send it to the program.
	p.renderer.setColorProfile(p.profile)
	go p.Send(ColorProfileMsg{p.profile})

	// Send the initial size to the program.
	go p.Send(resizeMsg)
	p.renderer.resize(resizeMsg.Width, resizeMsg.Height)

	// Send the environment variables used by the program.
	go p.Send(EnvMsg(p.environ))

	// Init the input reader and initial model.
	model := p.InitialModel
	if p.input != nil {
		if err := p.initInputReader(false); err != nil {
			return model, err
		}
	}

	// Hide the cursor before starting the renderer. This is handled by the
	// renderer so we don't need to write the sequence here.
	p.renderer.hideCursor()

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
	var err error
	model, err = p.eventLoop(model, cmds)

	if err == nil && len(p.errs) > 0 {
		err = <-p.errs // Drain a leftover error in case eventLoop crashed.
	}

	killed := ctx.Err() != nil || p.ctx.Err() != nil || err != nil
	if killed {
		if err == nil && ctx.Err() != nil {
			// Return also as context error the cancellation of an external context.
			// This is the context the user knows about and should be able to act on.
			err = fmt.Errorf("%w: %w", ErrProgramKilled, ctx.Err())
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
	if !p.initialized.Load() {
		// Wait for the program to be initialized.
		time.Sleep(10 * time.Millisecond)
		p.Send(msg)
		return
	}
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
	_, _ = p.outputBuf.WriteString(seq)
}

// flush flushes the output buffer to the program output.
func (p *Program) flush() error {
	if p.outputBuf.Len() == 0 {
		return nil
	}
	if p.logger != nil {
		p.logger.Printf("output: %q", p.outputBuf.String())
	}
	_, err := p.output.Write(p.outputBuf.Bytes())
	p.outputBuf.Reset()
	if err != nil {
		return fmt.Errorf("error writing to output: %w", err)
	}
	return nil
}

// shutdown performs operations to free up resources and restore the terminal
// to its original state.
func (p *Program) shutdown(kill bool) {
	p.shutdownOnce.Do(func() {
		p.cancel()

		// Wait for all handlers to finish.
		p.handlers.shutdown()

		// Check if the cancel reader has been setup before waiting and closing.
		if p.cancelReader != nil {
			// Wait for input loop to finish.
			if p.cancelReader.Cancel() {
				if !kill {
					p.waitForReadLoop()
				}
			}
			_ = p.cancelReader.Close()
		}

		if p.renderer != nil {
			p.stopRenderer(kill)
		}

		_ = p.restoreTerminalState()
	})
}

// recoverFromPanic recovers from a panic, prints the stack trace, and restores
// the terminal to a usable state.
func (p *Program) recoverFromPanic(r interface{}) {
	select {
	case p.errs <- ErrProgramPanic:
	default:
	}
	p.shutdown(true) // Ok to call here, p.Run() cannot do it anymore.
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

// recoverFromGoPanic recovers from a goroutine panic, prints a stack trace and
// signals for the program to be killed and terminal restored to a usable state.
func (p *Program) recoverFromGoPanic(r interface{}) {
	select {
	case p.errs <- ErrProgramPanic:
	default:
	}
	p.cancel()
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
	if p.cancelReader != nil {
		p.cancelReader.Cancel()
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

	p.startRenderer()

	// If the output is a terminal, it may have been resized while another
	// process was at the foreground, in which case we may not have received
	// SIGWINCH. Detect any size change now and propagate the new size as
	// needed.
	go p.checkResize()

	// Flush queued commands.
	return p.flush()
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
	p.renderer.start()
	go func() {
		for {
			select {
			case <-p.rendererDone:
				p.ticker.Stop()
				return

			case <-p.ticker.C:
				_ = p.flush()
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
