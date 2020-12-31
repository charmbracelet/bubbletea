// +build windows

package tea

import (
	"io"
	"os"

	"golang.org/x/sys/windows"
)

// enableAnsiColors enables support for ANSI color sequences in Windows
// default console. Note that this only works with Windows 10.
func enableAnsiColors(w io.Writer) {
	f, ok := w.(*os.File)
	if !ok {
		return
	}

	stdout := windows.Handle(f.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}
