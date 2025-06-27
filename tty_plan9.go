package tea

import (
	"fmt"
	"os"
)

func (p *Program) initInput() (err error) {
	/*
		// Check if input is a terminal
		if f, ok := p.input.(term.File); ok && term.IsTerminal(f.Fd()) {
			p.ttyInput = f
			p.previousTtyInputState, err = term.MakeRaw(p.ttyInput.Fd())
			if err != nil {
				return fmt.Errorf("error entering raw mode: %w", err)
			}
		}

		if f, ok := p.output.(term.File); ok && term.IsTerminal(f.Fd()) {
			p.ttyOutput = f
		}
	*/
	return fmt.Errorf("no yet")
}

func openInputTTY() (*os.File, error) {
	f, err := os.Open("/dev/cons")
	if err != nil {
		return nil, fmt.Errorf("could not open a new TTY: %w", err)
	}
	return f, nil
}

const suspendSupported = false

// Send SIGTSTP to the entire process group.
func suspendProcess() {
	/*
	   c := make(chan os.Signal, 1)
	   signal.Notify(c, syscall.SIGCONT)
	   _ = syscall.Kill(0, syscall.SIGTSTP)
	   // blocks until a CONT happens...
	   <-c
	*/
}
