//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tea

import (
	"os"
	"os/signal"
	"syscall"
)

// listenForResize sends messages (or errors) when the terminal resizes.
// Argument output should be the file descriptor for the terminal; usually
// os.Stdout.
func (p *Program[T]) listenForResize(done chan struct{}) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	defer func() {
		signal.Stop(sig)
		close(done)
	}()

	for {
		select {
		case <-p.Context.Done():
			return
		case <-sig:
		}

		p.checkResize()
	}
}
