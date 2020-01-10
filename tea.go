package tea

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/term"
)

// Escape sequence
const esc = "\033["

// Msg represents an action. It's used by Update to update the UI.
type Msg interface{}

// Model contains the data for an application
type Model interface{}

// Cmd is an IO operation. If it's nil it's considered a no-op.
type Cmd func() Msg

// Update is called when a message is received. It may update the model and/or
// send a command.
type Update func(Msg, Model) (Model, Cmd)

// View produces a string which will be rendered to the terminal
type View func(Model) string

// Program is a terminal user interface
type Program struct {
	model  Model
	update Update
	view   View
	rw     io.ReadWriter
}

// Quit command
func Quit() Msg {
	return quitMsg{}
}

// Signals that the program should quit
type quitMsg struct{}

// NewProgram creates a new Program
func NewProgram(model Model, update Update, view View) *Program {
	return &Program{
		model:  model,
		update: update,
		view:   view,
		// TODO: subscriptions
	}
}

// Start initializes the program
// TODO: error channel
func (p *Program) Start() error {
	var (
		model = p.model
		cmd   Cmd
		cmds  = make(chan Cmd)
		msgs  = make(chan Msg)
		done  = make(chan struct{})
	)

	tty, err := term.Open("/dev/tty")
	if err != nil {
		return err
	}

	p.rw = tty
	tty.SetRaw()
	defer func() {
		showCursor()
		tty.Restore()
	}()

	// Render initial view
	hideCursor()
	p.render(model, true)

	// Subscribe to user input
	// TODO: move to program struct to allow for subscriptions to other things,
	// too, like timers, frames, download/upload progress and so on.
	go func() {
		for {
			select {
			default:
				msg, _ := ReadKey(p.rw)
				msgs <- KeyPressMsg(msg)
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
						msgs <- cmd()
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
			p.render(model, false)
			p.model = model
		}
	}
}

// Render a view
func (p *Program) render(model Model, init bool) {
	view := p.view(model)

	// We need to add carriage returns to ensure that the cursor travels to the
	// start of a column after a newline
	view = strings.Replace(view, "\n", "\r\n", -1)

	if !init {
		clearLines(strings.Count(view, "\r\n"))
	}
	io.WriteString(p.rw, view)
}

func hideCursor() {
	fmt.Printf(esc + "?25l")
}

func showCursor() {
	fmt.Printf(esc + "?25h")
}

func cursorUp(n int) {
	fmt.Printf(esc+"%dF", n)
}

func clearLine() {
	fmt.Printf(esc + "2K")
}

func clearLines(n int) {
	for i := 0; i < n; i++ {
		cursorUp(1)
		clearLine()
	}
}

func clearScreen() {
	fmt.Printf(esc + "2J" + esc + "3J" + esc + "1;1H")
}
