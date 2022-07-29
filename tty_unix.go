//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix

package tea

import (
	"os"

	"github.com/containerd/console"
)

func (p *Program) initInput() error {
	// If input's a file, use console to manage it
	if f, ok := p.input.(*os.File); ok {
		c, err := console.ConsoleFromFile(f)
		if err != nil {
			return nil //nolint:nilerr // ignore error, this was just a test
		}
		p.console = c
	}

	return nil
}

// On unix systems, RestoreInput closes any TTYs we opened for input. Note that
// we don't do this on Windows as it causes the prompt to not be drawn until
// the terminal receives a keypress rather than appearing promptly after the
// program exits.
func (p *Program) restoreInput() error {
	if p.console != nil {
		return p.console.Reset()
	}
	return nil
}

func openInputTTY() (*os.File, error) {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return nil, err
	}
	return f, nil
}
