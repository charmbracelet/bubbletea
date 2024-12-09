package tea

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/input"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/cancelreader"
)

func (p *Program) suspend() {
	if err := p.ReleaseTerminal(); err != nil {
		// If we can't release input, abort.
		return
	}

	suspendProcess()

	_ = p.RestoreTerminal()
	go p.Send(ResumeMsg{})
}

func (p *Program) initTerminal() error {
	if _, ok := p.renderer.(*nilRenderer); ok {
		// No need to initialize the terminal if we're not rendering
		return nil
	}

	return p.initInput()
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p *Program) restoreTerminalState() error {
	if p.modes[ansi.BracketedPasteMode] {
		p.execute(ansi.ResetBracketedPasteMode)
	}
	if p.modes[ansi.ButtonEventMouseMode] || p.modes[ansi.AnyEventMouseMode] {
		p.execute(ansi.ResetButtonEventMouseMode)
		p.execute(ansi.ResetAnyEventMouseMode)
		p.execute(ansi.ResetSgrExtMouseMode)
	}
	if p.keyboard.modifyOtherKeys != 0 {
		p.execute(ansi.DisableModifyOtherKeys)
	}
	if p.keyboard.kittyFlags != 0 {
		p.execute(ansi.DisableKittyKeyboard)
	}
	if p.modes[ansi.FocusEventMode] {
		p.execute(ansi.ResetFocusEventMode)
	}
	if p.modes[ansi.GraphemeClusteringMode] {
		p.execute(ansi.ResetGraphemeClusteringMode)
	}

	// Restore terminal colors.
	if p.setBg != nil {
		p.execute(ansi.ResetBackgroundColor)
	}
	if p.setFg != nil {
		p.execute(ansi.ResetForegroundColor)
	}
	if p.setCc != nil {
		p.execute(ansi.ResetCursorColor)
	}

	return p.restoreInput()
}

// restoreInput restores the tty input to its original state.
func (p *Program) restoreInput() error {
	if p.ttyInput != nil && p.previousTtyInputState != nil {
		if err := term.Restore(p.ttyInput.Fd(), p.previousTtyInputState); err != nil {
			return fmt.Errorf("error restoring console: %w", err)
		}
	}
	if p.ttyOutput != nil && p.previousOutputState != nil {
		if err := term.Restore(p.ttyOutput.Fd(), p.previousOutputState); err != nil {
			return fmt.Errorf("error restoring console: %w", err)
		}
	}
	return nil
}

// initInputReader (re)commences reading inputs.
func (p *Program) initInputReader() error {
	term := p.getenv("TERM")

	// Initialize the input reader.
	// This need to be done after the terminal has been initialized and set to
	// raw mode.
	// On Windows, this will change the console mode to enable mouse and window
	// events.
	var flags int // TODO: make configurable through environment variables?
	drv, err := input.NewReader(p.input, term, flags)
	if err != nil {
		return err
	}

	if p.traceInput {
		drv.SetLogger(log.Default())
	}
	p.inputReader = drv
	p.readLoopDone = make(chan struct{})
	go p.readLoop()

	return nil
}

func readInputs(ctx context.Context, msgs chan<- Msg, reader *input.Reader) error {
	for {
		events, err := reader.ReadEvents()
		if err != nil {
			return err
		}

		for _, msg := range events {
			if m := translateInputEvent(msg); m != nil {
				select {
				case msgs <- m:
				case <-ctx.Done():
					err := ctx.Err()
					if err != nil {
						err = fmt.Errorf("found context error while reading input: %w", err)
					}
					return err
				}
			}
		}
	}
}

func (p *Program) readLoop() {
	defer close(p.readLoopDone)

	err := readInputs(p.ctx, p.msgs, p.inputReader)
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

	var resizeMsg WindowSizeMsg
	resizeMsg.Width = w
	resizeMsg.Height = h
	p.Send(resizeMsg)
}
