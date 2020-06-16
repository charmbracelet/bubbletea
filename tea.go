package tea

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/muesli/termenv"
	"golang.org/x/crypto/ssh/terminal"
)

// Msg represents an action and is usually the result of an IO operation. It's
// triggers the  Update function, and henceforth, the UI.
type Msg interface{}

// Model contains the program's state.
type Model interface{}

// Cmd is an IO operation. If it's nil it's considered a no-op.
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

// Update is called when a message is received. It may update the model and/or
// send a command.
type Update func(Msg, Model) (Model, Cmd)

// View produces a string which will be rendered to the terminal.
type View func(Model) string

// Program is a terminal user interface.
type Program struct {
	init   Init
	update Update
	view   View
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

// WindowSizeMsg is used to report on the terminal size. It's fired once initially
// and then on every terminal resize.
type WindowSizeMsg struct {
	width  int
	height int
}

// NewProgram creates a new Program.
func NewProgram(init Init, update Update, view View) *Program {
	return &Program{
		init:   init,
		update: update,
		view:   view,
	}
}

// Start initializes the program.
func (p *Program) Start() error {
	var (
		model Model
		cmd   Cmd
		cmds  = make(chan Cmd)
		msgs  = make(chan Msg)
		errs  = make(chan error)
		done  = make(chan struct{})

		output     *os.File = os.Stdout
		mrRenderer          = newRenderer(output)
	)

	err := initTerminal()
	if err != nil {
		return err
	}
	defer restoreTerminal()

	// Initialize program
	model, cmd = p.init()
	if cmd != nil {
		go func() {
			cmds <- cmd
		}()
	}

	// Start renderer
	mrRenderer.start()

	// Render initial view
	mrRenderer.write(p.view(model))

	// Subscribe to user input
	go func() {
		for {
			msg, err := ReadKey(os.Stdin)
			if err != nil {
				errs <- err
			}
			msgs <- KeyMsg(msg)
		}
	}()

	// Get initial terminal size
	go func() {
		w, h, err := terminal.GetSize(int(output.Fd()))
		if err != nil {
			errs <- err
		}
		msgs <- WindowSizeMsg{w, h}
	}()

	// Listen for window resizes
	go func() {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGWINCH)
		for {
			<-sig
			w, h, err := terminal.GetSize(int(output.Fd()))
			if err != nil {
				errs <- err
			}
			msgs <- WindowSizeMsg{w, h}
		}
	}()

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

			// Report resizes to the renderer. This only matters if we're doing
			// higher performance scroll-based rendering.
			if size, ok := msg.(WindowSizeMsg); ok {
				mrRenderer.width = size.width
				mrRenderer.height = size.height
			}

			// Handle messages telling the renderer to ignore ranges of lines
			if ignore, ok := msg.(IgnoreLinesMsg); ok {
				mrRenderer.setIgnoredLines(ignore.from, ignore.to)
			}

			// Handle messages telling the renderer to stop ignoring lines
			if _, ok := msg.(IgnoreLinesMsg); ok {
				mrRenderer.clearIgnoredLines()
			}

			// Process batch commands
			if batchedCmds, ok := msg.(batchMsg); ok {
				for _, cmd := range batchedCmds {
					cmds <- cmd
				}
				continue
			}

			model, cmd = p.update(msg, model) // run update
			cmds <- cmd                       // process command (if any)
			mrRenderer.write(p.view(model))   // send view to renderer
		}
	}
}

// AltScreen exits the altscreen. This is just a wrapper around the termenv
// function.
func AltScreen() {
	termenv.AltScreen()
}

// ExitAltScreen exits the altscreen. This is just a wrapper around the termenv
// function.
func ExitAltScreen() {
	termenv.ExitAltScreen()
}
