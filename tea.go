package tea

import (
	"fmt"
	"os"
	"sync"

	te "github.com/muesli/termenv"
	"golang.org/x/crypto/ssh/terminal"
)

// Msg represents an action and is usually the result of an IO operation. It's
// triggers the Update function, and henceforth, the UI.
type Msg interface{}

// Model contains the program's state.
type Model interface{}

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

// Init is the first function that will be called. It returns your initial
// model and runs an optional command.
type Init func() (Model, Cmd)

// Update is called when a message is received. Use it to inspect messages and,
// in repsonse,  update the model and/or send a command.
type Update func(Msg, Model) (Model, Cmd)

// View renders the program's UI: a string which will be printed to the
// terminal. The view is rendered after every Update.
type View func(Model) string

// Program is a terminal user interface.
type Program struct {
	init   Init
	update Update
	view   View

	mtx    sync.Mutex
	output *os.File // where to send output. this will usually be os.Stdout.
}

// Quit is a special command that tells the program to exit.
func Quit() Msg {
	return quitMsg{}
}

// quitMsg in an internal message signals that the program should quit. You can
// send a quitMsg with Quit.
type quitMsg struct{}

// batchMsg is the internal message used to perform a bunch of commands. You
// can send a batchMsg with Batch.
type batchMsg []Cmd

// WindowSizeMsg is used to report on the terminal size. It's sent once
// initially and then on every terminal resize.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// NewProgram creates a new Program.
func NewProgram(init Init, update Update, view View) *Program {
	return &Program{
		init:   init,
		update: update,
		view:   view,

		output: os.Stdout,
	}
}

// Start initializes the program.
func (p *Program) Start() error {
	var (
		cmds       = make(chan Cmd)
		msgs       = make(chan Msg)
		errs       = make(chan error)
		done       = make(chan struct{})
		mrRenderer = newRenderer(p.output, &p.mtx)
	)

	err := initTerminal()
	if err != nil {
		return err
	}
	defer restoreTerminal()

	// Initialize program
	model, initCmd := p.init()
	if initCmd != nil {
		go func() {
			cmds <- initCmd
		}()
	}

	// Start renderer
	mrRenderer.start()

	// Render initial view
	mrRenderer.write(p.view(model))

	// Subscribe to user input
	go func() {
		for {
			msg, err := ReadInput(os.Stdin)
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

			// Handle quit message
			if _, ok := msg.(quitMsg); ok {
				mrRenderer.stop()
				close(done)
				return nil
			}

			// Process batch commands
			if batchedCmds, ok := msg.(batchMsg); ok {
				for _, cmd := range batchedCmds {
					cmds <- cmd
				}
				continue
			}

			// Process internal messages for the renderer
			mrRenderer.handleMessages(msg)
			var cmd Cmd
			model, cmd = p.update(msg, model) // run update
			cmds <- cmd                       // process command (if any)
			mrRenderer.write(p.view(model))   // send view to renderer
		}
	}
}

// EnterAltScreen enters the alternate screen buffer.
func (p *Program) EnterAltScreen() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.AltScreenSeq)
	moveCursor(p.output, 0, 0)
}

// ExitAltScreen exits the alternate screen buffer.
func (p *Program) ExitAltScreen() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.ExitAltScreenSeq)
}

// EnableMouseCellMotion enables mouse click, release, wheel and motion events if a
// button is pressed.
func (p *Program) EnableMouseCellMotion() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	fmt.Fprintf(p.output, te.CSI+te.EnableMouseCellMotionSeq)
}

// DisableMouseCellMotino disables Mouse Cell Motion tracking. If you've
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
