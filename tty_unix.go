// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

import (
	"github.com/muesli/termenv"
	"github.com/pkg/term"
)

var (
	tty *term.Term
)

func initTerminal() error {
	var err error
	tty, err = term.Open("/dev/tty")
	if err != nil {
		return err
	}

	err = tty.SetRaw()
	if err != nil {
		return err
	}

	termenv.HideCursor()
	return nil
}

func restoreTerminal() {
	termenv.ShowCursor()
	_ = tty.Restore()
}
