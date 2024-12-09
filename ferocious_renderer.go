package tea

import (
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
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

func newScreenRenderer(p colorprofile.Profile, term string) (s *screenRenderer) {
	s = new(screenRenderer)
	s.term = term
	s.profile = p
	return
}

// close implements renderer.
func (s *screenRenderer) close() (err error) {
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

// update implements renderer.
func (s *screenRenderer) update(msg Msg) {
	switch msg := msg.(type) {
	case ColorProfileMsg:
		s.profile = msg.Profile
	case WindowSizeMsg:
		s.width, s.height = msg.Width, msg.Height
		if s.altScreen {
			// Resize alternate screen
			s.scr.Resize(s.width, s.height)
		}
	case clearScreenMsg:
		s.scr.Clear()
		s.repaint()
	case repaintMsg:
		s.repaint()
	case rendererWriter:
		s.w = msg.Writer
		s.reset()
	case enableModeMsg:
		switch ansi.Mode(ansi.DECMode(msg)) {
		case ansi.AltScreenSaveCursorMode:
			s.altScreen = true
			s.scr.EnterAltScreen()
			s.scr.SetRelativeCursor(!s.altScreen)
			s.scr.Resize(s.width, s.height)
			s.repaint()
		case ansi.TextCursorEnableMode:
			s.cursorHidden = false
			s.scr.ShowCursor()
		}
	case disableModeMsg:
		switch ansi.Mode(ansi.DECMode(msg)) {
		case ansi.AltScreenSaveCursorMode:
			s.altScreen = false
			s.scr.ExitAltScreen()
			s.scr.SetRelativeCursor(!s.altScreen)
			s.scr.Resize(s.width, strings.Count(s.lastFrame, "\n")+1)
			s.repaint()
		case ansi.TextCursorEnableMode:
			s.cursorHidden = true
			s.scr.HideCursor()
		}
	case printLineMessage:
		s.scr.InsertAbove(msg.messageBody)
	case setCursorPosMsg:
		s.scr.MoveTo(msg.X, msg.Y)
	}
}

func (s *screenRenderer) repaint() {
	s.lastFrame = ""
}
