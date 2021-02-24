package tea

import (
	"fmt"
	"os"

	"github.com/containerd/console"
)

func (p Program) initTerminal() error {
	var err error

	const assertionErrTpl = "could not create console for %s; could not perform file assertion"

	// Setup input console
	if p.inputIsTTY {
		f, ok := p.input.(*os.File)
		if ok {
			p.inputConsole, err = console.ConsoleFromFile(f)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf(assertionErrTpl, "input")
		}
	}

	// Setup output console
	if p.outputIsTTY {
		f, ok := p.output.(*os.File)
		if ok {
			p.outputConsole, err = console.ConsoleFromFile(f)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf(assertionErrTpl, "output")
		}
	}

	// Enter raw mode
	if p.inputConsole != nil {
		err := p.inputConsole.SetRaw()
		if err != nil {
			return err
		}
	}

	// Prep terminal for TUI output
	if p.outputIsTTY {
		enableAnsiColors(p.output) // windows only, no-op otherwise
		hideCursor(p.output)
	}

	return nil
}

func (p Program) restoreTerminal() error {
	if !p.outputIsTTY {
		return nil
	}

	showCursor(p.output)

	if p.outputConsole != nil {
		return p.outputConsole.Reset() // in particular, exit RAW mode
	}
	return nil
}
