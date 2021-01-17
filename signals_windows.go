// +build windows

package tea

import "os"

// listenForResize is not available on windows because windows does not
// implement syscall.SIGWINCH.
func listenForResize(_ *os.File, _ chan Msg, _ chan error) {}
