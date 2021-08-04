// +build windows

package tea

import "os"

// listenForResize is not available on windows because windows does not
// implement syscall.SIGWINCH.
func listenForResize(ctx context.Context, output *os.File, msgs chan Msg,
	errs chan error, done chan struct{}) {
	close(done)
}
