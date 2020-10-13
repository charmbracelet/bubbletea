package tea

import (
	"os"

	"github.com/muesli/termenv"
	"golang.org/x/crypto/ssh/terminal"
)

var origTTYState *terminal.State

func initTerminal() error {
	var err error
	origTTYState, err = terminal.MakeRaw(int(os.Stdin.Fd())) // enter raw mode
	if err != nil {
		return err
	}

	enableAnsiColors()
	termenv.HideCursor()
	return nil
}

func restoreTerminal() error {
	termenv.ShowCursor()
	return terminal.Restore(int(os.Stdin.Fd()), origTTYState) // exit raw mode
}
