package tea

import (
	"github.com/containerd/console"
	"github.com/muesli/termenv"
)

var tty console.Console

func initTerminal() error {
	tty = console.Current()
	err := tty.SetRaw()
	if err != nil {
		return err
	}

	enableAnsiColors()
	termenv.HideCursor()
	return nil
}

func restoreTerminal() error {
	termenv.ShowCursor()
	return tty.Reset()
}
