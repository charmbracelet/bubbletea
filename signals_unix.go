// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// listenForResize sends messages (or errors) when the terminal resizes.
// Argument output should be the file descriptor for the terminal; usually
// os.Stdout.
func listenForResize(output *os.File, msgs chan Msg, errs chan error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)
	for {
		<-sig
		w, h, err := term.GetSize(int(output.Fd()))
		if err != nil {
			errs <- err
		}
		msgs <- WindowSizeMsg{w, h}
	}
}
