package tea

import (
	"fmt"
	"image/color"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

type cursedRenderer struct {
	w                   io.Writer
	scr                 *uv.TerminalRenderer
	buf                 uv.ScreenBuffer
	lastFrame           *string
	lastCur             *Cursor
	env                 []string
	term                string // the terminal type $TERM
	width, height       int
	lastFrameHeight     int // the height of the last rendered frame, used to determine if we need to resize the screen buffer
	mu                  sync.Mutex
	profile             colorprofile.Profile
	cursor              Cursor
	method              ansi.Method
	logger              uv.Logger
	layer               Layer // the last rendered layer
	setCc, setFg, setBg color.Color
	windowTitleSet      string // the last set window title
	windowTitle         string // the desired title of the terminal window
	altScreen           bool
	cursorHidden        bool
	hardTabs            bool // whether to use hard tabs to optimize cursor movements
	backspace           bool // whether to use backspace to optimize cursor movements
	mapnl               bool
}

var _ renderer = &cursedRenderer{}

func newCursedRenderer(w io.Writer, env []string, width, height int, hardTabs, backspace, mapnl bool, logger uv.Logger) (s *cursedRenderer) {
	s = new(cursedRenderer)
	s.w = w
	s.env = env
	s.term = uv.Environ(env).Getenv("TERM")
	s.logger = logger
	s.hardTabs = hardTabs
	s.backspace = backspace
	s.mapnl = mapnl
	s.width, s.height = width, height // This needs to happen before [cursedRenderer.reset].
	s.buf = uv.NewScreenBuffer(s.width, s.height)
	reset(s)
	return
}

// close implements renderer.
func (s *cursedRenderer) close() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Go to the bottom of the screen.
	s.scr.MoveTo(0, s.buf.Height()-1)

	// Exit the altScreen and show cursor before closing. It's important that
	// we don't change the [cursedRenderer] altScreen and cursorHidden states
	// so that we can restore them when we start the renderer again. This is
	// used when the user suspends the program and then resumes it.
	if s.altScreen {
		s.scr.ExitAltScreen()
	}
	if s.cursorHidden {
		s.scr.ShowCursor()
		s.cursorHidden = false
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

	// Reset cursor style state so that we can restore it again when we start
	// the renderer again.
	s.cursor = Cursor{}

	return nil
}

// writeString implements renderer.
func (s *cursedRenderer) writeString(str string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.scr.WriteString(str)
}

// resetLinesRendered implements renderer.
func (s *cursedRenderer) resetLinesRendered() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.altScreen {
		var frameHeight int
		if s.lastFrame != nil {
			frameHeight = strings.Count(*s.lastFrame, "\n") + 1
		}

		io.WriteString(s.w, strings.Repeat("\n", max(0, frameHeight-1))) //nolint:errcheck,gosec
	}
}

// flush implements renderer.
func (s *cursedRenderer) flush(p *Program) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set window title.
	if s.windowTitle != s.windowTitleSet {
		_, _ = s.scr.WriteString(ansi.SetWindowTitle(s.windowTitle))
		s.windowTitleSet = s.windowTitle
	}
	// Set terminal colors.
	for _, c := range []struct {
		rendererColor *color.Color
		programColor  *color.Color
		reset         string
		setter        func(string) string
	}{
		{rendererColor: &s.setCc, programColor: &p.setCc, reset: ansi.ResetCursorColor, setter: ansi.SetCursorColor},
		{rendererColor: &s.setFg, programColor: &p.setFg, reset: ansi.ResetForegroundColor, setter: ansi.SetForegroundColor},
		{rendererColor: &s.setBg, programColor: &p.setBg, reset: ansi.ResetBackgroundColor, setter: ansi.SetBackgroundColor},
	} {
		if *c.rendererColor != *c.programColor {
			if *c.rendererColor == nil {
				// Reset the color if it was set to nil.
				_, _ = s.scr.WriteString(c.reset)
			} else {
				// Set the color.
				col, ok := colorful.MakeColor(*c.rendererColor)
				if ok {
					_, _ = s.scr.WriteString(c.setter(col.Hex()))
				}
			}
			*c.programColor = *c.rendererColor
		}
	}

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
				c, ok := colorful.MakeColor(s.lastCur.Color)
				if ok {
					seq = ansi.SetCursorColor(c.Hex())
				}
			}
			_, _ = s.scr.WriteString(seq)
			s.cursor.Color = s.lastCur.Color
		}
	}

	// Render and queue changes to the screen buffer.
	s.scr.Render(s.buf.Buffer)
	if s.lastCur != nil {
		// MoveTo must come after [uv.TerminalRenderer.Render] because the
		// cursor position might get updated during rendering.
		s.scr.MoveTo(s.lastCur.X, s.lastCur.Y)
		s.cursor.Position = s.lastCur.Position
	}

	if err := s.scr.Flush(); err != nil {
		return fmt.Errorf("bubbletea: error flushing screen writer: %w", err)
	}
	return nil
}

// render implements renderer.
func (s *cursedRenderer) render(v View) {
	s.mu.Lock()
	defer s.mu.Unlock()

	frameArea := uv.Rect(0, 0, s.width, s.height)
	if v.Layer == nil {
		// If the component is nil, we should clear the screen buffer.
		frameArea.Max.Y = 0
	}

	if !s.altScreen {
		// Inline mode resizes the screen based on the frame height and
		// terminal width. This is because the frame height can change based on
		// the content of the frame. For example, if the frame contains a list
		// of items, the height of the frame will be the number of items in the
		// list. This is different from the alt screen buffer, which has a
		// fixed height and width.
		switch l := v.Layer.(type) {
		case *uv.StyledString:
			frameArea.Max.Y = l.Height()
		case interface{ Bounds() uv.Rectangle }:
			frameArea.Max.Y = l.Bounds().Dy()
		}

		// Resize the screen buffer to match the frame area. This is necessary
		// to ensure that the screen buffer is the same size as the frame area
		// and to avoid rendering issues when the frame area is smaller than
		// the screen buffer.
		s.buf.Resize(frameArea.Dx(), frameArea.Dy())
	}
	// Clear our screen buffer before copying the new frame into it to ensure
	// we erase any old content.
	s.buf.Clear()
	if v.Layer != nil {
		v.Layer.Draw(s.buf, frameArea)
	}

	frame := s.buf.Render()

	// If an empty string was passed we should clear existing output and
	// rendering nothing. Rather than introduce additional state to manage
	// this, we render a single space as a simple (albeit less correct)
	// solution.
	if frame == "" {
		frame = " "
	}

	cur := v.Cursor

	s.windowTitle = v.WindowTitle

	// Ensure we have any desired terminal colors set.
	s.setBg = v.BackgroundColor
	s.setFg = v.ForegroundColor
	if cur != nil {
		s.setCc = cur.Color
	}
	if s.lastFrame != nil && frame == *s.lastFrame &&
		(s.lastCur == nil && cur == nil || s.lastCur != nil && cur != nil && *s.lastCur == *cur) {
		return
	}

	s.layer = v.Layer
	s.lastCur = cur
	s.lastFrameHeight = frameArea.Dy()

	// Cache the last rendered frame so we can avoid re-rendering it if
	// the frame hasn't changed.
	lastFrame := frame
	s.lastFrame = &lastFrame

	if cur == nil {
		enableTextCursor(s, false)
	} else {
		enableTextCursor(s, true)
	}
}

// hit implements renderer.
func (s *cursedRenderer) hit(mouse MouseMsg) []Msg {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.layer != nil {
		if h, ok := s.layer.(Hittable); ok {
			m := mouse.Mouse()
			if id := h.Hit(m.X, m.Y); id != "" {
				return []Msg{LayerHitMsg{
					ID:    id,
					Mouse: mouse,
				}}
			}
		}
	}

	return []Msg{}
}

// setCursorColor implements renderer.
func (s *cursedRenderer) setCursorColor(c color.Color) {
	s.mu.Lock()
	s.setCc = c
	s.mu.Unlock()
}

// setForegroundColor implements renderer.
func (s *cursedRenderer) setForegroundColor(c color.Color) {
	s.mu.Lock()
	s.setFg = c
	s.mu.Unlock()
}

// setBackgroundColor implements renderer.
func (s *cursedRenderer) setBackgroundColor(c color.Color) {
	s.mu.Lock()
	s.setBg = c
	s.mu.Unlock()
}

// setWindowTitle implements renderer.
func (s *cursedRenderer) setWindowTitle(title string) {
	s.mu.Lock()
	s.windowTitle = title
	s.mu.Unlock()
}

// reset implements renderer.
func (s *cursedRenderer) reset() {
	s.mu.Lock()
	reset(s)
	s.mu.Unlock()
}

func reset(s *cursedRenderer) {
	scr := uv.NewTerminalRenderer(s.w, s.env)
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
		s.scr.Erase()
	}
	if s.altScreen {
		s.buf.Resize(w, h)
	}

	// We need to reset the touched lines buffer to match the new height.
	s.buf.Touched = nil

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
	s.scr.Redraw(s.buf.Buffer) // force redraw
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
	strLines := strings.Split(lines, "\n")
	for i, line := range strLines {
		if ansi.StringWidth(line) > s.width {
			// If the line is wider than the screen, truncate it.
			line = ansi.Truncate(line, s.width, "")
		}
		strLines[i] = line
	}
	s.scr.PrependString(strings.Join(strLines, "\n"))
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
