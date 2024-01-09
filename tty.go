package tea

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/muesli/cancelreader"
	"golang.org/x/term"
)

func (p *Program) initTerminal() error {
	if err := p.initInput(); err != nil {
		return err
	}

	p.renderer.hideCursor()
	return nil
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p *Program) restoreTerminalState() error {
	if p.renderer != nil {
		p.renderer.disableBracketedPaste()
		p.renderer.showCursor()
		p.disableMouse()

		if p.renderer.altScreen() {
			p.renderer.exitAltScreen()

			// give the terminal a moment to catch up
			time.Sleep(time.Millisecond * 10) //nolint:gomnd
		}
	}

	return p.restoreInput()
}

// restoreInput restores the tty input to its original state.
func (p *Program) restoreInput() error {
	if p.tty != nil && p.previousTtyState != nil {
		if err := term.Restore(int(p.tty.Fd()), p.previousTtyState); err != nil {
			return fmt.Errorf("error restoring console: %w", err)
		}
	}
	return nil
}

// initCancelReader (re)commences reading inputs.
func (p *Program) initCancelReader() error {
	var err error
	p.cancelReader, err = newInputReader(p.input)
	if err != nil {
		return fmt.Errorf("error creating cancelreader: %w", err)
	}

	p.readLoopDone = make(chan struct{})
	go p.readLoop()

	return nil
}

func (p *Program) readLoop() {
	defer close(p.readLoopDone)

	err := readInputs(p.ctx, p.msgs, p.cancelReader)
	if !errors.Is(err, io.EOF) && !errors.Is(err, cancelreader.ErrCanceled) {
		select {
		case <-p.ctx.Done():
		case p.errs <- err:
		}
	}
}

// waitForReadLoop waits for the cancelReader to finish its read loop.
func (p *Program) waitForReadLoop() {
	select {
	case <-p.readLoopDone:
	case <-time.After(500 * time.Millisecond): //nolint:gomnd
		// The read loop hangs, which means the input
		// cancelReader's cancel function has returned true even
		// though it was not able to cancel the read.
	}
}

// checkResize detects the current size of the output and informs the program
// via a WindowSizeMsg.
func (p *Program) checkResize() {
	f, ok := p.output.TTY().(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
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
