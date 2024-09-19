package tea

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
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

// setAltScreenBuffer restores the terminal screen buffer state.
func (p *Program) setAltScreenBuffer(on bool) {
	if on {
		// Ensure that the terminal is cleared, even when it doesn't support
		// alt screen (or alt screen support is disabled, like GNU screen by
		// default).
		p.execute(ansi.EraseEntireDisplay)
		p.execute(ansi.MoveCursorOrigin)
	}

	// cmd.exe and other terminals keep separate cursor states for the AltScreen
	// and the main buffer. We have to explicitly reset the cursor visibility
	// whenever we exit AltScreen.
	if !p.modes[ansi.CursorVisibilityMode] {
		p.execute(ansi.HideCursor)
	} else {
		p.execute(ansi.ShowCursor)
	}
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p *Program) restoreTerminalState() error {
	if p.modes[ansi.BracketedPasteMode] {
		p.execute(ansi.DisableBracketedPaste)
	}
	if !p.modes[ansi.CursorVisibilityMode] {
		p.execute(ansi.ShowCursor)
	}
	if p.modes[ansi.MouseCellMotionMode] || p.modes[ansi.MouseAllMotionMode] {
		p.execute(ansi.DisableMouseCellMotion)
		p.execute(ansi.DisableMouseAllMotion)
		p.execute(ansi.DisableMouseSgrExt)
	}
	if p.keyboard.modifyOtherKeys != 0 {
		p.execute(ansi.DisableModifyOtherKeys)
	}
	if p.keyboard.kittyFlags != 0 {
		p.execute(ansi.DisableKittyKeyboard)
	}
	if p.modes[ansi.ReportFocusMode] {
		p.execute(ansi.DisableReportFocus)
	}
	if p.modes[ansi.GraphemeClusteringMode] {
		p.execute(ansi.DisableGraphemeClustering)
	}
	if p.modes[ansi.AltScreenBufferMode] {
		p.execute(ansi.DisableAltScreenBuffer)
		// cmd.exe and other terminals keep separate cursor states for the AltScreen
		// and the main buffer. We have to explicitly reset the cursor visibility
		// whenever we exit AltScreen.
		p.execute(ansi.ShowCursor)

		// give the terminal a moment to catch up
		time.Sleep(time.Millisecond * 10) //nolint:gomnd
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
	var term string
	for i := len(p.environ) - 1; i >= 0; i-- {
		// We iterate backwards to find the last TERM variable set in the
		// environment. This is because the last one is the one that will be
		// used by the terminal.
		parts := strings.SplitN(p.environ[i], "=", 2)
		if len(parts) == 2 && parts[0] == "TERM" {
			term = parts[1]
			break
		}
	}

	// Initialize the input reader.
	// This need to be done after the terminal has been initialized and set to
	// raw mode.
	// On Windows, this will change the console mode to enable mouse and window
	// events.
	var flags int // TODO: make configurable through environment variables?
	drv, err := newDriver(p.input, term, flags)
	if err != nil {
		return err
	}

	p.inputReader = drv
	p.readLoopDone = make(chan struct{})
	go p.readLoop()

	return nil
}

func readInputs(ctx context.Context, msgs chan<- Msg, reader *driver) error {
	for {
		events, err := reader.ReadEvents()
		if err != nil {
			return err
		}

		for _, msg := range events {
			incomingMsgs := []Msg{msg}

			for _, m := range incomingMsgs {
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

	p.Send(WindowSizeMsg{
		Width:  w,
		Height: h,
	})
}
