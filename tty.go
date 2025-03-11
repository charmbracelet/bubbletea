package tea

import (
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
	if err := p.releaseTerminal(true); err != nil {
		// If we can't release input, abort.
		return
	}

	suspendProcess()

	_ = p.RestoreTerminal()
	go p.Send(ResumeMsg{})
}

func (p *Program) initTerminal() error {
	if !hasView(p.initialModel) {
		// No need to initialize the terminal if we're not rendering
		return nil
	}

	return p.initInput()
}

// restoreTerminalState restores the terminal to the state prior to running the
// Bubble Tea program.
func (p *Program) restoreTerminalState() error {
	// We don't need to reset [ansi.AltScreenSaveCursorMode] and
	// [ansi.TextCursorEnableMode] because they are automatically reset when we
	// close the renderer. See [screenRenderer.close] and
	// [cellbuf.Screen.Close].

	if p.modes.IsSet(ansi.BracketedPasteMode) {
		p.execute(ansi.ResetBracketedPasteMode)
	}

	btnEvents := p.modes.IsSet(ansi.ButtonEventMouseMode)
	allEvents := p.modes.IsSet(ansi.AnyEventMouseMode)
	if btnEvents || allEvents {
		if btnEvents {
			p.execute(ansi.ResetButtonEventMouseMode)
		}
		if allEvents {
			p.execute(ansi.ResetAnyEventMouseMode)
		}
		p.execute(ansi.ResetSgrExtMouseMode)
	}
	if p.activeEnhancements.modifyOtherKeys != 0 {
		p.execute(ansi.ResetModifyOtherKeys)
	}
	if p.activeEnhancements.kittyFlags != 0 {
		p.execute(ansi.DisableKittyKeyboard)
	}
	if p.modes.IsSet(ansi.FocusEventMode) {
		p.execute(ansi.ResetFocusEventMode)
	}
	if p.modes.IsSet(ansi.GraphemeClusteringMode) {
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
	if cancel && p.inputReader != nil {
		p.inputReader.Cancel()
		p.waitForReadLoop()
	}

	term := p.getenv("TERM")

	// Initialize the input reader.
	// This need to be done after the terminal has been initialized and set to
	// raw mode.
	// On Windows, this will change the console mode to enable mouse and window
	// events.
	var flags int
	if p.mouseMode {
		flags |= input.FlagMouseMode
	}

	drv, err := input.NewReader(p.input, term, flags)
	if err != nil {
		return fmt.Errorf("bubbletea: error initializing input reader: %w", err)
	}

	if p.traceInput {
		drv.SetLogger(log.Default())
	}
	p.inputReader = drv
	p.readLoopDone = make(chan struct{})
	go p.readLoop()

	return nil
}

func (p *Program) readInputs() error {
	for {
		events, err := p.inputReader.ReadEvents()
		if err != nil {
			return fmt.Errorf("bubbletea: error reading input: %w", err)
		}

		for _, msg := range events {
			if m := p.translateInputEvent(msg); m != nil {
				select {
				case p.msgs <- m:
				case <-p.ctx.Done():
					err := p.ctx.Err()
					if err != nil {
						err = fmt.Errorf("bubbletea: found context error while reading input: %w", err)
					}
					return err
				}
			}
		}
	}
}

func (p *Program) readLoop() {
	defer close(p.readLoopDone)

	err := p.readInputs()
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

	var resizeMsg WindowSizeMsg
	resizeMsg.Width = w
	resizeMsg.Height = h
	p.Send(resizeMsg)
}
