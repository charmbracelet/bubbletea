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
	w               io.Writer
	scr             *uv.TerminalRenderer
	buf             uv.ScreenBuffer
	lastFrame       *string
	lastView        *View
	env             []string
	term            string // the terminal type $TERM
	width, height   int
	lastFrameHeight int // the height of the last rendered frame, used to determine if we need to resize the screen buffer
	mu              sync.Mutex
	profile         colorprofile.Profile
	logger          uv.Logger
	view            View
	hardTabs        bool // whether to use hard tabs to optimize cursor movements
	backspace       bool // whether to use backspace to optimize cursor movements
	mapnl           bool
}

var _ renderer = &cursedRenderer{}

func newCursedRenderer(w io.Writer, env []string, width, height int) (s *cursedRenderer) {
	s = new(cursedRenderer)
	s.w = w
	s.env = env
	s.term = uv.Environ(env).Getenv("TERM")
	s.width, s.height = width, height // This needs to happen before [cursedRenderer.reset].
	s.buf = uv.NewScreenBuffer(s.width, s.height)
	reset(s)
	return
}

// setLogger sets the logger for the renderer.
func (s *cursedRenderer) setLogger(logger uv.Logger) {
	s.mu.Lock()
	s.logger = logger
	s.scr.SetLogger(logger)
	s.mu.Unlock()
}

// setOptimizations sets the cursor movement optimizations.
func (s *cursedRenderer) setOptimizations(hardTabs, backspace, mapnl bool) {
	s.mu.Lock()
	s.hardTabs = hardTabs
	s.backspace = backspace
	s.mapnl = mapnl
	s.scr.SetTabStops(s.width)
	s.scr.SetBackspace(s.backspace)
	s.scr.SetMapNewline(s.mapnl)
	s.mu.Unlock()
}

// start implements renderer.
func (s *cursedRenderer) start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Force a full repaint to ensure the screen is in a consistent state.
	repaint(s)

	if s.lastView == nil {
		return
	}

	if s.lastView.AltScreen {
		enableAltScreen(s, true)
	}
	if s.lastView.Cursor != nil {
		enableTextCursor(s, true)
		if s.lastView.Cursor.Color != nil {
			col, ok := colorful.MakeColor(s.lastView.Cursor.Color)
			if ok {
				_, _ = s.scr.WriteString(ansi.SetCursorColor(col.Hex()))
			}
		}
		curStyle := encodeCursorStyle(s.lastView.Cursor.Shape, s.lastView.Cursor.Blink)
		if curStyle != 0 && curStyle != 1 {
			_, _ = s.scr.WriteString(ansi.SetCursorStyle(curStyle))
		}
	}
	if s.lastView.ForegroundColor != nil {
		col, ok := colorful.MakeColor(s.lastView.ForegroundColor)
		if ok {
			_, _ = s.scr.WriteString(ansi.SetForegroundColor(col.Hex()))
		}
	}
	if s.lastView.BackgroundColor != nil {
		col, ok := colorful.MakeColor(s.lastView.BackgroundColor)
		if ok {
			_, _ = s.scr.WriteString(ansi.SetBackgroundColor(col.Hex()))
		}
	}
	if !s.lastView.DisableBracketedPasteMode {
		_, _ = s.scr.WriteString(ansi.SetBracketedPasteMode)
	}
	if s.lastView.ReportFocus {
		_, _ = s.scr.WriteString(ansi.SetFocusEventMode)
	}
	switch s.lastView.MouseMode {
	case MouseModeNone:
	case MouseModeCellMotion:
		_, _ = s.scr.WriteString(ansi.SetButtonEventMouseMode + ansi.SetSgrExtMouseMode)
	case MouseModeAllMotion:
		_, _ = s.scr.WriteString(ansi.SetAnyEventMouseMode + ansi.SetSgrExtMouseMode)
	}
	if s.lastView.WindowTitle != "" {
		_, _ = s.scr.WriteString(ansi.SetWindowTitle(s.lastView.WindowTitle))
	}
	if s.lastView.ProgressBar != nil {
		setProgressBar(s, s.lastView.ProgressBar)
	}
	if !s.lastView.DisableKeyEnhancements {
		kittyFlags := ansi.KittyDisambiguateEscapeCodes
		if s.lastView.KeyReleases {
			kittyFlags |= ansi.KittyReportEventTypes
		}
		if s.lastView.UniformKeyLayout {
			kittyFlags |= ansi.KittyReportAlternateKeys | ansi.KittyReportAllKeysAsEscapeCodes
		}
		_, _ = s.scr.WriteString(ansi.KittyKeyboard(kittyFlags, 1))
	}
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
	if lv := s.lastView; lv != nil { //nolint:nestif
		if lv.AltScreen {
			s.scr.ExitAltScreen()
		} else {
			_, _ = s.scr.WriteString("\r" + ansi.EraseScreenBelow)
		}
		if lv.Cursor == nil {
			s.scr.ShowCursor()
		}
		if !lv.DisableBracketedPasteMode {
			_, _ = s.scr.WriteString(ansi.ResetBracketedPasteMode)
		}
		if lv.ReportFocus {
			_, _ = s.scr.WriteString(ansi.ResetFocusEventMode)
		}
		switch lv.MouseMode {
		case MouseModeNone:
		case MouseModeCellMotion, MouseModeAllMotion:
			_, _ = s.scr.WriteString(ansi.ResetButtonEventMouseMode +
				ansi.ResetAnyEventMouseMode +
				ansi.ResetSgrExtMouseMode)
		}

		if lv.WindowTitle != "" {
			// Clear the window title if it was set.
			_, _ = s.scr.WriteString(ansi.SetWindowTitle(""))
		}
		if lc := lv.Cursor; lc != nil {
			curShape := encodeCursorStyle(lc.Shape, lc.Blink)
			if curShape != 0 && curShape != 1 {
				// Reset the cursor style to default if it was set to something other
				// blinking block.
				_, _ = s.scr.WriteString(ansi.SetCursorStyle(0))
			}

			if lc.Color != nil {
				_, _ = s.scr.WriteString(ansi.ResetCursorColor)
			}
		}

		if lv.BackgroundColor != nil {
			_, _ = s.scr.WriteString(ansi.ResetBackgroundColor)
		}
		if lv.ForegroundColor != nil {
			_, _ = s.scr.WriteString(ansi.ResetForegroundColor)
		}
		if lv.ProgressBar != nil {
			_, _ = s.scr.WriteString(ansi.ResetProgressBar)
		}

		if s.lastView != nil && !s.lastView.DisableKeyEnhancements {
			// NOTE: This needs to happen after we exit the alt screen.
			_, _ = s.scr.WriteString(ansi.KittyKeyboard(0, 1))
		}
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

	return s.scr.WriteString(str) //nolint:wrapcheck
}

// resetLinesRendered implements renderer.
func (s *cursedRenderer) resetLinesRendered() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastView != nil && !s.lastView.AltScreen {
		var frameHeight int
		if s.lastFrame != nil {
			frameHeight = strings.Count(*s.lastFrame, "\n") + 1
		}

		io.WriteString(s.w, strings.Repeat("\n", max(0, frameHeight-1))) //nolint:errcheck,gosec
	}
}

// flush implements renderer.
func (s *cursedRenderer) flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	view := s.view

	// Alt screen mode.
	enableAltScreen(s, view.AltScreen)
	// Cursor visibility.
	enableTextCursor(s, view.Cursor != nil)

	// bracketed paste mode.
	if s.lastView == nil || view.DisableBracketedPasteMode != s.lastView.DisableBracketedPasteMode {
		if !view.DisableBracketedPasteMode {
			_, _ = s.scr.WriteString(ansi.SetBracketedPasteMode)
		} else if s.lastView != nil {
			_, _ = s.scr.WriteString(ansi.ResetBracketedPasteMode)
		}
	}

	// report focus events mode.
	if s.lastView == nil || s.lastView.ReportFocus != view.ReportFocus {
		if view.ReportFocus {
			_, _ = s.scr.WriteString(ansi.SetFocusEventMode)
		} else if s.lastView != nil {
			_, _ = s.scr.WriteString(ansi.ResetFocusEventMode)
		}
	}

	// mouse events mode.
	if s.lastView == nil || view.MouseMode != s.lastView.MouseMode {
		switch view.MouseMode {
		case MouseModeNone:
			if s.lastView != nil && s.lastView.MouseMode != MouseModeNone {
				_, _ = s.scr.WriteString(ansi.ResetButtonEventMouseMode +
					ansi.ResetAnyEventMouseMode +
					ansi.ResetSgrExtMouseMode)
			}
		case MouseModeCellMotion:
			if s.lastView != nil && s.lastView.MouseMode == MouseModeAllMotion {
				_, _ = s.scr.WriteString(ansi.ResetAnyEventMouseMode)
			}
			_, _ = s.scr.WriteString(ansi.SetButtonEventMouseMode + ansi.SetSgrExtMouseMode)
		case MouseModeAllMotion:
			if s.lastView != nil && s.lastView.MouseMode == MouseModeCellMotion {
				_, _ = s.scr.WriteString(ansi.ResetButtonEventMouseMode)
			}
			_, _ = s.scr.WriteString(ansi.SetAnyEventMouseMode + ansi.SetSgrExtMouseMode)
		}
	}

	// Set window title.
	if s.lastView == nil || view.WindowTitle != s.lastView.WindowTitle {
		if s.lastView != nil || view.WindowTitle != "" {
			_, _ = s.scr.WriteString(ansi.SetWindowTitle(view.WindowTitle))
		}
	}

	// kitty keyboard protocol
	//nolint:nestif
	if s.lastView == nil || view.DisableKeyEnhancements != s.lastView.DisableKeyEnhancements ||
		view.KeyReleases != s.lastView.KeyReleases ||
		view.UniformKeyLayout != s.lastView.UniformKeyLayout ||
		view.AltScreen != s.lastView.AltScreen {
		// NOTE: We need to reset the keyboard protocol when switching
		// between main and alt screen. This is because the specs specify
		// two different states for the main and alt screen.
		if view.DisableKeyEnhancements {
			_, _ = s.scr.WriteString(ansi.KittyKeyboard(0, 1))
		} else {
			kittyFlags := ansi.KittyDisambiguateEscapeCodes // always enable basic key disambiguation
			if view.KeyReleases {
				kittyFlags |= ansi.KittyReportEventTypes
			}
			if view.UniformKeyLayout {
				kittyFlags |= ansi.KittyReportAlternateKeys | ansi.KittyReportAllKeysAsEscapeCodes
			}
			_, _ = s.scr.WriteString(ansi.KittyKeyboard(kittyFlags, 1))
		}
		// Request keyboard enhancements when they change
		_, _ = s.scr.WriteString(ansi.RequestKittyKeyboard)
	}

	// Set terminal colors.
	var (
		cc, lcc  color.Color
		lfg, lbg color.Color
	)
	if view.Cursor != nil {
		cc = view.Cursor.Color
	}
	if s.lastView != nil {
		if s.lastView.Cursor != nil {
			lcc = s.lastView.Cursor.Color
		}
		lfg = s.lastView.ForegroundColor
		lbg = s.lastView.BackgroundColor
	}
	for _, c := range []struct {
		newColor color.Color
		oldColor color.Color
		reset    string
		setter   func(string) string
	}{
		{newColor: cc, oldColor: lcc, reset: ansi.ResetCursorColor, setter: ansi.SetCursorColor},
		{newColor: view.ForegroundColor, oldColor: lfg, reset: ansi.ResetForegroundColor, setter: ansi.SetForegroundColor},
		{newColor: view.BackgroundColor, oldColor: lbg, reset: ansi.ResetBackgroundColor, setter: ansi.SetBackgroundColor},
	} {
		if c.newColor != c.oldColor {
			if c.newColor == nil {
				// Reset the color if it was set to nil.
				_, _ = s.scr.WriteString(c.reset)
			} else {
				// Set the color.
				col, ok := colorful.MakeColor(c.newColor)
				if ok {
					_, _ = s.scr.WriteString(c.setter(col.Hex()))
				}
			}
		}
	}

	if cur := view.Cursor; cur != nil {
		// Set cursor shape and blink if set.
		var lcur *Cursor
		lv := s.lastView
		if lv != nil {
			lcur = lv.Cursor
		}
		if lv == nil || lcur == nil || cur.Shape != lcur.Shape || cur.Blink != lcur.Blink {
			curStyle := encodeCursorStyle(cur.Shape, cur.Blink)
			_, _ = s.scr.WriteString(ansi.SetCursorStyle(curStyle))
		}
	}

	if (view.ProgressBar == nil) != (s.lastView == nil || s.lastView.ProgressBar == nil) {
		// Render or clear the progress bar if it was added or removed.
		setProgressBar(s, view.ProgressBar)
	}

	// Render and queue changes to the screen buffer.
	s.scr.Render(s.buf.Buffer)

	if cur := view.Cursor; cur != nil {
		// MoveTo must come after [uv.TerminalRenderer.Render] because the
		// cursor position might get updated during rendering.
		s.scr.MoveTo(view.Cursor.X, view.Cursor.Y)
	}

	if err := s.scr.Flush(); err != nil {
		return fmt.Errorf("bubbletea: error flushing screen writer: %w", err)
	}

	s.lastView = &view

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

	if !v.AltScreen {
		// Inline mode resizes the screen based on the frame height and
		// terminal width. This is because the frame height can change based on
		// the content of the frame. For example, if the frame contains a list
		// of items, the height of the frame will be the number of items in the
		// list. This is different from the alt screen buffer, which has a
		// fixed height and width.
		frameHeight := frameArea.Dy()
		switch l := v.Layer.(type) {
		case *uv.StyledString:
			frameHeight = l.Height()
		case interface{ Bounds() uv.Rectangle }:
			frameHeight = l.Bounds().Dy()
		}

		if frameHeight != frameArea.Dy() {
			frameArea.Max.Y = frameHeight
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
		v.Layer.Draw(s.buf, s.buf.Bounds())
	}

	// If the frame height is greater than the screen height, we drop the
	// lines from the top of the buffer.
	if frameHeight := frameArea.Dy(); frameHeight > s.height {
		s.buf.Lines = s.buf.Lines[frameHeight-s.height:]
	}

	frame := s.buf.Render()

	// If an empty string was passed we should clear existing output and
	// rendering nothing. Rather than introduce additional state to manage
	// this, we render a single space as a simple (albeit less correct)
	// solution.
	if frame == "" {
		frame = " "
	}

	s.view = v

	if s.lastFrame != nil && frame == *s.lastFrame && s.lastView != nil && viewEquals(v, *s.lastView) {
		return
	}

	s.lastFrameHeight = frameArea.Dy()

	// Cache the last rendered frame so we can avoid re-rendering it if
	// the frame hasn't changed.
	lastFrame := frame
	s.lastFrame = &lastFrame
}

// hit implements renderer.
func (s *cursedRenderer) hit(mouse MouseMsg) []Msg {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastView == nil {
		return nil
	}

	if l := s.lastView.Layer; l != nil {
		if h, ok := l.(Hittable); ok {
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

// reset implements renderer.
func (s *cursedRenderer) reset() {
	s.mu.Lock()
	reset(s)
	s.mu.Unlock()
}

func reset(s *cursedRenderer) {
	scr := uv.NewTerminalRenderer(s.w, s.env)
	scr.SetColorProfile(s.profile)
	scr.SetRelativeCursor(s.lastView == nil || !s.lastView.AltScreen)
	scr.SetTabStops(s.width)
	scr.SetBackspace(s.backspace)
	scr.SetMapNewline(s.mapnl)
	scr.SetLogger(s.logger)
	if s.lastView != nil && s.lastView.AltScreen {
		scr.EnterAltScreen()
	} else {
		scr.ExitAltScreen()
	}
	if s.lastView != nil && s.lastView.Cursor != nil {
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
	if s.view.AltScreen || w != s.width {
		// We need to mark the screen for clear to force a redraw. However, we
		// only do so if we're using alt screen or the width has changed.
		// That's because redrawing is expensive and we can avoid it if the
		// width hasn't changed in inline mode. On the other hand, when using
		// alt screen mode, we always want to redraw because some terminals
		// would scroll the screen and our content would be lost.
		s.scr.Erase()
	}
	if s.view.AltScreen {
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
	if enable {
		s.scr.EnterAltScreen()
	} else {
		s.scr.ExitAltScreen()
	}
	s.scr.SetRelativeCursor(!enable)
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

func setProgressBar(s *cursedRenderer, pb *ProgressBar) {
	if pb == nil {
		_, _ = s.scr.WriteString(ansi.ResetProgressBar)
		return
	}

	var seq string
	switch pb.State {
	case ProgressBarNone:
		seq = ansi.ResetProgressBar
	case ProgressBarDefault:
		seq = ansi.SetProgressBar(pb.Value)
	case ProgressBarError:
		seq = ansi.SetErrorProgressBar(pb.Value)
	case ProgressBarIndeterminate:
		seq = ansi.SetIndeterminateProgressBar
	case ProgressBarWarning:
		seq = ansi.SetWarningProgressBar(pb.Value)
	}
	if seq != "" {
		_, _ = s.scr.WriteString(seq)
	}
}

// viewEquals reports whether two views are equal. It compares every field of
// the [View] struct except for the [View.Layer] field, which is compared using
// pointer equality.
func viewEquals(a, b View) bool {
	if a.AltScreen != b.AltScreen {
		return false
	}
	if a.WindowTitle != b.WindowTitle {
		return false
	}
	if a.MouseMode != b.MouseMode {
		return false
	}
	if a.ReportFocus != b.ReportFocus {
		return false
	}
	if a.DisableBracketedPasteMode != b.DisableBracketedPasteMode {
		return false
	}
	if a.DisableKeyEnhancements != b.DisableKeyEnhancements {
		return false
	}
	if a.KeyReleases != b.KeyReleases {
		return false
	}
	if a.UniformKeyLayout != b.UniformKeyLayout {
		return false
	}
	if (a.Cursor == nil) != (b.Cursor == nil) {
		return false
	}
	if a.Cursor != nil && b.Cursor != nil { //nolint:nestif
		if (a.Cursor.Color == nil) != (b.Cursor.Color == nil) {
			return false
		}
		if a.Cursor.Color != nil && b.Cursor.Color != nil {
			ar, ag, ab, aa := a.Cursor.Color.RGBA()
			br, bg, bb, ba := b.Cursor.Color.RGBA()
			if ar != br || ag != bg || ab != bb || aa != ba {
				return false
			}
		}
		if a.Cursor.Position != b.Cursor.Position || a.Cursor.Shape != b.Cursor.Shape || a.Cursor.Blink != b.Cursor.Blink {
			return false
		}
	}
	if (a.ProgressBar == nil) != (b.ProgressBar == nil) {
		return false
	}
	if a.ProgressBar != nil && b.ProgressBar != nil {
		if *a.ProgressBar != *b.ProgressBar {
			return false
		}
	}
	if (a.BackgroundColor == nil) != (b.BackgroundColor == nil) {
		return false
	}
	if a.BackgroundColor != nil && b.BackgroundColor != nil {
		ar, ag, ab, aa := a.BackgroundColor.RGBA()
		br, bg, bb, ba := b.BackgroundColor.RGBA()
		if ar != br || ag != bg || ab != bb || aa != ba {
			return false
		}
	}
	if (a.ForegroundColor == nil) != (b.ForegroundColor == nil) {
		return false
	}
	if a.ForegroundColor != nil && b.ForegroundColor != nil {
		ar, ag, ab, aa := a.ForegroundColor.RGBA()
		br, bg, bb, ba := b.ForegroundColor.RGBA()
		if ar != br || ag != bg || ab != bb || aa != ba {
			return false
		}
	}
	return true
}
