package tea

import (
	"github.com/containerd/console"
)

var tty console.Console

func (p Program) initTerminal() error {
	if p.outputIsTTY {
		tty = console.Current()
	}

	if p.inputIsTTY {
		err := tty.SetRaw()
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
	if !p.outputIsTTY {
		return nil
	}
	showCursor(p.output)
	return tty.Reset()
}
