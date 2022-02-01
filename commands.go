package tea

// Convenience commands. Not part of the Bubble Tea core, but potentially
// handy.

import (
	"time"
)

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
//   type TickMsg time.Time
//
//   cmd := Every(time.Second, func(t time.Time) Msg {
//      return TickMsg(t)
//   })
func Every(duration time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		n := time.Now()
		d := n.Truncate(duration).Add(duration).Sub(n)
		t := time.NewTimer(d)
		return fn(<-t.C)
	}
}

// Tick produces a command at an interval independent of the system clock at
// the given duration. That is, the timer begins when precisely when invoked,
// and runs for its entire duration.
//
// To produce the command, pass a duration and a function which returns
// a message containing the time at which the tick occurred.
//
//   type TickMsg time.Time
//
//   cmd := Tick(time.Second, func(t time.Time) Msg {
//      return TickMsg(t)
//   })
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		t := time.NewTimer(d)
		return fn(<-t.C)
	}
}

// Sequentially produces a command that sequentially executes the given
// commands.
// The Msg returned is the first non-nil message returned by a Cmd.
//
//   func saveStateCmd() Msg {
//   	if err := save(); err != nil {
//    		return errMsg{err}
//    	}
//   	return nil
//   }
//
//   cmd := Sequentially(saveStateCmd, Quit)
//
func Sequentially(cmds ...Cmd) Cmd {
	return func() Msg {
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			if msg := cmd(); msg != nil {
				return msg
			}
		}
		return nil
	}
}
