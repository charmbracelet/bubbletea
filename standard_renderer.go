package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

const (
	// defaultFramerate specifies the maximum interval at which we should
	// update the view.
	defaultFPS = 60
	maxFPS     = 120
)

// standardRenderer is a framerate-based terminal renderer, updating the view
// at a given framerate to avoid overloading the terminal emulator.
//
// In cases where very high performance is needed the renderer can be told
// to exclude ranges of lines, allowing them to be written to directly.
type standardRenderer struct {
	mtx *sync.Mutex
	out io.Writer

	// the color profile to use
	profile colorprofile.Profile

	buf                bytes.Buffer
	queuedMessageLines []string
	lastRender         string
	linesRendered      int

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

// newStandardRenderer creates a new renderer. Normally you'll want to initialize it
// with os.Stdout as the first argument.
func newStandardRenderer(p colorprofile.Profile) renderer {
	r := &standardRenderer{
		mtx:                &sync.Mutex{},
		queuedMessageLines: []string{},
		profile:            p,
	}
	return r
}

// setOutput sets the output for the renderer.
func (r *standardRenderer) setOutput(out io.Writer) {
	r.mtx.Lock()
	r.out = &colorprofile.Writer{
		Forward: out,
		Profile: r.profile,
	}
	r.mtx.Unlock()
}

// close closes the renderer and flushes any remaining data.
func (r *standardRenderer) close() (err error) {
	// Move the cursor back to the beginning of the line
	// NOTE: execute locks the mutex
	r.execute(ansi.EraseEntireLine + "\r")

	return
}

// execute writes the given sequence to the output.
func (r *standardRenderer) execute(seq string) {
	r.mtx.Lock()
	_, _ = io.WriteString(r.out, seq)
	r.mtx.Unlock()
}

// flush renders the buffer.
func (r *standardRenderer) flush() (err error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.buf.Len() == 0 || r.buf.String() == r.lastRender {
		// Nothing to do
		return nil
	}

	// Output buffer
	buf := &bytes.Buffer{}

	newLines := strings.Split(r.buf.String(), "\n")

	// If we know the output's height, we can use it to determine how many
	// lines we can render. We drop lines from the top of the render buffer if
	// necessary, as we can't navigate the cursor into the terminal's scrollback
	// buffer.
	if r.height > 0 && len(newLines) > r.height {
		newLines = newLines[len(newLines)-r.height:]
	}

	numLinesThisFlush := len(newLines)
	oldLines := strings.Split(r.lastRender, "\n")
	skipLines := make(map[int]struct{})
	flushQueuedMessages := len(r.queuedMessageLines) > 0 && !r.altScreenActive

	// Clear any lines we painted in the last render.
	if r.linesRendered > 0 {
		for i := r.linesRendered - 1; i > 0; i-- {
			// if we are clearing queued messages, we want to clear all lines, since
			// printing messages allows for native terminal word-wrap, we
			// don't have control over the queued lines
			if flushQueuedMessages {
				buf.WriteString(ansi.EraseEntireLine)
			} else if (len(newLines) <= len(oldLines)) && (len(newLines) > i && len(oldLines) > i) && (newLines[i] == oldLines[i]) {
				// If the number of lines we want to render hasn't increased and
				// new line is the same as the old line we can skip rendering for
				// this line as a performance optimization.
				skipLines[i] = struct{}{}
			} else if _, exists := r.ignoreLines[i]; !exists {
				buf.WriteString(ansi.EraseEntireLine)
			}

			buf.WriteString(ansi.CursorUp1)
		}

		if _, exists := r.ignoreLines[0]; !exists {
			// We need to return to the start of the line here to properly
			// erase it. Going back the entire width of the terminal will
			// usually be farther than we need to go, but terminal emulators
			// will stop the cursor at the start of the line as a rule.
			//
			// We use this sequence in particular because it's part of the ANSI
			// standard (whereas others are proprietary to, say, VT100/VT52).
			// If cursor previous line (ESC[ + <n> + F) were better supported
			// we could use that above to eliminate this step.
			buf.WriteString(ansi.CursorLeft(r.width))
			buf.WriteString(ansi.EraseEntireLine)
		}
	}

	// Merge the set of lines we're skipping as a rendering optimization with
	// the set of lines we've explicitly asked the renderer to ignore.
	for k, v := range r.ignoreLines {
		skipLines[k] = v
	}

	if flushQueuedMessages {
		// Dump the lines we've queued up for printing
		for _, line := range r.queuedMessageLines {
			_, _ = buf.WriteString(line)
			_, _ = buf.WriteString("\r\n")
		}
		// clear the queued message lines
		r.queuedMessageLines = []string{}
	}

	// Paint new lines
	for i := 0; i < len(newLines); i++ {
		if _, skip := skipLines[i]; skip {
			// Unless this is the last line, move the cursor down.
			if i < len(newLines)-1 {
				buf.WriteString(ansi.CursorDown1)
			}
		} else {
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

			_, _ = buf.WriteString(line)

			if i < len(newLines)-1 {
				_, _ = buf.WriteString("\r\n")
			}
		}
	}
	r.linesRendered = numLinesThisFlush

	// Make sure the cursor is at the start of the last line to keep rendering
	// behavior consistent.
	if r.altScreenActive {
		// This case fixes a bug in macOS terminal. In other terminals the
		// other case seems to do the job regardless of whether or not we're
		// using the full terminal window.
		buf.WriteString(ansi.SetCursorPosition(0, r.linesRendered))
	} else {
		buf.WriteString(ansi.CursorLeft(r.width))
	}

	_, err = r.out.Write(buf.Bytes())
	r.lastRender = r.buf.String()
	r.buf.Reset()
	return
}

// render renders the frame to the internal buffer. The buffer will be
// outputted via the ticker which calls flush().
func (r *standardRenderer) render(s string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.buf.Reset()

	// If an empty string was passed we should clear existing output and
	// rendering nothing. Rather than introduce additional state to manage
	// this, we render a single space as a simple (albeit less correct)
	// solution.
	if s == "" {
		s = " "
	}

	_, _ = r.buf.WriteString(s)
}

// repaint forces a full repaint.
func (r *standardRenderer) repaint() {
	r.lastRender = ""
}

// reset resets the standardRenderer to its initial state.
func (r *standardRenderer) reset() {
	r.repaint()
}

func (r *standardRenderer) clearScreen() {
	r.execute(ansi.EraseEntireScreen + ansi.CursorOrigin)

	r.repaint()
}

// update handles internal messages for the renderer.
func (r *standardRenderer) update(msg Msg) {
	switch msg := msg.(type) {
	case enableModeMsg:
		switch string(msg) {
		case ansi.AltScreenBufferMode.String():
			if r.altScreenActive {
				return
			}

			r.altScreenActive = true
			r.repaint()
		case ansi.CursorEnableMode.String():
			if !r.cursorHidden {
				return
			}

			r.cursorHidden = false
		}

	case disableModeMsg:
		switch string(msg) {
		case ansi.AltScreenBufferMode.String():
			if !r.altScreenActive {
				return
			}

			r.altScreenActive = false
			r.repaint()
		case ansi.CursorEnableMode.String():
			if r.cursorHidden {
				return
			}

			r.cursorHidden = true
		}

	case rendererWriter:
		r.setOutput(msg.Writer)

	case WindowSizeMsg:
		r.resize(msg.Width, msg.Height)

	case clearScreenMsg:
		r.clearScreen()

	case printLineMessage:
		r.insertAbove(msg.messageBody)

	case repaintMsg:
		// Force a repaint by clearing the render cache as we slide into a
		// render.
		r.mtx.Lock()
		r.repaint()
		r.mtx.Unlock()
	}
}

// resize sets the size of the terminal.
func (r *standardRenderer) resize(w int, h int) {
	r.mtx.Lock()
	r.width = w
	r.height = h
	r.repaint()
	r.mtx.Unlock()
}

// insertAbove inserts lines above the current frame. This only works in
// inline mode.
func (r *standardRenderer) insertAbove(s string) {
	if r.altScreenActive {
		return
	}

	lines := strings.Split(s, "\n")
	r.mtx.Lock()
	r.queuedMessageLines = append(r.queuedMessageLines, lines...)
	r.repaint()
	r.mtx.Unlock()
}

type printLineMessage struct {
	messageBody string
}

// Println prints above the Program. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Println (but similar to log.Println) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func Println(args ...interface{}) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprint(args...),
		}
	}
}

// Printf prints above the Program. It takes a format template followed by
// values similar to fmt.Printf. This output is unmanaged by the program and
// will persist across renders by the Program.
//
// Unlike fmt.Printf (but similar to log.Printf) the message will be print on
// its own line.
//
// If the altscreen is active no output will be printed.
func Printf(template string, args ...interface{}) Cmd {
	return func() Msg {
		return printLineMessage{
			messageBody: fmt.Sprintf(template, args...),
		}
	}
}
