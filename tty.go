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

	if p.console != nil {
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

	if p.console != nil {
		err := p.console.Reset()
		if err != nil {
			return err
		}
	}

	if err := p.restoreInput(); err != nil {
		return err
	}

	return nil
}
