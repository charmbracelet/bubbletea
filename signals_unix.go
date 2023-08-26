//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix

package tea

import (
	"os"
	"os/signal"
	"syscall"
)

// listenForHangup sends a message when the terminal is closed.
// Argument output should be the file descriptor for the terminal; usually
// os.Stdout.
func (p *Program) listenForHangup(done chan struct{}) {
	p.listen(done, syscall.SIGHUP, p.sendHangupMsg)
}

func (p *Program) sendHangupMsg() {
	p.Send(HangupMsg{})
}

// listenForResize sends messages (or errors) when the terminal resizes.
// Argument output should be the file descriptor for the terminal; usually
// os.Stdout.
func (p *Program) listenForResize(done chan struct{}) {
	p.listen(done, syscall.SIGWINCH, p.checkResize)
}

func (p *Program) listen(done chan struct{}, sig syscall.Signal, f func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig)

	defer func() {
		signal.Stop(c)
		close(done)
	}()

	for {
		select {
			case <-p.ctx.Done():
			return
			case <-c:
		}

		f()
	}
}
