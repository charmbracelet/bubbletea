package tea

import (
	"fmt"
	"io"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/x/ansi"
)

type cursedRenderer struct {
	w             io.Writer
	scr           *tv.TerminalRenderer
	buf           *tv.Buffer
	lastFrame     *string
	lastCur       *Cursor
	env           []string
	term          string // the terminal type $TERM
	width, height int
	mu            sync.Mutex
	profile       colorprofile.Profile
	cursor        Cursor
	method        ansi.Method
	logger        tv.Logger
	altScreen     bool
	cursorHidden  bool
	hardTabs      bool // whether to use hard tabs to optimize cursor movements
	backspace     bool // whether to use backspace to optimize cursor movements
	mapnl         bool
}

var _ renderer = &cursedRenderer{}

func newCursedRenderer(w io.Writer, env []string, width, height int, hardTabs, backspace, mapnl bool, logger tv.Logger) (s *cursedRenderer) {
	s = new(cursedRenderer)
	s.w = w
	s.env = env
	s.term = tv.Environ(env).Getenv("TERM")
	s.logger = logger
	s.hardTabs = hardTabs
	s.backspace = backspace
	s.mapnl = mapnl
	s.width, s.height = width, height // This needs to happen before [cursedRenderer.reset].
	s.buf = tv.NewBuffer(s.width, s.height)
	// TODO: Use [ansi.WcWidth] by default and upgrade to [ansi.GraphemeWidth]
	// if the terminal supports it.
	s.method = ansi.GraphemeWidth
	reset(s)
	return
}

// close implements renderer.
func (s *cursedRenderer) close() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Go to the bottom of the screen.
	s.scr.MoveTo(s.buf, 0, s.buf.Height()-1)

	// Exit the altScreen and show cursor before closing. It's important that
	// we don't change the [cursedRenderer] altScreen and cursorHidden states
	// so that we can restore them when we start the renderer again. This is
	// used when the user suspends the program and then resumes it.
	if s.altScreen {
		s.scr.ExitAltScreen()
	}
	if s.cursorHidden {
		s.scr.ShowCursor()
	}

	if err := s.scr.Flush(); err != nil {
		return fmt.Errorf("bubbletea: error closing screen writer: %w", err)
	}

	x, y := s.scr.Position()

	// We want to clear the renderer state but not the cursor position. This is
	// because we might be putting the tea process in the background, run some
	// other process, and then return to the tea process. We want to keep the
	// cursor position so that we can continue where we left off.
	reset(s)
	s.scr.SetPosition(x, y)

	return nil
}

// writeString implements renderer.
func (s *cursedRenderer) writeString(str string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.scr.WriteString(str)
}

// flush implements renderer.
func (s *cursedRenderer) flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Render and queue changes to the screen buffer.
	s.scr.Render(s.buf)
	if s.lastCur != nil {
		if s.lastCur.Shape != s.cursor.Shape || s.lastCur.Blink != s.cursor.Blink {
			cursorStyle := encodeCursorStyle(s.lastCur.Shape, s.lastCur.Blink)
			_, _ = s.scr.WriteString(ansi.SetCursorStyle(cursorStyle))
			s.cursor.Shape = s.lastCur.Shape
			s.cursor.Blink = s.lastCur.Blink
		}
		if s.lastCur.Color != s.cursor.Color {
			seq := ansi.ResetCursorColor
			if s.lastCur.Color != nil {
				seq = ansi.SetCursorColor(s.lastCur.Color)
			}
			_, _ = s.scr.WriteString(seq)
			s.cursor.Color = s.lastCur.Color
		}

		// MoveTo must come after [cellbuf.Screen.Render] because the cursor
		// position might get updated during rendering.
		s.scr.MoveTo(s.buf, s.lastCur.X, s.lastCur.Y)
		s.cursor.Position = s.lastCur.Position
	}

	if err := s.scr.Flush(); err != nil {
		return fmt.Errorf("bubbletea: error flushing screen writer: %w", err)
	}
	return nil
}

// render implements renderer.
func (s *cursedRenderer) render(frame string, cur *Cursor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastFrame != nil && frame == *s.lastFrame &&
		(s.lastCur == nil && cur == nil || s.lastCur != nil && cur != nil && *s.lastCur == *cur) {
		return
	}

	s.lastFrame = &frame
	s.lastCur = cur
	ss := tv.NewStyledString(s.method, frame)
	bufHeight := s.height
	if !s.altScreen {
		// Inline mode resizes the screen based on the frame height and
		// terminal width. This is because the frame height can change based on
		// the content of the frame. For example, if the frame contains a list
		// of items, the height of the frame will be the number of items in the
		// list. This is different from the alt screen buffer, which has a
		// fixed height and width.
		bufHeight = ss.Buffer.Height()
	}

	// Clear our screen buffer before copying the new frame into it to ensure
	// we erase any old content.
	s.buf.Resize(s.width, bufHeight)
	s.buf.Clear()
	for y := 0; y < ss.Buffer.Height(); y++ {
		for x := 0; x < ss.Buffer.Width(); x++ {
			s.buf.SetCell(x, y, ss.Buffer.CellAt(x, y))
		}
	}

	if cur == nil {
		enableTextCursor(s, false)
	} else {
		enableTextCursor(s, true)
	}
}

// reset implements renderer.
func (s *cursedRenderer) reset() {
	s.mu.Lock()
	reset(s)
	s.mu.Unlock()
}

func reset(s *cursedRenderer) {
	scr := tv.NewTerminalRenderer(s.w, s.env)
	scr.SetColorProfile(s.profile)
	scr.SetRelativeCursor(!s.altScreen)
	scr.SetTabStops(s.width)
	scr.SetBackspace(s.backspace)
	scr.SetMapNewline(s.mapnl)
	scr.SetLogger(s.logger)
	if s.altScreen {
		scr.EnterAltScreen()
	} else {
		scr.ExitAltScreen()
	}
	if !s.cursorHidden {
		scr.ShowCursor()
	} else {
		scr.HideCursor()
	}
	s.scr = scr
}

// setColorProfile implements renderer.
func (s *cursedRenderer) setColorProfile(p colorprofile.Profile) {
	s.mu.Lock()
	s.profile = p
	s.scr.SetColorProfile(p)
	s.mu.Unlock()
}

// resize implements renderer.
func (s *cursedRenderer) resize(w, h int) {
	s.mu.Lock()
	if s.altScreen || w != s.width {
		// We need to mark the screen for clear to force a redraw. However, we
		// only do so if we're using alt screen or the width has changed.
		// That's because redrawing is expensive and we can avoid it if the
		// width hasn't changed in inline mode. On the other hand, when using
		// alt screen mode, we always want to redraw because some terminals
		// would scroll the screen and our content would be lost.
		s.scr.Clear()
	}

	s.scr.Resize(s.width, s.height)
	s.width, s.height = w, h
	repaint(s)
	s.mu.Unlock()
}

// clearScreen implements renderer.
func (s *cursedRenderer) clearScreen() {
	s.mu.Lock()
	// Move the cursor to the top left corner of the screen and trigger a full
	// screen redraw.
	_, _ = s.scr.WriteString(ansi.CursorHomePosition)
	s.scr.Redraw(s.buf) // force redraw
	repaint(s)
	s.mu.Unlock()
}

// enableAltScreen sets the alt screen mode.
func enableAltScreen(s *cursedRenderer, enable bool) {
	s.altScreen = enable
	if enable {
		s.scr.EnterAltScreen()
	} else {
		s.scr.ExitAltScreen()
	}
	s.scr.SetRelativeCursor(!s.altScreen)
	repaint(s)
}

// enterAltScreen implements renderer.
func (s *cursedRenderer) enterAltScreen() {
	s.mu.Lock()
	enableAltScreen(s, true)
	s.mu.Unlock()
}

// exitAltScreen implements renderer.
func (s *cursedRenderer) exitAltScreen() {
	s.mu.Lock()
	enableAltScreen(s, false)
	s.mu.Unlock()
}

// enableTextCursor sets the text cursor mode.
func enableTextCursor(s *cursedRenderer, enable bool) {
	s.cursorHidden = !enable
	if enable {
		s.scr.ShowCursor()
	} else {
		s.scr.HideCursor()
	}
}

// showCursor implements renderer.
func (s *cursedRenderer) showCursor() {
	s.mu.Lock()
	enableTextCursor(s, true)
	s.mu.Unlock()
}

// hideCursor implements renderer.
func (s *cursedRenderer) hideCursor() {
	s.mu.Lock()
	enableTextCursor(s, false)
	s.mu.Unlock()
}

// insertAbove implements renderer.
func (s *cursedRenderer) insertAbove(lines string) {
	s.mu.Lock()
	s.scr.PrependStyledString(s.buf, s.method, lines)
	s.mu.Unlock()
}

func (s *cursedRenderer) repaint() {
	s.mu.Lock()
	repaint(s)
	s.mu.Unlock()
}

func repaint(s *cursedRenderer) {
	s.lastFrame = nil
}
