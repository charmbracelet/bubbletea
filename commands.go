package tea

import (
	"time"
)

// Batch performs a bunch of commands concurrently with no ordering guarantees
// about the results. Use a Batch to return several commands.
//
// Example:
//
//	    func (m model) Init() (Model, Cmd) {
//		       return m, tea.Batch(someCommand, someOtherCommand)
//	    }
func Batch(cmds ...Cmd) Cmd {
	return compactCmds[BatchMsg](cmds)
}

// BatchMsg is a message used to perform a bunch of commands concurrently with
// no ordering guarantees. You can send a BatchMsg with Batch.
type BatchMsg []Cmd

// Sequence runs the given commands one at a time, in order. Contrast this with
// Batch, which runs commands concurrently.
func Sequence(cmds ...Cmd) Cmd {
	return compactCmds[sequenceMsg](cmds)
}

// sequenceMsg is used internally to run the given commands in order.
type sequenceMsg []Cmd

// compactCmds ignores any nil commands in cmds, and returns the most direct
// command possible. That is, considering the non-nil commands, if there are
// none it returns nil, if there is exactly one it returns that command
// directly, else it returns the non-nil commands as type T.
func compactCmds[T ~[]Cmd](cmds []Cmd) Cmd {
	var validCmds []Cmd
	for _, c := range cmds {
		if c == nil {
			continue
		}
		validCmds = append(validCmds, c)
	}
	switch len(validCmds) {
	case 0:
		return nil
	case 1:
		return validCmds[0]
	default:
		return func() Msg {
			return T(validCmds)
		}
	}
}

// Every is a command that ticks in sync with the system clock. So, if you
// wanted to tick with the system clock every second, minute or hour you
// could use this. It's also handy for having different things tick in sync.
//
// Because we're ticking with the system clock the tick will likely not run for
// the entire specified duration. For example, if we're ticking for one minute
// and the clock is at 12:34:20 then the next tick will happen at 12:35:00, 40
// seconds later.
//
// To produce the command, pass a duration and a function which returns
// a message containing the time at which the tick occurred.
//
//	type TickMsg time.Time
//
//	cmd := Every(time.Second, func(t time.Time) Msg {
//	   return TickMsg(t)
//	})
//
// Beginners' note: Every sends a single message and won't automatically
// dispatch messages at an interval. To do that, you'll want to return another
// Every command after receiving your tick message. For example:
//
//	type TickMsg time.Time
//
//	// Send a message every second.
//	func tickEvery() Cmd {
//	    return Every(time.Second, func(t time.Time) Msg {
//	        return TickMsg(t)
//	    })
//	}
//
//	func (m model) Init() (Model, Cmd) {
//	    // Start ticking.
//	    return m, tickEvery()
//	}
//
//	func (m model) Update(msg Msg) (Model, Cmd) {
//	    switch msg.(type) {
//	    case TickMsg:
//	        // Return your Every command again to loop.
//	        return m, tickEvery()
//	    }
//	    return m, nil
//	}
//
// Every is analogous to Tick in the Elm Architecture.
func Every(duration time.Duration, fn func(time.Time) Msg) Cmd {
	n := time.Now()
	d := n.Truncate(duration).Add(duration).Sub(n)
	t := time.NewTimer(d)
	return func() Msg {
		ts := <-t.C
		t.Stop()
		for len(t.C) > 0 {
			<-t.C
		}
		return fn(ts)
	}
}

// Tick produces a command at an interval independent of the system clock at
// the given duration. That is, the timer begins precisely when invoked,
// and runs for its entire duration.
//
// To produce the command, pass a duration and a function which returns
// a message containing the time at which the tick occurred.
//
//	type TickMsg time.Time
//
//	cmd := Tick(time.Second, func(t time.Time) Msg {
//	   return TickMsg(t)
//	})
//
// Beginners' note: Tick sends a single message and won't automatically
// dispatch messages at an interval. To do that, you'll want to return another
// Tick command after receiving your tick message. For example:
//
//	type TickMsg time.Time
//
//	func doTick() Cmd {
//	    return Tick(time.Second, func(t time.Time) Msg {
//	        return TickMsg(t)
//	    })
//	}
//
//	func (m model) Init() (Model, Cmd) {
//	    // Start ticking.
//	    return m, doTick()
//	}
//
//	func (m model) Update(msg Msg) (Model, Cmd) {
//	    switch msg.(type) {
//	    case TickMsg:
//	        // Return your Tick command again to loop.
//	        return m, doTick()
//	    }
//	    return m, nil
//	}
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd {
	t := time.NewTimer(d)
	return func() Msg {
		ts := <-t.C
		t.Stop()
		for len(t.C) > 0 {
			<-t.C
		}
		return fn(ts)
	}
}

// Chain creates a command that pipes the output of each command as input to
// the next. Unlike [Sequence], which simply runs commands in order and sends
// each result to Update independently, Chain passes each command's resulting
// Msg to the next function in the chain. The final Msg is sent to Update.
//
// Each function in the chain receives the Msg produced by the previous command
// and returns a new Cmd. The first function receives nil as its input.
//
// This is useful when you need to transform or react to a command's result
// before deciding what to do next, without going through the Update cycle.
//
// Example:
//
//	cmd := tea.Chain(
//	    func(msg tea.Msg) tea.Cmd {
//	        // First step: start an HTTP request
//	        return fetchData
//	    },
//	    func(msg tea.Msg) tea.Cmd {
//	        // Second step: msg contains the result of fetchData
//	        resp := msg.(fetchResponseMsg)
//	        return processData(resp)
//	    },
//	    func(msg tea.Msg) tea.Cmd {
//	        // Third step: msg contains the result of processData
//	        return func() tea.Msg { return doneMsg{} }
//	    },
//	)
func Chain(steps ...func(Msg) Cmd) Cmd {
	if len(steps) == 0 {
		return nil
	}
	return func() Msg {
		var msg Msg
		for _, step := range steps {
			cmd := step(msg)
			if cmd == nil {
				msg = nil
				continue
			}
			msg = cmd()
		}
		return msg
	}
}

type windowSizeMsg struct{}

// RequestWindowSize is a command that queries the terminal for its current
// size. It delivers the results to Update via a [WindowSizeMsg]. Keep in mind
// that WindowSizeMsgs will automatically be delivered to Update when the
// [Program] starts and when the window dimensions change so in many cases you
// will not need to explicitly invoke this command.
func RequestWindowSize() Msg {
	return windowSizeMsg{}
}
