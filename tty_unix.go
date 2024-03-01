//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tea

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func (p *Program) initInput() (err error) {
	// Check if input is a terminal
	if f, ok := p.input.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		p.tty = f
		p.previousTtyState, err = term.MakeRaw(int(p.tty.Fd()))
		if err != nil {
			return fmt.Errorf("error entering raw mode: %w", err)
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
