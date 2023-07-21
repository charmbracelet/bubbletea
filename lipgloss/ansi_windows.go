//go:build windows
// +build windows

package lipgloss

import (
	"sync"

	"github.com/muesli/termenv"
)

var enableANSI sync.Once

// enableANSIColors enables support for ANSI color sequences in the Windows
// default console (cmd.exe and the PowerShell application). Note that this
// only works with Windows 10. Also note that Windows Terminal supports colors
// by default.
func enableLegacyWindowsANSI() {
	enableANSI.Do(func() {
		_, _ = termenv.EnableWindowsANSIConsole()
	})
}
