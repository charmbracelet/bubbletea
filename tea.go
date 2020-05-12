package tea

import (
	"io"
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// Msg represents an action and is usually the result of an IO operation. It's
// triggers the  Update function, and henceforth, the UI.
type Msg interface{}

// Model contains the program's state.
type Model interface{}

// Cmd is an IO operation that runs once. If it's nil it's considered a no-op.
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

// Sub is an event subscription; generally a recurring IO operation. If it
// returns nil it's considered a no-op, but there's really no reason to have
// a nil subscription.
type Sub func() Msg

// Subs is a keyed set of subscriptions. The key should be a unique
// identifier: two different subscriptions should not have the same key or
// weird behavior will occur.
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
}

// Quit is a command that tells the program to exit
func Quit() Msg {
	return quitMsg{}
}

// Signals that the program should quit
type quitMsg struct{}

// batchMsg is used to perform a bunch of commands
type batchMsg []Cmd

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
		model         Model
		cmd           Cmd
		subs          = make(subManager)
		cmds          = make(chan Cmd)
		msgs          = make(chan Msg)
		errs          = make(chan error)
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
			msg, err := ReadKey(os.Stdin)
			if err != nil {
				errs <- err
			}
			msgs <- KeyMsg(msg)
		}
	}()

	// Initialize subscriptions
	subs = p.processSubs(msgs, model, subs)

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

			model, cmd = p.update(msg, model)              // run update
			cmds <- cmd                                    // process command (if any)
			subs = p.processSubs(msgs, model, subs)        // check for new and outdated subscriptions
			linesRendered = p.render(model, linesRendered) // render to terminal
		}
	}
}

// Render a view to the terminal. Returns the number of lines rendered.
func (p *Program) render(model Model, linesRendered int) int {
	view := p.view(model)

	// We need to add carriage returns to ensure that the cursor travels to the
	// start of a column after a newline
	view = strings.Replace(view, "\n", "\r\n", -1)

	if linesRendered > 0 {
		termenv.ClearLines(linesRendered)
	}
	_, _ = io.WriteString(os.Stdout, view)
	return strings.Count(view, "\r\n")
}

// Manage subscriptions. Here we run the program's Subscription function and
// inspect the functions it returns (a Subs map). If we notice existing
// subscriptions have disappeared from the map we stop those subscriptions
// by ending the Goroutines they run on. If we notice new subscriptions which
// aren't currently running, we run them as loops in a new Goroutine.
//
// This function should be called on initialization and after every update.
func (p *Program) processSubs(msgs chan Msg, model Model, activeSubs subManager) subManager {

	// Nothing to do.
	if p.subscriptions == nil && activeSubs == nil {
		return activeSubs
	}

	// There are no subscriptions. Cancel active ones and return.
	if p.subscriptions == nil && activeSubs != nil {
		activeSubs.endAll()
		return activeSubs
	}

	newSubs := p.subscriptions(model)

	// newSubs is an empty map. Cancel any active subscriptions and return.
	if newSubs == nil {
		activeSubs.endAll()
		return activeSubs
	}

	// Stop subscriptions that don't exist in the new subscription map and
	// stop subscriptions where the new subscription is mapped to a nil.
	if len(activeSubs) > 0 {
		for key, sub := range activeSubs {
			_, exists := newSubs[key]
			if !exists || exists && newSubs[key] == nil {
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
						case msgs <- s():
							continue
						}
					}
				}(activeSubs[key].done, activeSubs[key].sub)

			}
		}
	}

	return activeSubs
}

// AltScreen exits the altscreen. This is just a wrapper around the termenv
// function
func AltScreen() {
	termenv.AltScreen()
}

// ExitAltScreen exits the altscreen. This is just a wrapper around the termenv
// function
func ExitAltScreen() {
	termenv.ExitAltScreen()
}
