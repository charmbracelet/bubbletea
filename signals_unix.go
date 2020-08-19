// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// listenForResize sends messages (or errors) when the terminal resizes.
// Argument output should be the file descriptor for the terminal; usually
// os.Stdout.
func listenForResize(output *os.File, msgs chan Msg, errs chan error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)
	for {
		<-sig
		w, h, err := terminal.GetSize(int(output.Fd()))
		if err != nil {
			errs <- err
		}
		msgs <- WindowSizeMsg{w, h}
	}
}
