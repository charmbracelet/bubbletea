package tea

import (
	"errors"
	"io"
	"os"
	"strings"
)

// Msg represents an action. It's used by Update to update the UI.
type Msg interface{}

// Model contains the updatable data for an application
type Model interface{}

// Cmd is an IO operation. If it's nil it's considered a no-op.
type Cmd func(Model) Msg

// Sub is an event subscription. If it returns nil it's considered a no-op.
type Sub func(Model) Msg

// Init is the first function that will be called. It returns your initial
// model and runs an optional command
type Init func() (Model, Cmd)

// Update is called when a message is received. It may update the model and/or
// send a command.
type Update func(Msg, Model) (Model, Cmd)

// View produces a string which will be rendered to the terminal
type View func(Model) string

// Program is a terminal user interface
type Program struct {
	init          Init
	update        Update
	view          View
	subscriptions []Sub
}

// ErrMsg is just a regular message containing an error. We handle it in Update
// just like a regular message by case switching. Of course, the developer
// could also define her own errors as well.
type ErrMsg struct {
	error
}

func (e ErrMsg) String() string {
	return e.Error()
}

// NewErrMsg is a convenience function for creating a generic ErrMsg
func NewErrMsg(s string) ErrMsg {
	return ErrMsg{errors.New(s)}
}

// NewErrMsgFromErr is a convenience function for creating an ErrMsg from an
// existing error
func NewErrMsgFromErr(e error) ErrMsg {
	return ErrMsg{e}
}

// Quit is a command that tells the program to exit
func Quit(_ Model) Msg {
	return quitMsg{}
}

// Signals that the program should quit
type quitMsg struct{}

// NewProgram creates a new Program
func NewProgram(init Init, update Update, view View, subs []Sub) *Program {
	return &Program{
		init:          init,
		update:        update,
		view:          view,
		subscriptions: subs,
	}
}

// Start initializes the program
func (p *Program) Start() error {
	var (
		model         Model
		cmd           Cmd
		cmds          = make(chan Cmd)
		msgs          = make(chan Msg)
		done          = make(chan struct{})
		linesRendered int
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

	// Render initial view
	linesRendered = p.render(model, linesRendered)

	// Subscribe to user input. We could move this out of here and offer it
	// as a subscription, but it blocks nicely and seems to be a common enough
	// need that we're enabling it by default.
	go func() {
		for {
			msg, _ := ReadKey(os.Stdin)
			msgs <- KeyMsg(msg)
		}
	}()

	// Process subscriptions
	go func() {
		if len(p.subscriptions) > 0 {
			for _, sub := range p.subscriptions {
				go func(s Sub) {
					for {
						msgs <- s(model)
					}
				}(sub)
			}
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
						msgs <- cmd(model)
					}()
				}
			}
		}
	}()

	// Handle updates and draw
	for {
		select {
		case msg := <-msgs:
			if _, ok := msg.(quitMsg); ok {
				close(done)
				return nil
			}

			model, cmd = p.update(msg, model)
			cmds <- cmd // process command (if any)
			linesRendered = p.render(model, linesRendered)
		}
	}
}

// Render a view to the terminal. Returns the number of lines rendered.
func (p *Program) render(model Model, linesRendered int) int {
	view := p.view(model) + "\n"

	// We need to add carriage returns to ensure that the cursor travels to the
	// start of a column after a newline
	view = strings.Replace(view, "\n", "\r\n", -1)

	if linesRendered > 0 {
		clearLines(linesRendered)
	}
	io.WriteString(os.Stdout, view)
	return strings.Count(view, "\r\n")
}
