package tea

import (
	"time"
)

// Every is a subscription that ticks with the system clock at the given
// duration, similar to cron. It's particularly useful if you have several
// subscriptions that need to run in sync.
func Every(duration time.Duration, newMsg func(time.Time) Msg) Sub {
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
func Tick(d time.Duration, newMsg func(time.Time) Msg) Sub {
	return func() Msg {
		t := time.NewTimer(d)
		select {
		case now := <-t.C:
			return newMsg(now)
		}
	}
}
