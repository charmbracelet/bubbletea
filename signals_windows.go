//go:build windows
// +build windows

package tea

import (
	"context"
	"os"
)

// listenForResize is not available on windows because windows does not
// implement syscall.SIGWINCH.
func listenForResize(_ context.Context, _ *os.File, _ chan Msg, _ chan error, done chan struct{}) {
	close(done)
}
