package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/muesli/ansi/compressor"
	"github.com/muesli/reflow/truncate"
)

const (
	// defaultFramerate specifies the maximum interval at which we should
	// update the view.
	defaultFramerate = time.Second / 60
)

// standardRenderer is a framerate-based terminal renderer, updating the view
// at a given framerate to avoid overloading the terminal emulator.
//
// In cases where very high performance is needed the renderer can be told
// to exclude ranges of lines, allowing them to be written to directly.
type standardRenderer struct {
	out                io.Writer
	buf                bytes.Buffer
	queuedMessageLines []string
	framerate          time.Duration
	ticker             *time.Ticker
	mtx                *sync.Mutex
	done               chan struct{}
	lastRender         string
	linesRendered      int
	useANSICompressor  bool
	once               sync.Once

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
func newRenderer(out io.Writer, mtx *sync.Mutex, useANSICompressor bool) renderer {
	r := &standardRenderer{
		out:                out,
		mtx:                mtx,
		framerate:          defaultFramerate,
		useANSICompressor:  useANSICompressor,
		queuedMessageLines: []string{},
	}
	if r.useANSICompressor {
		r.out = &compressor.Writer{Forward: out}
	}
	return r
}

// start starts the renderer.
func (r *standardRenderer) start() {
	if r.ticker == nil {
		r.ticker = time.NewTicker(r.framerate)
	}
	r.done = make(chan struct{})
	go r.listen()
}

// stop permanently halts the renderer, rendering the final frame.
func (r *standardRenderer) stop() {
	r.flush()
	clearLine(r.out)
	r.once.Do(func() {
		close(r.done)
	})

	if r.useANSICompressor {
		if w, ok := r.out.(io.WriteCloser); ok {
			_ = w.Close()
		}
	}
}

// kill halts the renderer. The final frame will not be rendered.
func (r *standardRenderer) kill() {
	clearLine(r.out)
	r.once.Do(func() {
		close(r.done)
	})
}

// listen waits for ticks on the ticker, or a signal to stop the renderer.
func (r *standardRenderer) listen() {
	for {
		select {
		case <-r.ticker.C:
			if r.ticker != nil {
				r.flush()
			}
		case <-r.done:
			r.ticker.Stop()
			r.ticker = nil
			return
		}
	}
}

// flush renders the buffer.
func (r *standardRenderer) flush() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.buf.Len() == 0 || r.buf.String() == r.lastRender {
		// Nothing to do
		return
	}

	// Output buffer
	out := new(bytes.Buffer)

	newLines := strings.Split(r.buf.String(), "\n")
	numLinesThisFlush := len(newLines)
	oldLines := strings.Split(r.lastRender, "\n")
	skipLines := make(map[int]struct{})
	flushQueuedMessages := len(r.queuedMessageLines) > 0 && !r.altScreenActive

	// Add any queued messages to this render
	if flushQueuedMessages {
		newLines = append(r.queuedMessageLines, newLines...)
		r.queuedMessageLines = []string{}
	}

	// Clear any lines we painted in the last render.
	if r.linesRendered > 0 {
		for i := r.linesRendered - 1; i > 0; i-- {
			// If the number of lines we want to render hasn't increased and
			// new line is the same as the old line we can skip rendering for
			// this line as a performance optimization.
			if (len(newLines) <= len(oldLines)) && (len(newLines) > i && len(oldLines) > i) && (newLines[i] == oldLines[i]) {
				skipLines[i] = struct{}{}
			} else if _, exists := r.ignoreLines[i]; !exists {
				clearLine(out)
			}

			cursorUp(out)
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
			cursorBack(out, r.width)
			clearLine(out)
		}
	}

	// Merge the set of lines we're skipping as a rendering optimization with
	// the set of lines we've explicitly asked the renderer to ignore.
	if r.ignoreLines != nil {
		for k, v := range r.ignoreLines {
			skipLines[k] = v
		}
	}

	// Paint new lines
	for i := 0; i < len(newLines); i++ {
		if _, skip := skipLines[i]; skip {
			// Unless this is the last line, move the cursor down.
			if i < len(newLines)-1 {
				cursorDown(out)
			}
		} else {
			line := newLines[i]

			// Truncate lines wider than the width of the window to avoid
			// wrapping, which will mess up rendering. If we don't have the
			// width of the window this will be ignored.
			//
			// Note that on Windows we only get the width of the window on
			// program initialization, so after a resize this won't perform
			// correctly (signal SIGWINCH is not supported on Windows).
			if r.width > 0 {
				line = truncate.String(line, uint(r.width))
			}

			_, _ = io.WriteString(out, line)

			if i < len(newLines)-1 {
				_, _ = io.WriteString(out, "\r\n")
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
		moveCursor(out, r.linesRendered, 0)
	} else {
		cursorBack(out, r.width)
	}

	_, _ = r.out.Write(out.Bytes())
	r.lastRender = r.buf.String()
	r.buf.Reset()
}

// write writes to the internal buffer. The buffer will be outputted via the
// ticker which calls flush().
func (r *standardRenderer) write(s string) {
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

func (r *standardRenderer) repaint() {
	r.lastRender = ""
}

func (r *standardRenderer) altScreen() bool {
	return r.altScreenActive
}

func (r *standardRenderer) setAltScreen(v bool) {
	r.altScreenActive = v
	r.repaint()
}

// setIgnoredLines specifies lines not to be touched by the standard Bubble Tea
// renderer.
func (r *standardRenderer) setIgnoredLines(from int, to int) {
	// Lock if we're going to be clearing some lines since we don't want
	// anything jacking our cursor.
	if r.linesRendered > 0 {
		r.mtx.Lock()
		defer r.mtx.Unlock()
	}

	if r.ignoreLines == nil {
		r.ignoreLines = make(map[int]struct{})
	}
	for i := from; i < to; i++ {
		r.ignoreLines[i] = struct{}{}
	}

	// Erase ignored lines
	if r.linesRendered > 0 {
		out := new(bytes.Buffer)
		for i := r.linesRendered - 1; i >= 0; i-- {
			if _, exists := r.ignoreLines[i]; exists {
				clearLine(out)
			}
			cursorUp(out)
		}
		moveCursor(out, r.linesRendered, 0) // put cursor back
		_, _ = r.out.Write(out.Bytes())
	}
}

// clearIgnoredLines returns control of any ignored lines to the standard
// Bubble Tea renderer. That is, any lines previously set to be ignored can be
// rendered to again.
func (r *standardRenderer) clearIgnoredLines() {
	r.ignoreLines = nil
}

// insertTop effectively scrolls up. It inserts lines at the top of a given
// area designated to be a scrollable region, pushing everything else down.
// This is roughly how ncurses does it.
//
// To call this function use command ScrollUp().
//
// For this to work renderer.ignoreLines must be set to ignore the scrollable
// region since we are bypassing the normal Bubble Tea renderer here.
//
// Because this method relies on the terminal dimensions, it's only valid for
// full-window applications (generally those that use the alternate screen
// buffer).
//
// This method bypasses the normal rendering buffer and is philosophically
// different than the normal way we approach rendering in Bubble Tea. It's for
// use in high-performance rendering, such as a pager that could potentially
// be rendering very complicated ansi. In cases where the content is simpler
// standard Bubble Tea rendering should suffice.
func (r *standardRenderer) insertTop(lines []string, topBoundary, bottomBoundary int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	b := new(bytes.Buffer)

	changeScrollingRegion(b, topBoundary, bottomBoundary)
	moveCursor(b, topBoundary, 0)
	insertLine(b, len(lines))
	_, _ = io.WriteString(b, strings.Join(lines, "\r\n"))
	changeScrollingRegion(b, 0, r.height)

	// Move cursor back to where the main rendering routine expects it to be
	moveCursor(b, r.linesRendered, 0)

	_, _ = r.out.Write(b.Bytes())
}

// insertBottom effectively scrolls down. It inserts lines at the bottom of
// a given area designated to be a scrollable region, pushing everything else
// up. This is roughly how ncurses does it.
//
// To call this function use the command ScrollDown().
//
// See note in insertTop() for caveats, how this function only makes sense for
// full-window applications, and how it differs from the normal way we do
// rendering in Bubble Tea.
func (r *standardRenderer) insertBottom(lines []string, topBoundary, bottomBoundary int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	b := new(bytes.Buffer)

	changeScrollingRegion(b, topBoundary, bottomBoundary)
	moveCursor(b, bottomBoundary, 0)
	_, _ = io.WriteString(b, "\r\n"+strings.Join(lines, "\r\n"))
	changeScrollingRegion(b, 0, r.height)

	// Move cursor back to where the main rendering routine expects it to be
	moveCursor(b, r.linesRendered, 0)

	_, _ = r.out.Write(b.Bytes())
}

// handleMessages handles internal messages for the renderer.
func (r *standardRenderer) handleMessages(msg Msg) {
	switch msg := msg.(type) {
	case repaintMsg:
		// Force a repaint by clearing the render cache as we slide into a
		// render.
		r.mtx.Lock()
		r.repaint()
		r.mtx.Unlock()

	case WindowSizeMsg:
		r.mtx.Lock()
		r.width = msg.Width
		r.height = msg.Height
		r.mtx.Unlock()

	case clearScrollAreaMsg:
		r.clearIgnoredLines()

		// Force a repaint on the area where the scrollable stuff was in this
		// update cycle
		r.mtx.Lock()
		r.repaint()
		r.mtx.Unlock()

	case syncScrollAreaMsg:
		// Re-render scrolling area
		r.clearIgnoredLines()
		r.setIgnoredLines(msg.topBoundary, msg.bottomBoundary)
		r.insertTop(msg.lines, msg.topBoundary, msg.bottomBoundary)

		// Force non-scrolling stuff to repaint in this update cycle
		r.mtx.Lock()
		r.repaint()
		r.mtx.Unlock()

	case scrollUpMsg:
		r.insertTop(msg.lines, msg.topBoundary, msg.bottomBoundary)

	case scrollDownMsg:
		r.insertBottom(msg.lines, msg.topBoundary, msg.bottomBoundary)

	case printLineMessage:
		if !r.altScreenActive {
			lines := strings.Split(msg.messageBody, "\n")
			r.mtx.Lock()
			r.queuedMessageLines = append(r.queuedMessageLines, lines...)
			r.repaint()
			r.mtx.Unlock()
		}
	}
}

// HIGH-PERFORMANCE RENDERING STUFF

type syncScrollAreaMsg struct {
	lines          []string
	topBoundary    int
	bottomBoundary int
}

// SyncScrollArea performs a paint of the entire region designated to be the
// scrollable area. This is required to initialize the scrollable region and
// should also be called on resize (WindowSizeMsg).
//
// For high-performance, scroll-based rendering only.
func SyncScrollArea(lines []string, topBoundary int, bottomBoundary int) Cmd {
	return func() Msg {
		return syncScrollAreaMsg{
			lines:          lines,
			topBoundary:    topBoundary,
			bottomBoundary: bottomBoundary,
		}
	}
}

type clearScrollAreaMsg struct{}

// ClearScrollArea deallocates the scrollable region and returns the control of
// those lines to the main rendering routine.
//
// For high-performance, scroll-based rendering only.
func ClearScrollArea() Msg {
	return clearScrollAreaMsg{}
}

type scrollUpMsg struct {
	lines          []string
	topBoundary    int
	bottomBoundary int
}

// ScrollUp adds lines to the top of the scrollable region, pushing existing
// lines below down. Lines that are pushed out the scrollable region disappear
// from view.
//
// For high-performance, scroll-based rendering only.
func ScrollUp(newLines []string, topBoundary, bottomBoundary int) Cmd {
	return func() Msg {
		return scrollUpMsg{
			lines:          newLines,
			topBoundary:    topBoundary,
			bottomBoundary: bottomBoundary,
		}
	}
}

type scrollDownMsg struct {
	lines          []string
	topBoundary    int
	bottomBoundary int
}

// ScrollDown adds lines to the bottom of the scrollable region, pushing
// existing lines above up. Lines that are pushed out of the scrollable region
// disappear from view.
//
// For high-performance, scroll-based rendering only.
func ScrollDown(newLines []string, topBoundary, bottomBoundary int) Cmd {
	return func() Msg {
		return scrollDownMsg{
			lines:          newLines,
			topBoundary:    topBoundary,
			bottomBoundary: bottomBoundary,
		}
	}
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
