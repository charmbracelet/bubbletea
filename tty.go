package tea

import (
	"errors"
	"io"

	"github.com/muesli/cancelreader"
)

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

	hideCursor(p.output)
	return nil
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p Program) restoreTerminalState() error {
	showCursor(p.output)

	if p.console != nil {
		err := p.console.Reset()
		if err != nil {
			return err
		}
	}

	return p.restoreInput()
}

// initCancelReader (re)commences reading inputs.
func (p *Program) initCancelReader() error {
	var err error
	p.cancelReader, err = cancelreader.NewReader(p.input)
	if err != nil {
		return err
	}

	p.readLoopDone = make(chan struct{})
	go func() {
		defer close(p.readLoopDone)

		for {
			if p.ctx.Err() != nil {
				return
			}

			msgs, err := readInputs(p.cancelReader)
			if err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, cancelreader.ErrCanceled) {
					p.errs <- err
				}

				return
			}

			for _, msg := range msgs {
				p.msgs <- msg
			}
		}
	}()

	return nil
}

// cancelInput cancels the input reader.
func (p *Program) cancelInput() {
	p.cancelReader.Cancel()
}
