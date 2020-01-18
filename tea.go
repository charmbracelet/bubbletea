package tea

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"strings"

	"github.com/pkg/term"
)

// Escape sequence
const esc = "\033["

// The number of lines we last rendered
var linesRendered = 0

// Msg represents an action. It's used by Update to update the UI.
type Msg interface{}

// Model contains the updatable data for an application
type Model interface{}

// Cmd is an IO operation. If it's nil it's considered a no-op.
type Cmd func() Msg

// Sub is an event subscription. If it returns nil it's considered a no-op.
type Sub func(Model) Msg

// Update is called when a message is received. It may update the model and/or
// send a command.
type Update func(Msg, Model) (Model, Cmd)

// View produces a string which will be rendered to the terminal
type View func(Model) string

// Program is a terminal user interface
type Program struct {
	model         Model
	update        Update
	view          View
	subscriptions []Sub
	rw            io.ReadWriter
}

// Quit is a command that tells the program to exit
func Quit() Msg {
	return quitMsg{}
}

// Signals that the program should quit
type quitMsg struct{}

// NewProgram creates a new Program
func NewProgram(model Model, update Update, view View, subs []Sub) *Program {
	return &Program{
		model:         model,
		update:        update,
		view:          view,
		subscriptions: subs,
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
	p.render(model)

	// Subscribe to user input. We could move this out of here and offer it
	// as a subscription, but it blocks nicely and seems to be a common enough
	// need that we're enabling it by default.
	go func() {
		for {
			msg, _ := ReadKey(p.rw)
			msgs <- KeyMsg(msg)
		}
	}()

	// Initialize subscriptions
	go func() {
		if len(p.subscriptions) > 0 {
			for _, sub := range p.subscriptions {
				go func(s Sub) {
					for {
						msgs <- s(p.model)
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
			p.render(model)
			p.model = model
		}
	}
}

// Render a view to the terminal
func (p *Program) render(model Model) {
	view := p.view(model) + "\n"

	// We need to add carriage returns to ensure that the cursor travels to the
	// start of a column after a newline
	view = strings.Replace(view, "\n", "\r\n", -1)

	if linesRendered > 0 {
		clearLines(linesRendered)
	}
	io.WriteString(p.rw, view)
	linesRendered = strings.Count(view, "\r\n")
}

// Hide the cursor
func hideCursor() {
	fmt.Printf(esc + "?25l")
}

// Show the cursor
func showCursor() {
	fmt.Printf(esc + "?25h")
}

// Move the cursor down a given number of lines and place it at the beginning
// of the line
func cursorNextLine(n int) {
	fmt.Printf(esc+"%dE", n)
}

// Move the cursor up a given number of lines and place it at the beginning of
// the line
func cursorPrevLine(n int) {
	fmt.Printf(esc+"%dF", n)
}

// Clear the current line
func clearLine() {
	fmt.Printf(esc + "2K")
}

// Clear a given number of lines
func clearLines(n int) {
	clearLine()
	for i := 0; i < n; i++ {
		cursorPrevLine(1)
		clearLine()
	}
}

// Fullscreen switches to the altscreen and clears the terminal. The former
// view can be restored with ExitFullscreen().
func Fullscreen() {
	fmt.Print(esc + "?1049h" + esc + "H")
}

// ExitFullscreen exits the altscreen and returns the former terminal view
func ExitFullscreen() {
	fmt.Print(esc + "?1049l")
}

// ClearScreen clears the visible portion of the terminal. Effectively, it
// fills the terminal with blank spaces.
func ClearScreen() {
	fmt.Printf(esc + "2J" + esc + "3J" + esc + "1;1H")
}

// Invert inverts the foreground and background colors of a given string
func Invert(s string) string {
	return esc + "7m" + s + esc + "0m"
}

// UseSysLog logs to the system log. This becomes helpful when debugging since
// we can't easily print to the terminal since our TUI is occupying it!
//
// On macOS this is a just a matter of: tail -f /var/log/system.log
func UseSysLog(programName string) error {
	l, err := syslog.New(syslog.LOG_NOTICE, programName)
	if err != nil {
		return err
	}
	log.SetOutput(l)
	return nil
}
