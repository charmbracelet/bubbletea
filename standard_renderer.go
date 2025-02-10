package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/ansi/compressor"
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

	buf                bytes.Buffer
	queuedMessageLines []string
	framerate          time.Duration
	ticker             *time.Ticker
	done               chan struct{}
	lastRender         string
	lastRenderedLines  []string
	linesRendered      int
	altLinesRendered   int
	useANSICompressor  bool
	once               sync.Once

	// cursor visibility state
	cursorHidden bool

	// essentially whether or not we're using the full size of the terminal
	altScreenActive bool

	// whether or not we're currently using bracketed paste
	bpActive bool

	// reportingFocus whether reporting focus events is enabled
	reportingFocus bool

	// renderer dimensions; usually the size of the window
	width  int
	height int

	// lines explicitly set not to render
	ignoreLines map[int]struct{}
}

// newRenderer creates a new renderer. Normally you'll want to initialize it
// with os.Stdout as the first argument.
func newRenderer(out io.Writer, useANSICompressor bool, fps int) renderer {
	if fps < 1 {
		fps = defaultFPS
	} else if fps > maxFPS {
		fps = maxFPS
	}
	r := &standardRenderer{
		out:                out,
		mtx:                &sync.Mutex{},
		done:               make(chan struct{}),
		framerate:          time.Second / time.Duration(fps),
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
	} else {
		// If the ticker already exists, it has been stopped and we need to
		// reset it.
		r.ticker.Reset(r.framerate)
	}

	// Since the renderer can be restarted after a stop, we need to reset
	// the done channel and its corresponding sync.Once.
	r.once = sync.Once{}

	go r.listen()
}

// stop permanently halts the renderer, rendering the final frame.
func (r *standardRenderer) stop() {
	// Stop the renderer before acquiring the mutex to avoid a deadlock.
	r.once.Do(func() {
		r.done <- struct{}{}
	})

	// flush locks the mutex
	r.flush()

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.EraseEntireLine)
	// Move the cursor back to the beginning of the line
	r.execute("\r")

	if r.useANSICompressor {
		if w, ok := r.out.(io.WriteCloser); ok {
			_ = w.Close()
		}
	}
}

// execute writes a sequence to the terminal.
func (r *standardRenderer) execute(seq string) {
	_, _ = io.WriteString(r.out, seq)
}

// kill halts the renderer. The final frame will not be rendered.
func (r *standardRenderer) kill() {
	// Stop the renderer before acquiring the mutex to avoid a deadlock.
	r.once.Do(func() {
		r.done <- struct{}{}
	})

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.EraseEntireLine)
	// Move the cursor back to the beginning of the line
	r.execute("\r")
}

// listen waits for ticks on the ticker, or a signal to stop the renderer.
func (r *standardRenderer) listen() {
	for {
		select {
		case <-r.done:
			r.ticker.Stop()
			return

		case <-r.ticker.C:
			r.flush()
		}
	}
}

// flush renders the buffer.
func (r *standardRenderer) flush() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.buf.Len() == 0 || r.buf.String() == r.lastRender {
		// Nothing to do.
		return
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
	for i := 0; i < len(newLines); i++ {
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
	r.lastRenderedLines = nil
}

func (r *standardRenderer) clearScreen() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.EraseEntireScreen)
	r.execute(ansi.CursorHomePosition)

	r.repaint()
}

func (r *standardRenderer) altScreen() bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.altScreenActive
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

func (r *standardRenderer) enableMouseCellMotion() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.SetButtonEventMouseMode)
}

func (r *standardRenderer) disableMouseCellMotion() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.ResetButtonEventMouseMode)
}

func (r *standardRenderer) enableMouseAllMotion() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.SetAnyEventMouseMode)
}

func (r *standardRenderer) disableMouseAllMotion() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.ResetAnyEventMouseMode)
}

func (r *standardRenderer) enableMouseSGRMode() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.SetSgrExtMouseMode)
}

func (r *standardRenderer) disableMouseSGRMode() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.ResetSgrExtMouseMode)
}

func (r *standardRenderer) enableBracketedPaste() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.SetBracketedPasteMode)
	r.bpActive = true
}

func (r *standardRenderer) disableBracketedPaste() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.ResetBracketedPasteMode)
	r.bpActive = false
}

func (r *standardRenderer) bracketedPasteActive() bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.bpActive
}

func (r *standardRenderer) enableReportFocus() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.SetFocusEventMode)
	r.reportingFocus = true
}

func (r *standardRenderer) disableReportFocus() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.execute(ansi.ResetFocusEventMode)
	r.reportingFocus = false
}

func (r *standardRenderer) reportFocus() bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	return r.reportingFocus
}

// setWindowTitle sets the terminal window title.
func (r *standardRenderer) setWindowTitle(title string) {
	r.execute(ansi.SetWindowTitle(title))
}

// setIgnoredLines specifies lines not to be touched by the standard Bubble Tea
// renderer.
func (r *standardRenderer) setIgnoredLines(from int, to int) {
	// Lock if we're going to be clearing some lines since we don't want
	// anything jacking our cursor.
	if r.lastLinesRendered() > 0 {
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
	lastLinesRendered := r.lastLinesRendered()
	if lastLinesRendered > 0 {
		buf := &bytes.Buffer{}

		for i := lastLinesRendered - 1; i >= 0; i-- {
			if _, exists := r.ignoreLines[i]; exists {
				buf.WriteString(ansi.EraseEntireLine)
			}
			buf.WriteString(ansi.CUU1)
		}
		buf.WriteString(ansi.CursorPosition(0, lastLinesRendered)) // put cursor back
		_, _ = r.out.Write(buf.Bytes())
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
//
// Deprecated: This option is deprecated and will be removed in a future
// version of this package.
func (r *standardRenderer) insertTop(lines []string, topBoundary, bottomBoundary int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	buf := &bytes.Buffer{}

	buf.WriteString(ansi.SetTopBottomMargins(topBoundary, bottomBoundary))
	buf.WriteString(ansi.CursorPosition(0, topBoundary))
	buf.WriteString(ansi.InsertLine(len(lines)))
	_, _ = buf.WriteString(strings.Join(lines, "\r\n"))
	buf.WriteString(ansi.SetTopBottomMargins(0, r.height))

	// Move cursor back to where the main rendering routine expects it to be
	buf.WriteString(ansi.CursorPosition(0, r.lastLinesRendered()))

	_, _ = r.out.Write(buf.Bytes())
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
//
// Deprecated: This option is deprecated and will be removed in a future
// version of this package.
func (r *standardRenderer) insertBottom(lines []string, topBoundary, bottomBoundary int) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	buf := &bytes.Buffer{}

	buf.WriteString(ansi.SetTopBottomMargins(topBoundary, bottomBoundary))
	buf.WriteString(ansi.CursorPosition(0, bottomBoundary))
	_, _ = buf.WriteString("\r\n" + strings.Join(lines, "\r\n"))
	buf.WriteString(ansi.SetTopBottomMargins(0, r.height))

	// Move cursor back to where the main rendering routine expects it to be
	buf.WriteString(ansi.CursorPosition(0, r.lastLinesRendered()))

	_, _ = r.out.Write(buf.Bytes())
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
		r.repaint()
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
//
// Deprecated: This option will be removed in a future version of this package.
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
//
// Deprecated: This option will be removed in a future version of this package.
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
//
// Deprecated: This option will be removed in a future version of this package.
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
//
// Deprecated: This option will be removed in a future version of this package.
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
