package tea

import (
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/shampoo"
	"github.com/charmbracelet/x/ansi"
)

type screenRenderer struct {
	w             io.Writer
	width, height int
	screen        *shampoo.Renderer
	ticker        *time.Ticker
	donec         chan struct{}
	lastFrame     string
	bpActive      bool
}

var _ renderer = &screenRenderer{}

func newScreenRenderer(w io.Writer, width, height int) *screenRenderer {
	framerate := time.Second / 60
	screen := shampoo.NewRenderer(w, width, height)
	return &screenRenderer{
		screen: screen,
		w:      w,
		width:  width,
		height: height,
		ticker: time.NewTicker(framerate),
		donec:  make(chan struct{}, 1),
	}
}

// resize resizes the screen.
func (s *screenRenderer) resize(width, height int) {
	if width == s.width && height == s.height {
		return
	}
	if s.altScreen() || s.width != width {
		s.clearScreen()
	}
	s.width, s.height = width, height
	s.lastFrame = ""
}

// altScreen implements renderer.
func (s *screenRenderer) altScreen() bool {
	return s.screen.AltScreen()
}

// bracketedPasteActive implements renderer.
func (s *screenRenderer) bracketedPasteActive() bool {
	return s.bpActive
}

// clearScreen implements renderer.
func (s *screenRenderer) clearScreen() {
	s.screen.Clear()
}

// disableBracketedPaste implements renderer.
func (s *screenRenderer) disableBracketedPaste() {
	ansi.Execute(s.w, ansi.DisableBracketedPaste)
}

// disableMouseAllMotion implements renderer.
func (s *screenRenderer) disableMouseAllMotion() {
	ansi.Execute(s.w, ansi.DisableMouseAllMotion)
}

// disableMouseCellMotion implements renderer.
func (s *screenRenderer) disableMouseCellMotion() {
	ansi.Execute(s.w, ansi.DisableMouseCellMotion)
}

// disableMouseSGRMode implements renderer.
func (s *screenRenderer) disableMouseSGRMode() {
	ansi.Execute(s.w, ansi.DisableMouseSgrExt)
}

// enableBracketedPaste implements renderer.
func (s *screenRenderer) enableBracketedPaste() {
	ansi.Execute(s.w, ansi.EnableBracketedPaste)
}

// enableMouseAllMotion implements renderer.
func (s *screenRenderer) enableMouseAllMotion() {
	ansi.Execute(s.w, ansi.EnableMouseAllMotion)
}

// enableMouseCellMotion implements renderer.
func (s *screenRenderer) enableMouseCellMotion() {
	ansi.Execute(s.w, ansi.EnableMouseCellMotion)
}

// enableMouseSGRMode implements renderer.
func (s *screenRenderer) enableMouseSGRMode() {
	ansi.Execute(s.w, ansi.EnableMouseSgrExt)
}

// enterAltScreen implements renderer.
func (s *screenRenderer) enterAltScreen() {
	s.screen.SetAltScreen(true)
	s.screen.Clear()
	s.lastFrame = ""
}

// exitAltScreen implements renderer.
func (s *screenRenderer) exitAltScreen() {
	s.screen.SetAltScreen(false)
	s.screen.Clear()
	s.lastFrame = ""
}

// hideCursor implements renderer.
func (s *screenRenderer) hideCursor() {
	s.screen.SetCursorVisibility(false)
}

// kill implements renderer.
func (s *screenRenderer) kill() {
	sync.OnceFunc(func() {
		s.donec <- struct{}{}
	})
}

// repaint implements renderer.
func (s *screenRenderer) repaint() {
	s.screen.Draw(s.lastFrame)
}

// requestBackgroundColor implements renderer.
func (s *screenRenderer) requestBackgroundColor() {
	ansi.Execute(s.w, ansi.RequestBackgroundColor)
}

// requestDeviceAttributes implements renderer.
func (s *screenRenderer) requestDeviceAttributes() {
	ansi.Execute(s.w, ansi.RequestPrimaryDeviceAttributes)
}

// requestKittyKeyboard implements renderer.
func (s *screenRenderer) requestKittyKeyboard() {
	ansi.Execute(s.w, ansi.RequestKittyKeyboard)
}

// setWindowTitle implements renderer.
func (s *screenRenderer) setWindowTitle(title string) {
	ansi.Execute(s.w, ansi.SetWindowTitle(title))
}

// showCursor implements renderer.
func (s *screenRenderer) showCursor() {
	s.screen.SetCursorVisibility(true)
}

// start implements renderer.
func (s *screenRenderer) start() {
	go func() {
		for {
			select {
			case <-s.donec:
				s.ticker.Stop()
				return
			case <-s.ticker.C:
				if err := s.screen.Flush(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}()
}

// stop implements renderer.
func (s *screenRenderer) stop() {
	sync.OnceFunc(func() {
		s.donec <- struct{}{}
	})

	s.screen.Close()
}

// write implements renderer.
func (s *screenRenderer) write(content string) {
	if s.lastFrame == content {
		return
	}

	// If we know the output's height, we can use it to determine how many
	// lines we can render. We drop lines from the top of the render buffer if
	// necessary, as we can't navigate the cursor into the terminal's scrollback
	// buffer.
	if s.height > 0 {
		frameLines := strings.Split(content, "\n")
		if frameHeight := len(frameLines); frameHeight > s.height {
			content = strings.Join(frameLines[frameHeight-s.height:], "\n")
		}
	}

	if s.altScreen() {
		s.screen.Resize(s.width, s.height)
	} else {
		s.screen.Resize(s.width, lipgloss.Height(content))
	}

	s.screen.Draw(content) //nolint:errcheck
	s.lastFrame = content
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
