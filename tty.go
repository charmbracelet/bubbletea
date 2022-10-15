package tea

import (
	"errors"
	"io"
	"os"
	"time"

	isatty "github.com/mattn/go-isatty"
	"github.com/muesli/cancelreader"
	"golang.org/x/term"
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

	p.renderer.hideCursor()
	return nil
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p *Program) restoreTerminalState() error {
	if p.renderer != nil {
		p.renderer.showCursor()
		p.renderer.disableMouseCellMotion()
		p.renderer.disableMouseAllMotion()

		if p.renderer.altScreen() {
			p.renderer.exitAltScreen()

			// give the terminal a moment to catch up
			time.Sleep(time.Millisecond * 10)
		}
	}

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
	go p.readLoop()

	return nil
}

func (p *Program) readLoop() {
	defer close(p.readLoopDone)

	for {
		if p.ctx.Err() != nil {
			return
		}

		msgs, err := readInputs(p.cancelReader)
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, cancelreader.ErrCanceled) {
				select {
				case <-p.ctx.Done():
				case p.errs <- err:
				}
			}

			return
		}

		for _, msg := range msgs {
			p.msgs <- msg
		}
	}
}

// waitForReadLoop waits for the cancelReader to finish its read loop.
func (p *Program) waitForReadLoop() {
	select {
	case <-p.readLoopDone:
	case <-time.After(500 * time.Millisecond):
		// The read loop hangs, which means the input
		// cancelReader's cancel function has returned true even
		// though it was not able to cancel the read.
	}
}

// checkResize detects the current size of the output and informs the program
// via a WindowSizeMsg.
func (p *Program) checkResize() {
	f, ok := p.output.TTY().(*os.File)
	if !ok || !isatty.IsTerminal(f.Fd()) {
		// can't query window size
		return
	}

	w, h, err := term.GetSize(int(f.Fd()))
	if err != nil {
		select {
		case <-p.ctx.Done():
		case p.errs <- err:
		}

		return
	}

	p.Send(WindowSizeMsg{
		Width:  w,
		Height: h,
	})
}
