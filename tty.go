package tea

import (
	"errors"
)

var errInputIsNotAFile = errors.New("input is not a file")

func (p *Program) initTerminal() error {
	err := p.initInput()
	if err != nil {
		return err
	}

	if p.inputIsTTY {
		if p.console == nil {
			return errors.New("no console")
		}
		err = p.console.SetRaw()
		if err != nil {
			return err
		}
	}

	if p.outputIsTTY {
		enableAnsiColors(p.output)
		hideCursor(p.output)
	}

	return nil
}

func (p Program) restoreTerminal() error {
	if p.outputIsTTY {
		showCursor(p.output)
	}

	if err := p.restoreInput(); err != nil {
		return err
	}

	// Console will only be set if input is a TTY.
	if p.console != nil {
		err := p.console.Reset()
		if err != nil {
			return err
		}
	}

	return nil
}
