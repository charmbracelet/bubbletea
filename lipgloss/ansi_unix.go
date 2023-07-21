//go:build !windows
// +build !windows

package lipgloss

// enableLegacyWindowsANSI is only needed on Windows.
func enableLegacyWindowsANSI() {}
