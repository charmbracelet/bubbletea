//go:build windows
// +build windows

package tea

import (
	"os"
	"os/exec"
)

// preparePtyCommand is a no-op on Windows, where ConPTY handles the session
// and console setup of the command.
func preparePtyCommand(*exec.Cmd) {}

// ptyResizeSignals returns a channel that never receives a value: Windows has
// no resize signal, and ConPTY reports resizes through the pseudo-terminal
// itself.
func ptyResizeSignals() (<-chan os.Signal, func()) {
	return nil, func() {}
}
