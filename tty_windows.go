// +build windows

package tea

import (
	"os"

	"github.com/muesli/termenv"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/windows"
)

var (
	origTTYState *terminal.State
)

// enableAnsiColors enables support for ANSI color sequences in Windows
// default console. Note that this only works with Windows 10.
func enableAnsiColors() {
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

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

func restoreTerminal() {
	_ = terminal.Restore(int(os.Stdin.Fd()), origTTYState) // exit raw mode
	termenv.ShowCursor()
}
