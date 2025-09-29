package tea

import (
	"bytes"
	"fmt"
	"time"

	"github.com/charmbracelet/x/term"
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

		if p.renderer.reportFocus() {
			p.renderer.disableReportFocus()
		}

		if p.renderer.altScreen() {
			p.renderer.exitAltScreen()

			// give the terminal a moment to catch up
			time.Sleep(time.Millisecond * 10) //nolint:mnd
		}
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

// initCancelReader (re)commences reading inputs.
func (p *Program) initCancelReader(cancel bool) error {
	if cancel && p.cancelReader != nil {
		p.cancelReader.Cancel()
		p.waitForReadLoop()
	}

	var err error
	p.cancelReader, err = newInputReader(p.input, p.mouseMode)
	if err != nil {
		return fmt.Errorf("error creating cancelreader: %w", err)
	}

	p.readLoopDone = make(chan struct{})
	readc := make(chan []byte)

	go func() {
		defer close(readc)
		p.readInput(readc)
	}()
	go p.readLoop(readc)

	return nil
}

const (
	readTimeout = 50 * time.Millisecond
	readBufSize = 4096
)

func (p *Program) readData(readc chan<- []byte) error {
	for {
		var readBuf [readBufSize]byte
		n, err := p.cancelReader.Read(readBuf[:])
		if err != nil {
			return err //nolint:wrapcheck
		}

		select {
		case <-p.ctx.Done():
			return nil
		case readc <- readBuf[:n]:
		}
	}
}

func (p *Program) readLoop(readc chan []byte) {
	defer close(p.readLoopDone)

	var buf bytes.Buffer
	timer := time.NewTimer(readTimeout)
	expires := time.Now().Add(readTimeout)

	for {
		select {
		case <-p.ctx.Done():
			scanInput(buf.Bytes(), true, p.msgs)
			return
		case <-timer.C:
			timedout := time.Now().After(expires)
			if buf.Len() > 0 && timedout {
				buf.Next(scanInput(buf.Bytes(), timedout, p.msgs))
			}
			if buf.Len() > 0 {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}

				timer.Reset(readTimeout)
			}
		case data, ok := <-readc:
			if !ok {
				scanInput(buf.Bytes(), true, p.msgs)
				return
			}

			buf.Write(data)
			expires = time.Now().Add(readTimeout)
			n := scanInput(buf.Bytes(), false, p.msgs)
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			if n > 0 {
				buf.Next(n)
			}
			if buf.Len() > 0 {
				timer.Reset(readTimeout)
			}
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

	p.Send(WindowSizeMsg{
		Width:  w,
		Height: h,
	})
}
