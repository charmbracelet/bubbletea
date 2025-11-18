package tea

import (
	"fmt"
	"os"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/term"
)

func (p *Program) suspend() {
	if err := p.releaseTerminal(true); err != nil {
		// If we can't release input, abort.
		return
	}

	suspendProcess()

	_ = p.RestoreTerminal()
	go p.Send(ResumeMsg{})
}

func (p *Program) initTerminal() error {
	if p.disableRenderer {
		return nil
	}
	return p.initInput()
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p *Program) restoreTerminalState() error {
	// Flush queued commands.
	_ = p.flush()

	return p.restoreInput()
}

// restoreInput restores the tty input to its original state.
func (p *Program) restoreInput() error {
	if p.ttyInput != nil && p.previousTtyInputState != nil {
		if err := term.Restore(p.ttyInput.Fd(), p.previousTtyInputState); err != nil {
			return fmt.Errorf("bubbletea: error restoring console: %w", err)
		}
	}
	if p.ttyOutput != nil && p.previousOutputState != nil {
		if err := term.Restore(p.ttyOutput.Fd(), p.previousOutputState); err != nil {
			return fmt.Errorf("bubbletea: error restoring console: %w", err)
		}
	}
	return nil
}

// initInputReader (re)commences reading inputs.
func (p *Program) initInputReader(cancel bool) error {
	if cancel && p.cancelReader != nil {
		p.cancelReader.Cancel()
		p.waitForReadLoop()
	}

	term := p.environ.Getenv("TERM")

	// Initialize the input reader.
	// This need to be done after the terminal has been initialized and set to
	// raw mode.

	var err error
	p.cancelReader, err = uv.NewCancelReader(p.input)
	if err != nil {
		return fmt.Errorf("bubbletea: could not create cancelable reader: %w", err)
	}

	drv := uv.NewTerminalReader(p.cancelReader, term)
	drv.SetLogger(p.logger)
	p.inputScanner = drv
	p.readLoopDone = make(chan struct{})

	go p.readLoop()

	return nil
}

func (p *Program) readLoop() {
	defer close(p.readLoopDone)

	if err := p.inputScanner.StreamEvents(p.ctx, p.msgs); err != nil {
		select {
		case <-p.ctx.Done():
			return
		case p.errs <- err:
		}
	}
}

// waitForReadLoop waits for the cancelReader to finish its read loop.
func (p *Program) waitForReadLoop() {
	select {
	case <-p.readLoopDone:
	case <-time.After(500 * time.Millisecond): //nolint:mnd
		// The read loop hangs, which means the input
		// cancelReader's cancel function has returned true even
		// though it was not able to cancel the read.
	}
}

// checkResize detects the current size of the output and informs the program
// via a WindowSizeMsg.
func (p *Program) checkResize() {
	if p.ttyOutput == nil {
		// can't query window size
		return
	}

	w, h, err := term.GetSize(p.ttyOutput.Fd())
	if err != nil {
		select {
		case <-p.ctx.Done():
		case p.errs <- err:
		}

		return
	}

	p.width, p.height = w, h
	p.Send(WindowSizeMsg{Width: w, Height: h})
}

// OpenTTY opens the running terminal's TTY for reading and writing.
func OpenTTY() (*os.File, *os.File, error) {
	in, out, err := uv.OpenTTY()
	if err != nil {
		return nil, nil, fmt.Errorf("bubbletea: could not open TTY: %w", err)
	}
	return in, out, nil
}
