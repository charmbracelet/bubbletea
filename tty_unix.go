// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

import (
	"errors"
	"io"
	"os"

	"github.com/containerd/console"
)

func (p *Program) initInput() error {
	if !p.inputIsTTY {
		return nil
	}

	// If input's a TTY this should always succeed.
	f, ok := p.input.(*os.File)
	if !ok {
		return errInputIsNotAFile
	}

	c, err := console.ConsoleFromFile(f)
	if err != nil {
		return nil
	}
	p.console = c

	return nil
}

// On unix systems, RestoreInput closes any TTYs we opened for input. Note that
// we don't do this on Windows as it causes the prompt to not be drawn until the
// terminal receives a keypress rather than appearing promptly after the program
// exits.
func (p *Program) restoreInput() error {
	if p.inputStatus == managedInput {
		f, ok := p.input.(*os.File)
		if !ok {
			return errors.New("could not close input")
		}
		err := f.Close()
		if err != nil {
			return err
		}
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

// enableAnsiColors is only needed for Windows, so for other systems this is
// a no-op.
func enableAnsiColors(_ io.Writer) {}
