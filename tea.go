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
	"fmt"
	"os"
	"runtime/debug"
	"sync"

	te "github.com/muesli/termenv"
	"golang.org/x/crypto/ssh/terminal"
)

// Msg represents an action and is usually the result of an IO operation. It's
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

// Batch peforms a bunch of commands concurrently with no ordering guarantees
// about the results.
func Batch(cmds ...Cmd) Cmd {
	if len(cmds) == 0 {
		return nil
	}
	return func() Msg {
		return batchMsg(cmds)
	}
}

// Program is a terminal user interface.
type Program struct {
	initialModel Model

	mtx             sync.Mutex
	output          *os.File // where to send output. this will usually be os.Stdout.
	renderer        *renderer
	altScreenActive bool

	// CatchPanics is incredibly useful for restoring the terminal to a useable
	// state after a panic occurs. When this is set, Bubble Tea will recover
	// from panics, print the stack trace, and disable raw mode. This feature
	// is on by default.
	CatchPanics bool
}

// Quit is a special command that tells the Bubble Tea program to exit.
func Quit() Msg {
	return quitMsg{}
}

// quitMsg in an internal message signals that the program should quit. You can
// send a quitMsg with Quit.
type quitMsg struct{}

// batchMsg is the internal message used to perform a bunch of commands. You
// can send a batchMsg with Batch.
type batchMsg []Cmd

// WindowSizeMsg is used to report on the terminal size. It's sent to Update
// once initially and then on every terminal resize.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// HideCursor is a special command for manually instructing Bubble Tea to hide
// the cursor. In some rare cases, certain operations will cause the terminal
// to show the cursor, which is normally hidden for the duration of a Bubble
// Tea program's lifetime. You most likely will not need to use this command.
func HideCursor() Msg {
	return hideCursorMsg{}
}

// hideCursorMsg is an internal command used to hide the cursor. You can send
// this message with HideCursor.
type hideCursorMsg struct{}

// NewProgram creates a new Program.
func NewProgram(model Model) *Program {
	return &Program{
		initialModel: model,
		output:       os.Stdout,
		CatchPanics:  true,
	}
}

// Start initializes the program.
func (p *Program) Start() error {
	var (
		cmds = make(chan Cmd)
		msgs = make(chan Msg)
		errs = make(chan error)
		done = make(chan struct{})
	)

	if p.CatchPanics {
		defer func() {
			if r := recover(); r != nil {
				p.ExitAltScreen()
				fmt.Printf("Caught panic:\n\n%s\n\nRestoring terminal...\n\n", r)
				debug.PrintStack()
				return
			}
		}()
	}

	p.renderer = newRenderer(p.output, &p.mtx)

	err := initTerminal(p.output)
	if err != nil {
		return err
	}
	defer restoreTerminal(p.output)

	// Initialize program
	model := p.initialModel
	initCmd := model.Init()
	if initCmd != nil {
		go func() {
			cmds <- initCmd
		}()
	}

	// Start renderer
	p.renderer.start()
	p.renderer.altScreenActive = p.altScreenActive

	// Render initial view
	p.renderer.write(model.View())

	// Subscribe to user input
	go func() {
		for {
			msg, err := readInput(os.Stdin)
			if err != nil {
				errs <- err
			}
			msgs <- msg
		}
	}()

	// Get initial terminal size
	go func() {
		w, h, err := terminal.GetSize(int(p.output.Fd()))
		if err != nil {
			errs <- err
		}
		msgs <- WindowSizeMsg{w, h}
	}()

	// Listen for window resizes
	go listenForResize(p.output, msgs, errs)

	// Process commands
	go func() {
		for {
			select {
			case <-done:
				return
			case cmd := <-cmds:
				if cmd != nil {
					go func() {
						msgs <- cmd()
					}()
				}
			}
		}
	}()

	// Handle updates and draw
	for {
		select {
		case err := <-errs:
			close(done)
			return err
		case msg := <-msgs:

			// Handle special messages
			switch msg.(type) {
			case quitMsg:
				p.renderer.stop()
				close(done)
				return nil
			case hideCursorMsg:
				hideCursor(p.output)
			}

			// Process batch commands
			if batchedCmds, ok := msg.(batchMsg); ok {
				for _, cmd := range batchedCmds {
					cmds <- cmd
				}
				continue
			}

			// Process internal messages for the renderer
			p.renderer.handleMessages(msg)
			var cmd Cmd
			model, cmd = model.Update(msg) // run update
			cmds <- cmd                    // process command (if any)
			p.renderer.write(model.View()) // send view to renderer
		}
	}
}

// EnterAltScreen enters the alternate screen buffer, which consumes the entire
// terminal window. ExitAltScreen will return the terminal to its former state.
func (p *Program) EnterAltScreen() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.AltScreenSeq)
	moveCursor(p.output, 0, 0)

	p.altScreenActive = true
	if p.renderer != nil {
		p.renderer.altScreenActive = p.altScreenActive
	}
}

// ExitAltScreen exits the alternate screen buffer.
func (p *Program) ExitAltScreen() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.ExitAltScreenSeq)

	p.altScreenActive = false
	if p.renderer != nil {
		p.renderer.altScreenActive = p.altScreenActive
	}
}

// EnableMouseCellMotion enables mouse click, release, wheel and motion events
// if a button is pressed.
func (p *Program) EnableMouseCellMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.EnableMouseCellMotionSeq)
}

// DisableMouseCellMotion disables Mouse Cell Motion tracking. If you've
// enabled Cell Motion mouse trakcing be sure to call this as your program is
// exiting or your users will be very upset!
func (p *Program) DisableMouseCellMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.DisableMouseCellMotionSeq)
}

// EnableMouseAllMotion enables mouse click, release, wheel and motion events,
// regardless of whether a button is pressed. Many modern terminals support
// this, but not all.
func (p *Program) EnableMouseAllMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.EnableMouseAllMotionSeq)
}

// DisableMouseAllMotion disables All Motion mouse tracking. If you've enabled
// All Motion mouse tracking be sure you call this as your program is exiting
// or your users will be very upset!
func (p *Program) DisableMouseAllMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.DisableMouseAllMotionSeq)
}
