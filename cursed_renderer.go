package tea

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

type cursedRenderer struct {
	w             io.Writer
	scr           *cellbuf.Screen
	lastFrame     *Frame
	term          string // the terminal type $TERM
	width, height int
	mu            sync.Mutex
	profile       colorprofile.Profile
	cursor        Cursor
	altScreen     bool
	cursorHidden  bool
	hardTabs      bool // whether to use hard tabs to optimize cursor movements
}

var _ renderer = &cursedRenderer{}

func newCursedRenderer(w io.Writer, term string, hardTabs bool) (s *cursedRenderer) {
	s = new(cursedRenderer)
	s.w = w
	s.term = term
	s.hardTabs = hardTabs
	s.reset()
	return
}

// close implements renderer.
func (s *cursedRenderer) close() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.scr.Close()
}

// flush implements renderer.
func (s *cursedRenderer) flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastFrame != nil && s.lastFrame.Cursor != nil {
		cur := s.lastFrame.Cursor
		s.scr.MoveTo(cur.Position.X, cur.Position.Y)
		s.cursor.Position = cur.Position

		if cur.Style != s.cursor.Style || cur.Blink != s.cursor.Blink {
			cursorStyle := encodeCursorStyle(cur.Style, cur.Blink)
			io.WriteString(s.w, ansi.SetCursorStyle(cursorStyle)) //nolint:errcheck
			s.cursor.Style = cur.Style
			s.cursor.Blink = cur.Blink
		}
		if cur.Color != s.cursor.Color {
			io.WriteString(s.w, ansi.SetCursorColor(cur.Color)) //nolint:errcheck
			s.cursor.Color = cur.Color
		}
	}
	s.scr.Render()
	return nil
}

// render implements renderer.
func (s *cursedRenderer) render(frame fmt.Stringer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var f Frame
	switch frame := frame.(type) {
	case Frame:
		f = frame
	default:
		f.Content = frame.String()
	}

	if s.lastFrame != nil && f == *s.lastFrame {
		return
	}

	s.lastFrame = &f
	if !s.altScreen {
		// Inline mode resizes the screen based on the frame height and
		// terminal width. This is because the frame height can change based on
		// the content of the frame. For example, if the frame contains a list
		// of items, the height of the frame will be the number of items in the
		// list. This is different from the alt screen buffer, which has a
		// fixed height and width.
		frameHeight := strings.Count(f.Content, "\n") + 1
		s.scr.Resize(s.width, frameHeight)
	}

	if ctx := s.scr.DefaultWindow(); ctx != nil {
		ctx.SetContent(f.Content)
	}

	if f.Cursor == nil {
		hideCursor(s)
	} else {
		showCursor(s)
	}
}

// reset implements renderer.
func (s *cursedRenderer) reset() {
	s.mu.Lock()
	s.scr = cellbuf.NewScreen(s.w, &cellbuf.ScreenOptions{
		Term:           s.term,
		Profile:        s.profile,
		AltScreen:      s.altScreen,
		RelativeCursor: !s.altScreen,
		ShowCursor:     !s.cursorHidden,
		Width:          s.width,
		Height:         s.height,
		HardTabs:       s.hardTabs,
	})
	s.mu.Unlock()
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
	s.width, s.height = w, h
	if s.altScreen {
		// We only resize the screen if we're in the alt screen buffer. Inline
		// mode resizes the screen based on the frame height and terminal
		// width. See [screenRenderer.render] for more details.
		s.scr.Resize(s.width, s.height)
	}

	repaint(s)
	s.mu.Unlock()
}

// clearScreen implements renderer.
func (s *cursedRenderer) clearScreen() {
	s.mu.Lock()
	s.scr.Clear()
	repaint(s)
	s.mu.Unlock()
}

// enterAltScreen implements renderer.
func (s *cursedRenderer) enterAltScreen() {
	s.mu.Lock()
	s.altScreen = true
	s.scr.EnterAltScreen()
	s.scr.SetRelativeCursor(!s.altScreen)
	s.scr.Resize(s.width, s.height)
	s.lastFrame = nil
	s.mu.Unlock()
}

// exitAltScreen implements renderer.
func (s *cursedRenderer) exitAltScreen() {
	s.mu.Lock()
	s.altScreen = false
	s.scr.ExitAltScreen()
	s.scr.SetRelativeCursor(!s.altScreen)
	s.scr.Resize(s.width, strings.Count((*s.lastFrame).Content, "\n")+1)
	repaint(s)
	s.mu.Unlock()
}

// showCursor implements renderer.
func (s *cursedRenderer) showCursor() {
	s.mu.Lock()
	showCursor(s)
	s.mu.Unlock()
}

func showCursor(s *cursedRenderer) {
	s.cursorHidden = false
	s.scr.ShowCursor()
}

// hideCursor implements renderer.
func (s *cursedRenderer) hideCursor() {
	s.mu.Lock()
	hideCursor(s)
	s.mu.Unlock()
}

func hideCursor(s *cursedRenderer) {
	s.cursorHidden = true
	s.scr.HideCursor()
}

// insertAbove implements renderer.
func (s *cursedRenderer) insertAbove(lines string) {
	s.mu.Lock()
	s.scr.InsertAbove(lines)
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
