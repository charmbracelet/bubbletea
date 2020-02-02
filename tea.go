package tea

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// Msg represents an action. It's used by Update to update the UI.
type Msg interface{}

// Model contains the updatable data for an application
type Model interface{}

// Cmd is an IO operation. If it's nil it's considered a no-op.
type Cmd func(Model) Msg

// Sub is an event subscription. If it returns nil it's considered a no-op,
// but there's really no reason to have a nil subscription.
type Sub func(Model) Msg

// Subs is a keyed set of subscriptions. The key should be a unique
// identifier; two different subscriptions should not have the same key
type Subs map[string]Sub

// Subscriptions returns a map of subscriptions (Subs) our application will
// subscribe to. If Subscriptions is nil it's considered a no-op.
type Subscriptions func(Model) Subs

// subscription is an internal reference to a subscription used in subscription
// management.
type subscription struct {
	done chan struct{}
	sub  Sub
}

// subManager is used to manage active subscriptions, hence the pointers.
type subManager map[string]*subscription

// endAll stops all subscriptions and remove subscription references from
// subManager.
func (m *subManager) endAll() {
	if m != nil {
		for key, sub := range *m {
			close(sub.done)
			delete(*m, key)
		}
	}
}

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
	subscriptions Subscriptions
	model         Model
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
func NewProgram(init Init, update Update, view View, subs Subscriptions) *Program {
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
		cmd           Cmd
		subs          = make(subManager)
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
	p.model, cmd = p.init()
	if cmd != nil {
		go func() {
			cmds <- cmd
		}()
	}

	// Render initial view
	linesRendered = p.render(p.model, linesRendered)

	// Subscribe to user input. We could move this out of here and offer it
	// as a subscription, but it blocks nicely and seems to be a common enough
	// need that we're enabling it by default.
	go func() {
		for {
			msg, _ := ReadKey(os.Stdin)
			msgs <- KeyMsg(msg)
		}
	}()

	// Initialize subscriptions
	subs = p.processSubs(msgs, subs)

	// Process commands
	go func() {
		for {
			select {
			case <-done:
				return
			case cmd := <-cmds:
				if cmd != nil {
					go func() {
						msgs <- cmd(p.model)
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

			p.model, cmd = p.update(msg, p.model)            // run update
			cmds <- cmd                                      // process command (if any)
			subs = p.processSubs(msgs, subs)                 // check for new and outdated subscriptions
			linesRendered = p.render(p.model, linesRendered) // render to terminal
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
		termenv.ClearLines(linesRendered)
	}
	io.WriteString(os.Stdout, view)
	return strings.Count(view, "\r\n")
}

// Manage subscriptions. Here we run the program's Subscription function and
// inspect the functions it returns (a Subs map). If we notice existing
// subscriptions have disappeared from the map we stop those subscriptions
// by ending the Goroutines they run on. If we notice new subscriptions which
// aren't currently running, we run them as loops in a new Goroutine.
//
// This function should be called on initialization and after every update.
func (p *Program) processSubs(msgs chan Msg, activeSubs subManager) subManager {

	// Nothing to do.
	if p.subscriptions == nil && activeSubs == nil {
		return activeSubs
	}

	// There are no subscriptions. Cancel active ones and return.
	if p.subscriptions == nil && activeSubs != nil {
		activeSubs.endAll()
		return activeSubs
	}

	newSubs := p.subscriptions(p.model)

	// newSubs is an empty map. Cancel any active subscriptions and return.
	if newSubs == nil {
		activeSubs.endAll()
		return activeSubs
	}

	// Stop subscriptions that don't exist in the new subscription map
	if len(activeSubs) > 0 {
		for key, sub := range activeSubs {
			if _, exists := newSubs[key]; !exists {
				close(sub.done)
				delete(activeSubs, key)
			}
		}
	}

	// Start new subscriptions if they don't exist in the active subscription map
	if len(newSubs) > 0 {
		for key, sub := range newSubs {
			if _, exists := activeSubs[key]; !exists {

				if sub == nil {
					continue
				}

				activeSubs[key] = &subscription{
					done: make(chan struct{}),
					sub:  sub,
				}

				go func(done chan struct{}, s Sub) {
					for {
						select {
						case <-done:
							return
						case msgs <- s(p.model):
							continue
						}
					}
				}(activeSubs[key].done, activeSubs[key].sub)

			}
		}
	}

	return activeSubs
}

// AltScreen exits the altscreen. Just is just a wrapper around the termenv
// function
func AltScreen() {
	termenv.AltScreen()
}

// ExitAltScreen exits the altscreen. Just is just a wrapper around the termenv
// function
func ExitAltScreen() {
	termenv.ExitAltScreen()
}
