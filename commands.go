package tea

// Convenience commands. Note part of the Boba runtime, but potentially handy.

import (
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

// Every is a command that ticks in sync with the system clock. So, if you
// wanted to tick with the system clock every second, minute or hour you
// could use this. It's also handy for having different things tick in sync.
//
// Note that because we're ticking with the system clock the tick will likely
// not run for the entire specified duration. For example, if we're ticking for
// one minute and the clock is at 12:34:20 then the next tick will happen at
// 12:35:00, 40 seconds later.
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

// TerminalSizeMsg defines an interface for a message that contains terminal
// sizing as sent by GetTerminalSize.
type TerminalSizeMsg interface {

	// Size returns the terminal width and height, in that order
	Size() (int, int)

	// Error returns the error, if any, received when fetching the terminal size
	Error() error
}

// GetTerminalSize is a command used to retrieve the terminal dimensions. Pass
// a function used to construct your message, which would implement the
// TerminalSizeMsg interaface.
func GetTerminalSize(newMsgFunc func(int, int, error) TerminalSizeMsg) Cmd {
	w, h, err := terminal.GetSize(int(os.Stdout.Fd()))
	return func() Msg {
		return newMsgFunc(w, h, err)
	}
}
