package tea

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
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
	p.keepWriteErr(p.flush())

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

	// Close the reader we are about to replace, otherwise its resources (e.g.
	// the epoll file descriptor on Linux) leak on every ReleaseTerminal and
	// RestoreTerminal cycle. Only close it once its read loop has exited: if
	// waitForReadLoop timed out, the loop may still be using it.
	if p.cancelReader != nil && p.readLoopDone != nil {
		select {
		case <-p.readLoopDone:
			_ = p.cancelReader.Close()
		default:
		}
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

	go p.readLoop(p.ctx, drv, p.readLoopDone, p.errs)

	return nil
}

// readLoop streams input events to the program until ctx is done or reading
// fails. It takes the per-run state it uses as arguments rather than reading
// it from p: the loop can outlive [Program.shutdown], see
// [Program.shutdownInput], and must not race with a later run replacing it.
func (p *Program) readLoop(ctx context.Context, scanner *uv.TerminalReader, done chan struct{}, errs chan error) {
	defer close(done)

	if err := scanner.StreamEvents(ctx, p.msgs); err != nil {
		select {
		case <-ctx.Done():
			return
		case errs <- err:
		}
	}
}

const inputReadLoopTimeout = 500 * time.Millisecond

// shutdownInput cancels input before closing the reader's resources. Keep the
// bounded shutdown for slow readers, but retain ownership of a reader that is
// still in use until its input loop has actually finished.
func (p *Program) shutdownInput() {
	reader := p.cancelReader
	if reader == nil {
		return
	}
	// The reader is closed below, or later once its input loop finishes.
	// Forget it so that a later run doesn't close it again: closing it closes
	// raw file descriptors, and closing one twice can close an unrelated file
	// that has since reused the number.
	p.cancelReader = nil

	if reader.Cancel() && p.readLoopDone != nil {
		done := p.readLoopDone
		timer := time.NewTimer(inputReadLoopTimeout)
		defer timer.Stop()

		select {
		case <-done:
		case <-timer.C:
			go func() {
				<-done
				_ = reader.Close()
			}()
			return
		}
	}

	_ = reader.Close()
}

// waitForReadLoop waits for the cancelReader to finish its read loop.
func (p *Program) waitForReadLoop() {
	select {
	case <-p.readLoopDone:
	case <-time.After(inputReadLoopTimeout):
		// The read loop hangs, which means the input
		// cancelReader's cancel function has returned true even
		// though it was not able to cancel the read.
	}
}

// queryDrainTimeout bounds how long [Program.drainQueryReplies] waits for the
// terminal to answer.
const queryDrainTimeout = 500 * time.Millisecond

// queryReplyWindow is how long after a query its reply may still be on its
// way. The input reader reads replies that arrive later than that.
const queryReplyWindow = time.Second

// drainQueryReplies reads the replies to terminal queries that may still be in
// flight after the input reader has stopped, and returns them. Otherwise they
// stay in the tty input queue for the next reader: a process started with
// [ExecProcess], or the shell once the program exits (#1590).
//
// It sends a Primary Device Attributes (DA1) request and reads input until
// the terminal answers it. Terminals answer requests in order, so by then
// every earlier reply has been read. Nothing is drained when the last query
// was sent long enough ago for its reply to have been read already.
func (p *Program) drainQueryReplies(ctx context.Context) []Msg {
	last := p.lastQuery.Swap(0)
	if last == 0 || time.Since(time.Unix(0, last)) > queryReplyWindow ||
		p.ttyInput == nil || p.ttyOutput == nil || p.input == nil {
		return nil
	}

	// Make sure the pending queries themselves have been written, followed by
	// the DA1 request that marks the end of their replies.
	p.mu.Lock()
	_, _ = p.outputBuf.WriteString(ansi.RequestPrimaryDeviceAttributes)
	p.mu.Unlock()
	if err := p.flush(); err != nil {
		p.keepWriteErr(err)
		return nil
	}

	cr, err := uv.NewCancelReader(p.input)
	if err != nil {
		return nil
	}
	defer cr.Close() //nolint:errcheck

	// Stop reading once the DA1 reply arrives, the timeout expires, or ctx is
	// done.
	ctx, cancel := context.WithTimeout(ctx, queryDrainTimeout)
	defer cancel()
	context.AfterFunc(ctx, func() { cr.Cancel() })

	drv := uv.NewTerminalReader(cr, p.environ.Getenv("TERM"))
	drv.SetLogger(p.logger)

	eventc := make(chan uv.Event)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = drv.StreamEvents(ctx, eventc)
	}()

	var events []Msg
	for {
		select {
		case ev := <-eventc:
			if _, ok := ev.(uv.PrimaryDeviceAttributesEvent); ok {
				// Every earlier reply has been read.
				cancel()
				continue
			}
			events = append(events, ev)
		case <-done:
			return events
		}
	}
}

const (
	defaultWidth  = 80
	defaultHeight = 24
)

// fallbackDimensions returns width and height, falling back to COLUMNS/LINES
// environment variables or 80x24 defaults when dimensions are non-positive.
func (p *Program) fallbackDimensions(width, height int) (int, int) {
	if width <= 0 {
		var cols string
		if p != nil && p.environ != nil {
			cols = p.environ.Getenv("COLUMNS")
		} else {
			cols = os.Getenv("COLUMNS")
		}
		if c, err := strconv.Atoi(cols); err == nil && c > 0 {
			width = c
		} else {
			width = defaultWidth
		}
	}
	if height <= 0 {
		var lines string
		if p != nil && p.environ != nil {
			lines = p.environ.Getenv("LINES")
		} else {
			lines = os.Getenv("LINES")
		}
		if l, err := strconv.Atoi(lines); err == nil && l > 0 {
			height = l
		} else {
			height = defaultHeight
		}
	}
	return width, height
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

	w, h = p.fallbackDimensions(w, h)

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
