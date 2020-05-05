package tea

import (
	"time"
)

// Every is a subscription that ticks with the system clock at the given
// duration
//
// TODO: make it cancelable
func Every(duration time.Duration, msg Msg) Sub {
	return func() Msg {
		n := time.Now()
		d := n.Truncate(duration).Add(duration).Sub(n)
		t := time.NewTimer(d)
		select {
		case <-t.C:
			return msg
		}
	}
}
