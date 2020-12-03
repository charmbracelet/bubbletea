package tea

import (
	"io"

	"github.com/containerd/console"
)

var tty console.Console

func initTerminal(w io.Writer) error {
	tty = console.Current()
	err := tty.SetRaw()
	if err != nil {
		return err
	}

	enableAnsiColors()
	hideCursor(w)
	return nil
}

func restoreTerminal(w io.Writer) error {
	showCursor(w)
	return tty.Reset()
}
