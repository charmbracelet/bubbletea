package tea

import (
	"time"
)

// NewEveryMsg is used by Every to create a new message. It contains the time
// at which the timer finished.
type NewEveryMsg func(time.Time) Msg

// Every is a subscription that ticks with the system clock at the given
// duration, similar to cron. It's particularly useful if you have several
// subscriptions that need to run in sync.
//
// TODO: make it cancelable.
func Every(duration time.Duration, newMsg NewEveryMsg) Sub {
	return func() Msg {
		n := time.Now()
		d := n.Truncate(duration).Add(duration).Sub(n)
		t := time.NewTimer(d)
		select {
		case now := <-t.C:
			return newMsg(now)
		}
	}
}

// Tick is a subscription that at an interval independent of the system clock
// at the given duration. That is, it begins precisely when invoked.
//
// TODO: make it cancelable.
func Tick(d time.Duration, newMsg NewEveryMsg) Sub {
	return func() Msg {
		t := time.NewTimer(d)
		select {
		case now := <-t.C:
			return newMsg(now)
		}
	}
}
