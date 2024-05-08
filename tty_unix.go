//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tea

import (
	"fmt"
	"os"

	"github.com/charmbracelet/x/exp/term"
	"golang.org/x/sys/unix"
)

func (p *Program) initInput() (err error) {
	// Check if input is a terminal
	if f, ok := p.input.(*os.File); ok && term.IsTerminal(f.Fd()) {
		p.tty = f
		p.previousTtyState, err = term.GetState(p.tty.Fd())
		if err != nil {
			return fmt.Errorf("error entering raw mode: %w", err)
		}

		state := &term.State{}
		state.Termios = p.previousTtyState.Termios

		// XXX: We set the terminal to raw mode + OPOST (output processing) here.
		state.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
		state.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
		state.Cflag &^= unix.CSIZE | unix.PARENB
		state.Oflag |= unix.OPOST
		state.Cflag |= unix.CS8
		state.Cc[unix.VMIN] = 1
		state.Cc[unix.VTIME] = 0

		if err := term.SetState(p.tty.Fd(), state); err != nil {
			return fmt.Errorf("error setting terminal state: %w", err)
		}
	}

	return nil
}

func openInputTTY() (*os.File, error) {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return nil, fmt.Errorf("could not open a new TTY: %w", err)
	}
	return f, nil
}
