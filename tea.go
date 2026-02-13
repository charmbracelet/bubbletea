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

	// View renders the program's UI, which can be a string or a [Layer]. The
	// view is rendered after every Update.
	View() View
}

// NewView is a helper function to create a new [View] with the given styled
// string. A styled string represents text with styles and hyperlinks encoded
// as ANSI escape codes.
//
// Example:
//
//	```go
//	v := tea.NewView("Hello, World!")
//	```
func NewView(s string) View {
	var view View
	view.SetContent(s)
	return view
}

// View represents a terminal view that can be composed of multiple layers.
// It can also contain a cursor that will be rendered on top of the layers.
type View struct {
	// Content is the screen content of the view. It holds styled strings that
	// will be rendered to the terminal when the view is rendered.
	//
	// A styled string represents text with styles and hyperlinks encoded as
	// ANSI escape codes.
	//
	// Example:
	//
	//  ```go
	//  v := tea.NewView("Hello, World!")
	//  ```
	Content string

	// OnMouse is an optional mouse message handler that can be used to
	// intercept mouse messages that depends on view content from last render.
	// It can be useful for implementing view-specific behavior without
	// breaking the unidirectional data flow of Bubble Tea.
	//
	// Example:
	//
	//  ```go
	//  content := "Hello, World!"
	//  v := tea.NewView(content)
	//  v.OnMouse = func(msg tea.MouseMsg) tea.Cmd {
	//      return func() tea.Msg {
	//        m := msg.Mouse()
	//        // Check if the mouse is within the bounds of "World!"
	//        start := strings.Index(content, "World!")
	//        end := start + len("World!")
	//        if m.Y == 0 && m.X >= start && m.X < end {
	//          // Mouse is over "World!"
	//          return MyCustomMsg{
	//            MouseMsg: msg,
	//          }
	//		  }
	//      }
	//    }
	//    return nil
	//  }
	//  return v
	//  ```
	OnMouse func(msg MouseMsg) Cmd

	// Cursor represents the cursor position, style, and visibility on the
	// screen. When not nil, the cursor will be shown at the specified
	// position.
	Cursor *Cursor

	// BackgroundColor when not nil, sets the terminal background color. Use
	// nil to reset to the terminal's default background color.
	BackgroundColor color.Color

	// ForegroundColor when not nil, sets the terminal foreground color. Use
	// nil to reset to the terminal's default foreground color.
	ForegroundColor color.Color

	// WindowTitle sets the terminal window title. Support depends on the
	// terminal.
	WindowTitle string

	// ProgressBar when not nil, shows a progress bar in the terminal's
	// progress bar section. Support depends on the terminal.
	ProgressBar *ProgressBar

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

	// KeyboardEnhancements describes what keyboard enhancement features Bubble
	// Tea should request from the terminal.
	//
	// Bubble Tea supports requesting the following keyboard enhancement features:
	//   - ReportEventTypes: requests the terminal to report key repeat and
	//     release events.
	//
	// If the terminal supports any of these features, your program will
	// receive  a [KeyboardEnhancementsMsg] that indicates which features are
	// available.
	KeyboardEnhancements KeyboardEnhancements
}

// KeyboardEnhancements describes the requested keyboard enhancement features.
// If the terminal supports any of them, it will respond with a
// [KeyboardEnhancementsMsg] that indicates which features are supported.

// KeyboardEnhancements defines different keyboard enhancement features that
// can be requested from the terminal.

// KeyboardEnhancements defines different keyboard enhancement features that
// can be requested from the terminal.
//
// By default, Bubble Tea requests basic key disambiguation features from the
// terminal. If the terminal supports keyboard enhancements, or any of its
// additional features, it will respond with a [KeyboardEnhancementsMsg] that
// indicates which features are supported.
//
// Example:
//
//	```go
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	  switch msg := msg.(type) {
//	  case tea.KeyboardEnhancementsMsg:
//	    // We have basic key disambiguation support.
//	    // We can handle "shift+enter", "ctrl+i", etc.
//		m.keyboardEnhancements = msg
//		if msg.ReportEventTypes {
//		  // Even better! We can now handle key repeat and release events.
//		}
//	  case tea.KeyPressMsg:
//	    switch msg.String() {
//	    case "shift+enter":
//	      // Handle shift+enter
//	      // This would not be possible without keyboard enhancements.
//	    case "ctrl+j":
//	      // Handle ctrl+j
//	    }
//	  case tea.KeyReleaseMsg:
//	    // Whoa! A key was released!
//	  }
//
//	  return m, nil
//	}
//
//	func (m model) View() tea.View {
//	  v := tea.NewView("Press some keys!")
//	  // Request reporting key repeat and release events.
//	  v.KeyboardEnhancements.ReportEventTypes = true
//	  return v
//	}
//	```
type KeyboardEnhancements struct {
	// ReportEventTypes requests the terminal to report key repeat and release
	// events.
	// If supported, your program will receive [KeyReleaseMsg]s and
	// [KeyPressMsg] with the [Key.IsRepeat] field set indicating that this is
	// a it's part of a key repeat sequence.
	ReportEventTypes bool
}

// SetContent is a helper method to set the content of a [View] with a styled
// string. A styled string represents text with styles and hyperlinks encoded
// as ANSI escape codes.
//
// Example:
//
//	```go
//	var v tea.View
//	v.SetContent("Hello, World!")
//	```
func (v *View) SetContent(s string) {
	v.Content = s
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
	// disableInput disables all input. This is useful for programs that
	// don't need input, like a progress bar or a spinner.
	disableInput bool

	// disableSignalHandler disables the signal handler that Bubble Tea sets up
	// for Programs. This is useful if you want to handle signals yourself.
	disableSignalHandler bool

	// disableCatchPanics disables the panic catching that Bubble Tea does by
	// default. If panic catching is disabled the terminal will be in a fairly
	// unusable state after a panic because Bubble Tea will not perform its usual
	// cleanup on exit.
	disableCatchPanics bool

	// filter supplies an event filter that will be invoked before Bubble Tea
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
	//	p.filter = filter
	//
	//	if _,err := p.Run(context.Background()); err != nil {
	//		fmt.Println("Error running program:", err)
	//		os.Exit(1)
	//	}
	filter func(Model, Msg) Msg

	// fps sets a custom maximum fps at which the renderer should run. If less
	// than 1, the default value of 60 will be used. If over 120, the fps will
	// be capped at 120.
	fps int

	// initialModel is the initial model for the program and is the only
	// required field when creating a new program.
	initialModel Model

	// disableRenderer prevents the program from rendering to the terminal.
	// This can be useful for running daemon-like programs that don't require a
	// UI but still want to take advantage of Bubble Tea's architecture.
	disableRenderer bool

	// handlers is a list of channels that need to be waited on before the
	// program can exit.
	handlers channelHandlers

	// ctx is the programs's internal context for signalling internal teardown.
	// It is built and derived from the externalCtx in NewProgram().
	ctx    context.Context
	cancel context.CancelFunc

	// externalCtx is a context that was passed in via WithContext, otherwise defaulting
	// to ctx.Background() (in case it was not), the internal context is derived from it.
	externalCtx context.Context

	msgs         chan Msg
	errs         chan error
	finished     chan struct{}
	shutdownOnce sync.Once

	profile *colorprofile.Profile // the terminal color profile

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

	mu sync.Mutex
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
func NewProgram(model Model, opts ...ProgramOption) *Program {
	p := &Program{
		initialModel: model,
		msgs:         make(chan Msg),
		errs:         make(chan error, 1),
		rendererDone: make(chan struct{}),
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
					if !p.disableCatchPanics {
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
				switch msg.Content {
				case "RGB", "Tc":
					if *p.profile != colorprofile.TrueColor {
						tc := colorprofile.TrueColor
						p.profile = &tc
						go p.Send(ColorProfileMsg{*p.profile})
					}
				}

			case ModeReportMsg:
				switch msg.Mode {
				case ansi.ModeSynchronizedOutput:
					if msg.Value == ansi.ModeReset {
						// The terminal supports synchronized output and it's
						// currently disabled, so we can enable it on the renderer.
						p.renderer.setSyncdUpdates(true)
					}
				case ansi.ModeUnicodeCore:
					if msg.Value == ansi.ModeReset || msg.Value == ansi.ModeSet || msg.Value == ansi.ModePermanentlySet {
						p.renderer.setWidthMethod(ansi.GraphemeWidth)
					}
				}

			case MouseMsg:
				switch msg.(type) {
				case MouseClickMsg, MouseReleaseMsg, MouseWheelMsg, MouseMotionMsg:
					// Only send mouse messages to the renderer if they are an
					// actual mouse event.
					if cmd := p.renderer.onMouse(msg); cmd != nil {
						go p.Send(cmd())
					}
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

// render renders the given view to the renderer.
func (p *Program) render(model Model) {
	if p.renderer != nil {
		p.renderer.render(model.View()) // send view to renderer
	}
}

func (p *Program) execSequenceMsg(msg sequenceMsg) {
	if !p.disableCatchPanics {
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
	if !p.disableCatchPanics {
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

			if !p.disableCatchPanics {
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

// shouldQuerySynchronizedOutput determines whether the terminal should be
// queried for various capabilities.
//
// This function checks for terminals that are known to support mode 2026,
// while excluding SSH sessions which may be unreliable, unless it's a
// known-good terminal like Windows Terminal.
//
// The function returns true for:
//   - Terminals without TERM_PROGRAM set and not in SSH sessions
//   - Windows Terminal (WT_SESSION is set)
//   - Terminals with TERM_PROGRAM set (except Apple Terminal) and not in SSH sessions
//   - Specific terminal types: ghostty, wezterm, alacritty, kitty, rio
func shouldQuerySynchronizedOutput(environ uv.Environ) bool {
	termType := environ.Getenv("TERM")
	termProg, okTermProg := environ.LookupEnv("TERM_PROGRAM")
	_, okSSHTTY := environ.LookupEnv("SSH_TTY")
	_, okWTSession := environ.LookupEnv("WT_SESSION")

	return (!okTermProg && !okSSHTTY) ||
		okWTSession ||
		(okTermProg && !strings.Contains(termProg, "Apple") && !okSSHTTY) ||
		strings.Contains(termType, "ghostty") ||
		strings.Contains(termType, "wezterm") ||
		strings.Contains(termType, "alacritty") ||
		strings.Contains(termType, "kitty") ||
		strings.Contains(termType, "rio")
}

// Run initializes the program and runs its event loops, blocking until it gets
// terminated by either [Program.Quit], [Program.Kill], or its signal handler.
// Returns the final model.
func (p *Program) Run() (returnModel Model, returnErr error) {
	if p.initialModel == nil {
		return nil, errors.New("bubbletea: InitialModel cannot be nil")
	}

	// Initialize context and teardown channel.
	p.handlers = channelHandlers{}
	cmds := make(chan Cmd)

	p.finished = make(chan struct{})
	defer func() {
		close(p.finished)
	}()

	defer p.cancel()

	if p.disableInput {
		p.input = nil
	} else if p.input == nil {
		// Always open the TTY for input.
		ttyIn, _, err := OpenTTY()
		if err != nil {
			return p.initialModel, fmt.Errorf("bubbletea: error opening TTY: %w", err)
		}
		p.input = ttyIn
	}

	// Handle signals.
	if !p.disableSignalHandler {
		p.handlers.add(p.handleSignals())
	}

	// Recover from panics.
	if !p.disableCatchPanics {
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
	width, height := p.width, p.height
	if p.ttyOutput != nil {
		// Set the initial size of the terminal.
		w, h, err := term.GetSize(p.ttyOutput.Fd())
		if err != nil {
			return p.initialModel, fmt.Errorf("bubbletea: error getting terminal size: %w", err)
		}

		width, height = w, h
	}

	p.width, p.height = width, height
	resizeMsg := WindowSizeMsg{Width: p.width, Height: p.height}

	if p.renderer == nil {
		if p.disableRenderer {
			p.renderer = &nilRenderer{}
		} else {
			// If no renderer is set use the cursed one.
			r := newCursedRenderer(
				p.output,
				p.environ,
				p.width,
				p.height,
			)
			r.setLogger(p.logger)
			// XXX: This breaks many things especially when we want the output
			// to be compatible with terminals that are not necessary a TTY.
			// This was originally done to work around a Wish emulated-pty
			// issue where when a PTY session is detected, and we don't
			// allocate a real PTY, the terminal settings (Termios and WinCon)
			// don't change and the we end up working in cooked mode instead of
			// raw mode.
			mapNl := false // p.ttyInput == nil
			r.setOptimizations(p.useHardTabs, p.useBackspace, mapNl)
			p.renderer = r
		}
	}

	// Get the color profile and send it to the program.
	if p.profile == nil {
		cp := colorprofile.Detect(p.output, p.environ)
		p.profile = &cp
	}

	// Set the color profile on the renderer and send it to the program.
	p.renderer.setColorProfile(*p.profile)
	go p.Send(ColorProfileMsg{*p.profile})

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

	// Start the renderer.
	p.startRenderer()

	if shouldQuerySynchronizedOutput(p.environ) {
		// Query for synchronized updates support (mode 2026) and unicode core
		// (mode 2027). If the terminal supports it, the renderer will enable
		// it once we get the response.
		p.execute(ansi.RequestModeSynchronizedOutput +
			ansi.RequestModeUnicodeCore)
	}

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
	p.mu.Lock()
	_, _ = p.outputBuf.WriteString(seq)
	p.mu.Unlock()
}

// flush flushes the output buffer to the program output.
func (p *Program) flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()

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
	p.renderer.start()
	go func() {
		for {
			select {
			case <-p.rendererDone:
				p.ticker.Stop()
				return

			case <-p.ticker.C:
				_ = p.flush()
				_ = p.renderer.flush(false)
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
		_ = p.renderer.flush(true)
	}

	_ = p.renderer.close()
}
