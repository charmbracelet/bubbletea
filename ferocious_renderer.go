package tea

import (
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/cellbuf"
)

type screenRenderer struct {
	w             io.Writer
	scr           *cellbuf.Screen
	lastFrame     string
	term          string // the terminal type $TERM
	width, height int
	mu            sync.Mutex
	profile       colorprofile.Profile
	altScreen     bool
	cursorHidden  bool
}

var _ renderer = &screenRenderer{}

func newScreenRenderer(w io.Writer, term string) (s *screenRenderer) {
	s = new(screenRenderer)
	s.w = w
	s.term = term
	s.reset()
	return
}

// close implements renderer.
func (s *screenRenderer) close() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.scr.Close()
}

// flush implements renderer.
func (s *screenRenderer) flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scr.Render()
	return nil
}

// render implements renderer.
func (s *screenRenderer) render(frame string) {
	if frame == s.lastFrame {
		return
	}

	s.lastFrame = frame
	if !s.altScreen {
		// Inline mode resizes the screen based on the frame height and
		// terminal width. This is because the frame height can change based on
		// the content of the frame. For example, if the frame contains a list
		// of items, the height of the frame will be the number of items in the
		// list. This is different from the alt screen buffer, which has a
		// fixed height and width.
		frameHeight := strings.Count(frame, "\n") + 1
		s.scr.Resize(s.width, frameHeight)
	}

	cellbuf.Paint(s.scr, frame)
}

// reset implements renderer.
func (s *screenRenderer) reset() {
	s.scr = cellbuf.NewScreen(s.w, &cellbuf.ScreenOptions{
		Term:           s.term,
		Profile:        s.profile,
		AltScreen:      s.altScreen,
		RelativeCursor: !s.altScreen,
		ShowCursor:     !s.cursorHidden,
		Width:          s.width,
		Height:         s.height,
	})
}

// setColorProfile implements renderer.
func (s *screenRenderer) setColorProfile(p colorprofile.Profile) {
	s.profile = p
	s.scr.SetColorProfile(p)
}

// resize implements renderer.
func (s *screenRenderer) resize(w, h int) {
	s.width, s.height = w, h
	if s.altScreen {
		// We only resize the screen if we're in the alt screen buffer. Inline
		// mode resizes the screen based on the frame height and terminal
		// width. See [screenRenderer.render] for more details.
		s.scr.Resize(s.width, s.height)
	}
}

// clearScreen implements renderer.
func (s *screenRenderer) clearScreen() {
	s.scr.Clear()
	s.repaint()
}

// enterAltScreen implements renderer.
func (s *screenRenderer) enterAltScreen() {
	s.altScreen = true
	s.scr.EnterAltScreen()
	s.scr.SetRelativeCursor(!s.altScreen)
	s.scr.Resize(s.width, s.height)
	s.repaint()
}

// exitAltScreen implements renderer.
func (s *screenRenderer) exitAltScreen() {
	s.altScreen = false
	s.scr.ExitAltScreen()
	s.scr.SetRelativeCursor(!s.altScreen)
	s.scr.Resize(s.width, strings.Count(s.lastFrame, "\n")+1)
	s.repaint()
}

// showCursor implements renderer.
func (s *screenRenderer) showCursor() {
	s.cursorHidden = false
	s.scr.ShowCursor()
}

// hideCursor implements renderer.
func (s *screenRenderer) hideCursor() {
	s.cursorHidden = true
	s.scr.HideCursor()
}

// insertAbove implements renderer.
func (s *screenRenderer) insertAbove(lines string) {
	s.scr.InsertAbove(lines)
}

// moveTo implements renderer.
func (s *screenRenderer) moveTo(x, y int) {
	s.scr.MoveTo(x, y)
}

func (s *screenRenderer) repaint() {
	s.lastFrame = ""
}
