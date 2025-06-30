package tea

import (
	"bytes"
	"image/color"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// standardRenderer is a framerate-based terminal renderer, updating the view
// at a given framerate to avoid overloading the terminal emulator.
//
// In cases where very high performance is needed the renderer can be told
// to exclude ranges of lines, allowing them to be written to directly.
type standardRenderer struct {
	mtx *sync.Mutex
	out *colorprofile.Writer

	buf                bytes.Buffer
	queuedMessageLines []string
	done               chan struct{}
	lastRender         string
	lastRenderedLines  []string
	linesRendered      int
	altLinesRendered   int

	// cursor visibility state
	cursorHidden bool

	// essentially whether or not we're using the full size of the terminal
	altScreenActive bool

	// renderer dimensions; usually the size of the window
	width  int
	height int

	// lines explicitly set not to render
	ignoreLines map[int]struct{}
}

// newRenderer creates a new renderer. Normally you'll want to initialize it
// with os.Stdout as the first argument.
func newRenderer(out io.Writer) renderer {
	r := &standardRenderer{
		out: &colorprofile.Writer{
			Forward: out,
		},
		mtx:                &sync.Mutex{},
		done:               make(chan struct{}),
		queuedMessageLines: []string{},
	}
	return r
}

// setColorProfile sets the color profile.
func (r *standardRenderer) setColorProfile(p colorprofile.Profile) {
	r.mtx.Lock()
	r.out.Profile = p
	r.mtx.Unlock()
}

// reset resets the renderer to its initial state.
func (r *standardRenderer) reset() {
	// no-op
}

// close closes the renderer and flushes any remaining data.
func (r *standardRenderer) close() error {
	// flush locks the mutex
	_ = r.flush(nil)

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.EraseEntireLine)
	// Move the cursor back to the beginning of the line
	r.execute("\r")

	return nil
}

// execute writes a sequence to the terminal.
func (r *standardRenderer) execute(seq string) {
	_, _ = io.WriteString(r.out, seq)
}

// writeString writes a string to the internal buffer.
func (r *standardRenderer) writeString(s string) (int, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.buf.WriteString(s)
}

// flush renders the buffer.
func (r *standardRenderer) flush(*Program) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.buf.Len() == 0 || r.buf.String() == r.lastRender {
		// Nothing to do.
		return nil
	}

	// Output buffer.
	buf := &bytes.Buffer{}

	// Moving to the beginning of the section, that we rendered.
	if r.altScreenActive {
		buf.WriteString(ansi.CursorHomePosition)
	} else if r.linesRendered > 1 {
		buf.WriteString(ansi.CursorUp(r.linesRendered - 1))
	}

	newLines := strings.Split(r.buf.String(), "\n")

	// If we know the output's height, we can use it to determine how many
	// lines we can render. We drop lines from the top of the render buffer if
	// necessary, as we can't navigate the cursor into the terminal's scrollback
	// buffer.
	if r.height > 0 && len(newLines) > r.height {
		newLines = newLines[len(newLines)-r.height:]
	}

	flushQueuedMessages := len(r.queuedMessageLines) > 0 && !r.altScreenActive

	if flushQueuedMessages {
		// Dump the lines we've queued up for printing.
		for _, line := range r.queuedMessageLines {
			if ansi.StringWidth(line) < r.width {
				// We only erase the rest of the line when the line is shorter than
				// the width of the terminal. When the cursor reaches the end of
				// the line, any escape sequences that follow will only affect the
				// last cell of the line.

				// Removing previously rendered content at the end of line.
				line = line + ansi.EraseLineRight
			}

			_, _ = buf.WriteString(line)
			_, _ = buf.WriteString("\r\n")
		}
		// Clear the queued message lines.
		r.queuedMessageLines = []string{}
	}

	// Paint new lines.
	for i := range newLines {
		canSkip := !flushQueuedMessages && // Queuing messages triggers repaint -> we don't have access to previous frame content.
			len(r.lastRenderedLines) > i && r.lastRenderedLines[i] == newLines[i] // Previously rendered line is the same.

		if _, ignore := r.ignoreLines[i]; ignore || canSkip {
			// Unless this is the last line, move the cursor down.
			if i < len(newLines)-1 {
				buf.WriteByte('\n')
			}
			continue
		}

		if i == 0 && r.lastRender == "" {
			// On first render, reset the cursor to the start of the line
			// before writing anything.
			buf.WriteByte('\r')
		}

		line := newLines[i]

		// Truncate lines wider than the width of the window to avoid
		// wrapping, which will mess up rendering. If we don't have the
		// width of the window this will be ignored.
		//
		// Note that on Windows we only get the width of the window on
		// program initialization, so after a resize this won't perform
		// correctly (signal SIGWINCH is not supported on Windows).
		if r.width > 0 {
			line = ansi.Truncate(line, r.width, "")
		}

		if ansi.StringWidth(line) < r.width {
			// We only erase the rest of the line when the line is shorter than
			// the width of the terminal. When the cursor reaches the end of
			// the line, any escape sequences that follow will only affect the
			// last cell of the line.

			// Removing previously rendered content at the end of line.
			line = line + ansi.EraseLineRight
		}

		_, _ = buf.WriteString(line)

		if i < len(newLines)-1 {
			_, _ = buf.WriteString("\r\n")
		}
	}

	// Clearing left over content from last render.
	if r.lastLinesRendered() > len(newLines) {
		buf.WriteString(ansi.EraseScreenBelow)
	}

	if r.altScreenActive {
		r.altLinesRendered = len(newLines)
	} else {
		r.linesRendered = len(newLines)
	}

	// Make sure the cursor is at the start of the last line to keep rendering
	// behavior consistent.
	if r.altScreenActive {
		// This case fixes a bug in macOS terminal. In other terminals the
		// other case seems to do the job regardless of whether or not we're
		// using the full terminal window.
		buf.WriteString(ansi.CursorPosition(0, len(newLines)))
	} else {
		buf.WriteString(ansi.CursorBackward(r.width))
	}

	_, _ = r.out.Write(buf.Bytes())
	r.lastRender = r.buf.String()

	// Save previously rendered lines for comparison in the next render. If we
	// don't do this, we can't skip rendering lines that haven't changed.
	// See https://github.com/charmbracelet/bubbletea/pull/1233
	r.lastRenderedLines = newLines
	r.buf.Reset()

	return nil
}

// lastLinesRendered returns the number of lines rendered lastly.
func (r *standardRenderer) lastLinesRendered() int {
	if r.altScreenActive {
		return r.altLinesRendered
	}
	return r.linesRendered
}

// write writes to the internal buffer. The buffer will be outputted via the
// ticker which calls flush().
func (r *standardRenderer) render(v View) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.buf.Reset()

	area := uv.Rect(0, 0, r.width, r.height)
	if b, ok := v.Layer.(interface{ Bounds() uv.Rectangle }); ok {
		if !r.altScreenActive {
			area.Max.Y = b.Bounds().Max.Y
		}
	}

	buf := uv.NewScreenBuffer(area.Dx(), area.Dy())
	v.Layer.Draw(buf, area)
	s := buf.Render()

	// If an empty string was passed we should clear existing output and
	// rendering nothing. Rather than introduce additional state to manage
	// this, we render a single space as a simple (albeit less correct)
	// solution.
	if s == "" {
		s = " "
	}

	_, _ = r.buf.WriteString(s)
}

func (r *standardRenderer) repaint() {
	r.lastRender = ""
	r.lastRenderedLines = nil
}

func (r *standardRenderer) clearScreen() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.EraseEntireScreen)
	r.execute(ansi.CursorHomePosition)

	r.repaint()
}

func (r *standardRenderer) enterAltScreen() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.altScreenActive {
		return
	}

	r.altScreenActive = true
	r.execute(ansi.SetAltScreenSaveCursorMode)

	// Ensure that the terminal is cleared, even when it doesn't support
	// alt screen (or alt screen support is disabled, like GNU screen by
	// default).
	//
	// Note: we can't use r.clearScreen() here because the mutex is already
	// locked.
	r.execute(ansi.EraseEntireScreen)
	r.execute(ansi.CursorHomePosition)

	// cmd.exe and other terminals keep separate cursor states for the AltScreen
	// and the main buffer. We have to explicitly reset the cursor visibility
	// whenever we enter AltScreen.
	if r.cursorHidden {
		r.execute(ansi.HideCursor)
	} else {
		r.execute(ansi.ShowCursor)
	}

	// Entering the alt screen resets the lines rendered count.
	r.altLinesRendered = 0

	r.repaint()
}

func (r *standardRenderer) exitAltScreen() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if !r.altScreenActive {
		return
	}

	r.altScreenActive = false
	r.execute(ansi.ResetAltScreenSaveCursorMode)

	// cmd.exe and other terminals keep separate cursor states for the AltScreen
	// and the main buffer. We have to explicitly reset the cursor visibility
	// whenever we exit AltScreen.
	if r.cursorHidden {
		r.execute(ansi.HideCursor)
	} else {
		r.execute(ansi.ShowCursor)
	}

	r.repaint()
}

func (r *standardRenderer) showCursor() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.cursorHidden = false
	r.execute(ansi.ShowCursor)
}

func (r *standardRenderer) hideCursor() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.cursorHidden = true
	r.execute(ansi.HideCursor)
}

func (r *standardRenderer) resize(width, height int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.width = width
	r.height = height
	r.repaint()
}

func (r *standardRenderer) insertAbove(s string) {
	if !r.altScreenActive {
		lines := strings.Split(s, "\n")
		r.mtx.Lock()
		r.queuedMessageLines = append(r.queuedMessageLines, lines...)
		r.repaint()
		r.mtx.Unlock()
	}
}

func (r *standardRenderer) setCursorColor(c color.Color)     {}
func (r *standardRenderer) setForegroundColor(c color.Color) {}
func (r *standardRenderer) setBackgroundColor(c color.Color) {}
func (r *standardRenderer) setWindowTitle(s string)          {}
func (r *standardRenderer) hit(MouseMsg) []Msg               { return nil }
