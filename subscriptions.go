package tea

import (
	"time"
)

// NewEverMsg is used by Every to create a new message. It contains the time
// at which the timer finished.
type NewEveryMsg func(time.Time) Msg

// Every is a subscription that ticks with the system clock at the given
// duration.
//
// TODO: make it cancelable
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
