//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix

package tea

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// listenForResize sends messages (or errors) when the terminal resizes.
// Argument output should be the file descriptor for the terminal; usually
// os.Stdout.
func listenForResize(ctx context.Context, output *os.File, msgs chan Msg, errs chan error, done chan struct{}) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	defer func() {
		signal.Stop(sig)
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sig:
		}

		w, h, err := term.GetSize(int(output.Fd()))
		if err != nil {
			errs <- err
		}

		select {
		case <-ctx.Done():
			return
		case msgs <- WindowSizeMsg{w, h}:
		}
	}
}
