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
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/containerd/console"
	isatty "github.com/mattn/go-isatty"
	te "github.com/muesli/termenv"
	"golang.org/x/term"
)

// Msg represents an action and is usually the result of an IO operation. It
// triggers the Update function, and henceforth, the UI.
type Msg interface{}

// Model contains the program's state as well as it's core functions.
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

// Cmd is an IO operation. If it's nil it's considered a no-op. Use it for
// things like HTTP requests, timers, saving and loading from disk, and so on.
//
// There's almost never a need to use a command to send a message to another
// part of your program. Instead, it can almost always be done in the update
// function.
type Cmd func() Msg

// Options to customize the program during its initialization. These are
// generally set with ProgramOptions.
//
// The options here are treated as bits.
type startupOptions byte

func (s startupOptions) has(option startupOptions) bool {
	return s&option != 0
}

const (
	withAltScreen startupOptions = 1 << iota
	withMouseCellMotion
	withMouseAllMotion
	withInputTTY
	withCustomInput
)

// Program is a terminal user interface.
type Program struct {
	initialModel Model

	// Configuration options that will set as the program is initializing,
	// treated as bits. These options can be set via various ProgramOptions.
	startupOptions startupOptions

	mtx  *sync.Mutex
	done chan struct{}

	msgs chan Msg

	output          io.Writer // where to send output. this will usually be os.Stdout.
	input           io.Reader // this will usually be os.Stdin.
	renderer        renderer
	altScreenActive bool

	// CatchPanics is incredibly useful for restoring the terminal to a usable
	// state after a panic occurs. When this is set, Bubble Tea will recover
	// from panics, print the stack trace, and disable raw mode. This feature
	// is on by default.
	CatchPanics bool

	console console.Console

	// Stores the original reference to stdin for cases where input is not a
	// TTY on windows and we've automatically opened CONIN$ to receive input.
	// When the program exits this will be restored.
	//
	// Lint ignore note: the linter will find false positive on unix systems
	// as this value only comes into play on Windows, hence the ignore comment
	// below.
	windowsStdin *os.File //nolint:golint,structcheck,unused
}

// Batch performs a bunch of commands concurrently with no ordering guarantees
// about the results. Use a Batch to return several commands.
//
// Example:
//
//     func (m model) Init() Cmd {
//	       return tea.Batch(someCommand, someOtherCommand)
//     }
//
func Batch(cmds ...Cmd) Cmd {
	if len(cmds) == 0 {
		return nil
	}
	return func() Msg {
		return batchMsg(cmds)
	}
}

// batchMsg is the internal message used to perform a bunch of commands. You
// can send a batchMsg with Batch.
type batchMsg []Cmd

// Quit is a special command that tells the Bubble Tea program to exit.
func Quit() Msg {
	return quitMsg{}
}

// quitMsg in an internal message signals that the program should quit. You can
// send a quitMsg with Quit.
type quitMsg struct{}

// EnterAltScreen is a special command that tells the Bubble Tea program to
// enter alternate screen buffer.
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. To initialize your program with the altscreen enabled
// use the WithAltScreen ProgramOption instead.
func EnterAltScreen() Msg {
	return enterAltScreenMsg{}
}

// enterAltScreenMsg in an internal message signals that the program should
// enter alternate screen buffer. You can send a enterAltScreenMsg with
// EnterAltScreen.
type enterAltScreenMsg struct{}

// ExitAltScreen is a special command that tells the Bubble Tea program to exit
// the alternate screen buffer. This command should be used to exit the
// alternate screen buffer while the program is running.
//
// Note that the alternate screen buffer will be automatically exited when the
// program quits.
func ExitAltScreen() Msg {
	return exitAltScreenMsg{}
}

// exitAltScreenMsg in an internal message signals that the program should exit
// alternate screen buffer. You can send a exitAltScreenMsg with ExitAltScreen.
type exitAltScreenMsg struct{}

// EnableMouseCellMotion is a special command that enables mouse click,
// release, and wheel events. Mouse movement events are also captured if
// a mouse button is pressed (i.e., drag events).
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. Use the WithMouseCellMotion ProgramOption instead.
func EnableMouseCellMotion() Msg {
	return enableMouseCellMotionMsg{}
}

// enableMouseCellMotionMsg is a special command that signals to start
// listening for "cell motion" type mouse events (ESC[?1002l). To send an
// enableMouseCellMotionMsg, use the EnableMouseCellMotion command.
type enableMouseCellMotionMsg struct{}

// EnableMouseAllMotion is a special command that enables mouse click, release,
// wheel, and motion events, which are delivered regardless of whether a mouse
// button is pressed, effectively enabling support for hover interactions.
//
// Many modern terminals support this, but not all. If in doubt, use
// EnableMouseCellMotion instead.
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. Use the WithMouseAllMotion ProgramOption instead.
func EnableMouseAllMotion() Msg {
	return enableMouseAllMotionMsg{}
}

// enableMouseAllMotionMsg is a special command that signals to start listening
// for "all motion" type mouse events (ESC[?1003l). To send an
// enableMouseAllMotionMsg, use the EnableMouseAllMotion command.
type enableMouseAllMotionMsg struct{}

// DisableMouse is a special command that stops listening for mouse events.
func DisableMouse() Msg {
	return disableMouseMsg{}
}

// disableMouseMsg is an internal message that that signals to stop listening
// for mouse events. To send a disableMouseMsg, use the DisableMouse command.
type disableMouseMsg struct{}

// WindowSizeMsg is used to report the terminal size. It's sent to Update once
// initially and then on every terminal resize. Note that Windows does not
// have support for reporting when resizes occur as it does not support the
// SIGWINCH signal.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// HideCursor is a special command for manually instructing Bubble Tea to hide
// the cursor. In some rare cases, certain operations will cause the terminal
// to show the cursor, which is normally hidden for the duration of a Bubble
// Tea program's lifetime. You will most likely not need to use this command.
func HideCursor() Msg {
	return hideCursorMsg{}
}

// hideCursorMsg is an internal command used to hide the cursor. You can send
// this message with HideCursor.
type hideCursorMsg struct{}

// NewProgram creates a new Program.
func NewProgram(model Model, opts ...ProgramOption) *Program {
	p := &Program{
		mtx:          &sync.Mutex{},
		initialModel: model,
		output:       os.Stdout,
		input:        os.Stdin,
		CatchPanics:  true,
	}

	// Apply all options to program
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Start initializes the program.
func (p *Program) Start() error {
	p.msgs = make(chan Msg)
	p.done = make(chan struct{})

	var (
		cmds = make(chan Cmd)
		errs = make(chan error)
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	switch {
	case p.startupOptions.has(withInputTTY):
		// Open a new TTY, by request
		f, err := openInputTTY()
		if err != nil {
			return err
		}
		p.input = f

	case !p.startupOptions.has(withCustomInput):
		// If the user hasn't set a custom input, and input's not a terminal,
		// open a TTY so we can capture input as normal. This will allow things
		// to "just work" in cases where data was piped or redirected into this
		// application.
		f, isFile := p.input.(*os.File)
		if !isFile {
			break
		}

		if isatty.IsTerminal(f.Fd()) {
			break
		}

		f, err := openInputTTY()
		if err != nil {
			return err
		}
		p.input = f
	}

	// Listen for SIGINT. Note that in most cases ^C will not send an
	// interrupt because the terminal will be in raw mode and thus capture
	// that keystroke and send it along to Program.Update. If input is not a
	// TTY, however, ^C will be caught here.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT)
		defer signal.Stop(sig)

		select {
		case <-ctx.Done():
		case <-sig:
			p.msgs <- quitMsg{}
		}
	}()

	if p.CatchPanics {
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
		return err
	}

	// If no renderer is set use the standard one.
	if p.renderer == nil {
		p.renderer = newRenderer(p.output, p.mtx)
	}

	// Honor program startup options.
	if p.startupOptions&withAltScreen != 0 {
		p.EnterAltScreen()
	}
	if p.startupOptions&withMouseCellMotion != 0 {
		p.EnableMouseCellMotion()
	} else if p.startupOptions&withMouseAllMotion != 0 {
		p.EnableMouseAllMotion()
	}

	// Initialize program
	model := p.initialModel
	if initCmd := model.Init(); initCmd != nil {
		go func() {
			cmds <- initCmd
		}()
	}

	// Start renderer
	p.renderer.start()
	p.renderer.setAltScreen(p.altScreenActive)

	// Render initial view
	p.renderer.write(model.View())

	// Subscribe to user input
	if p.input != nil {
		go func() {
			for {
				msg, err := readInput(p.input)
				if err != nil {
					// If we get EOF just stop listening for input
					if err == io.EOF {
						break
					}
					errs <- err
				}
				p.msgs <- msg
			}
		}()
	}

	if f, ok := p.output.(*os.File); ok {
		// Get initial terminal size and send it to the program
		go func() {
			w, h, err := term.GetSize(int(f.Fd()))
			if err != nil {
				errs <- err
			}
			p.msgs <- WindowSizeMsg{w, h}
		}()

		// Listen for window resizes
		go listenForResize(f, p.msgs, errs)
	}

	// Process commands
	go func() {
		for {
			select {
			case <-p.done:
				return
			case cmd := <-cmds:
				if cmd != nil {
					go func() {
						p.msgs <- cmd()
					}()
				}
			}
		}
	}()

	// Handle updates and draw
	for {
		select {
		case err := <-errs:
			p.shutdown(false)
			return err
		case msg := <-p.msgs:

			// Handle special internal messages
			switch msg := msg.(type) {
			case quitMsg:
				p.shutdown(false)
				return nil

			case batchMsg:
				for _, cmd := range msg {
					cmds <- cmd
				}
				continue

			case WindowSizeMsg:
				p.renderer.repaint()

			case enterAltScreenMsg:
				p.EnterAltScreen()

			case exitAltScreenMsg:
				p.ExitAltScreen()

			case enableMouseCellMotionMsg:
				p.EnableMouseCellMotion()

			case enableMouseAllMotionMsg:
				p.EnableMouseAllMotion()

			case disableMouseMsg:
				p.DisableMouseCellMotion()
				p.DisableMouseAllMotion()

			case hideCursorMsg:
				hideCursor(p.output)
			}

			// Process internal messages for the renderer
			if r, ok := p.renderer.(*standardRenderer); ok {
				r.handleMessages(msg)
			}

			var cmd Cmd
			model, cmd = model.Update(msg) // run update
			cmds <- cmd                    // process command (if any)
			p.renderer.write(model.View()) // send view to renderer
		}
	}
}

// Send sends a message to the main update function, effectively allowing
// messages to be injected from outside the program for interoperability
// purposes.
//
// If the program is not running this this will be a no-op, so it's safe to
// send messages if the program is unstarted, or has exited.
//
// This method is currently provisional. The method signature may alter
// slightly, or it may be removed in a future version of this package.
func (p *Program) Send(msg Msg) {
	if p.msgs != nil {
		p.msgs <- msg
	}
}

// shutdown performs operations to free up resources and restore the terminal
// to its original state.
func (p *Program) shutdown(kill bool) {
	if kill {
		p.renderer.kill()
	} else {
		p.renderer.stop()
	}
	close(p.done)
	p.ExitAltScreen()
	p.DisableMouseCellMotion()
	p.DisableMouseAllMotion()
	_ = p.restoreTerminal()
}

// EnterAltScreen enters the alternate screen buffer, which consumes the entire
// terminal window. ExitAltScreen will return the terminal to its former state.
//
// Deprecated. Use the WithAltScreen ProgramOption instead.
func (p *Program) EnterAltScreen() {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.altScreenActive {
		return
	}

	fmt.Fprintf(p.output, te.CSI+te.AltScreenSeq)
	moveCursor(p.output, 0, 0)

	p.altScreenActive = true
	if p.renderer != nil {
		p.renderer.setAltScreen(p.altScreenActive)
	}
}

// ExitAltScreen exits the alternate screen buffer.
//
// Deprecated. The altscreen will exited automatically when the program exits.
func (p *Program) ExitAltScreen() {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if !p.altScreenActive {
		return
	}

	fmt.Fprintf(p.output, te.CSI+te.ExitAltScreenSeq)

	p.altScreenActive = false
	if p.renderer != nil {
		p.renderer.setAltScreen(p.altScreenActive)
	}
}

// EnableMouseCellMotion enables mouse click, release, wheel and motion events
// if a mouse button is pressed (i.e., drag events).
//
// Deprecated. Use the WithMouseCellMotion ProgramOption instead.
func (p *Program) EnableMouseCellMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.EnableMouseCellMotionSeq)
}

// DisableMouseCellMotion disables Mouse Cell Motion tracking. This will be
// called automatically when exiting a Bubble Tea program.
//
// Deprecated. The mouse will automatically be disabled when the program exits.
func (p *Program) DisableMouseCellMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.DisableMouseCellMotionSeq)
}

// EnableMouseAllMotion enables mouse click, release, wheel and motion events,
// regardless of whether a mouse button is pressed. Many modern terminals
// support this, but not all.
//
// Deprecated. Use the WithMouseAllMotion ProgramOption instead.
func (p *Program) EnableMouseAllMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.EnableMouseAllMotionSeq)
}

// DisableMouseAllMotion disables All Motion mouse tracking. This will be
// called automatically when exiting a Bubble Tea program.
//
// Deprecated. The mouse will automatically be disabled when the program exits.
func (p *Program) DisableMouseAllMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.DisableMouseAllMotionSeq)
}
