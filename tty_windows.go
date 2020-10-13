// +build windows

package tea

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableAnsiColors enables support for ANSI color sequences in Windows
// default console. Note that this only works with Windows 10.
func enableAnsiColors() {
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}
