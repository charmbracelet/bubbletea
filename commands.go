package tea

// Convenience commands. Note part of the Bubble Tea core, but potentially
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
func Every(duration time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		n := time.Now()
		d := n.Truncate(duration).Add(duration).Sub(n)
		t := time.NewTimer(d)
		return fn(<-t.C)
	}
}

// Tick is a command that at an interval independent of the system clock at the
// given duration. That is, the timer begins when precisely when invoked, and
// runs for its entire duration.
func Tick(d time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		t := time.NewTimer(d)
		return fn(<-t.C)
	}
}
