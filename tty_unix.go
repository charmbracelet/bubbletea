//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tea

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

func (p *Program[T]) initInput() (err error) {
	// Check if input is a terminal
	if f, ok := p.Input.(term.File); ok && term.IsTerminal(f.Fd()) {
		p.ttyInput = f
		p.previousTtyInputState, err = term.MakeRaw(p.ttyInput.Fd())
		if err != nil {
			return fmt.Errorf("error entering raw mode: %w", err)
		}

		// OPTIM: We can use hard tabs to optimize cursor movements if the
		// terminal doesn't have tab expansion enabled.
		p.useHardTabs = p.previousTtyInputState.Oflag&unix.TABDLY == 0
	}

	if f, ok := p.Output.(*safeWriter).Writer().(term.File); ok && term.IsTerminal(f.Fd()) {
		p.ttyOutput = f
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

const suspendSupported = true

// Send SIGTSTP to the entire process group.
func suspendProcess() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGCONT)
	_ = syscall.Kill(0, syscall.SIGTSTP)
	// blocks until a CONT happens...
	<-c
}
