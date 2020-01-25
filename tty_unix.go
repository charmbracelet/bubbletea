// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

import (
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

	tty.SetRaw()
	hideCursor()
	return nil
}

func restoreTerminal() {
	showCursor()
	tty.Restore()
}
