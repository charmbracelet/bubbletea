package tea

import (
	"bytes"
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
	w             io.Writer
	buf           bytes.Buffer // updates buffer to be flushed to [w]
	scr           *uv.TerminalRenderer
	cellbuf       uv.ScreenBuffer
	lastView      *View
	env           []string
	term          string // the terminal type $TERM
	width, height int
	mu            sync.Mutex
	profile       colorprofile.Profile
	logger        uv.Logger
	view          View
	hardTabs      bool // whether to use hard tabs to optimize cursor movements
	backspace     bool // whether to use backspace to optimize cursor movements
	mapnl         bool
	syncdUpdates  bool // whether to use synchronized output mode for updates
	prependLines  []string
}

var _ renderer = &cursedRenderer{}

func newCursedRenderer(w io.Writer, env []string, width, height int) (s *cursedRenderer) {
	s = new(cursedRenderer)
	s.w = w
	s.env = env
	s.term = uv.Environ(env).Getenv("TERM")
	s.width, s.height = width, height // This needs to happen before [cursedRenderer.reset].
	s.cellbuf = uv.NewScreenBuffer(s.width, s.height)
	reset(s)
	return
}

// setLogger sets the logger for the renderer.
func (s *cursedRenderer) setLogger(logger uv.Logger) {
	s.mu.Lock()
	s.logger = logger
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

	if s.lastView == nil {
		return
	}

	if s.lastView.AltScreen {
		enableAltScreen(s, true, true)
	}
	enableTextCursor(s, s.lastView.Cursor != nil)
	if s.lastView.Cursor != nil {
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
		_, _ = s.scr.WriteString(ansi.SetModeBracketedPaste)
	}
	if s.lastView.ReportFocus {
		_, _ = s.scr.WriteString(ansi.SetModeFocusEvent)
	}
	switch s.lastView.MouseMode {
	case MouseModeNone:
	case MouseModeCellMotion:
		_, _ = s.scr.WriteString(ansi.SetModeMouseButtonEvent + ansi.SetModeMouseExtSgr)
	case MouseModeAllMotion:
		_, _ = s.scr.WriteString(ansi.SetModeMouseAnyEvent + ansi.SetModeMouseExtSgr)
	}
	if s.lastView.WindowTitle != "" {
		_, _ = s.scr.WriteString(ansi.SetWindowTitle(s.lastView.WindowTitle))
	}
	if s.lastView.ProgressBar != nil {
		setProgressBar(s, s.lastView.ProgressBar)
	}
	kittyFlags := ansi.KittyDisambiguateEscapeCodes
	if s.lastView.KeyboardEnhancements.ReportEventTypes {
		kittyFlags |= ansi.KittyReportEventTypes
	}
	_, _ = s.scr.WriteString(ansi.KittyKeyboard(kittyFlags, 1))
}

// close implements renderer.
func (s *cursedRenderer) close() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Exit the altScreen and show cursor before closing. It's important that
	// we don't change the [cursedRenderer] altScreen and cursorHidden states
	// so that we can restore them when we start the renderer again. This is
	// used when the user suspends the program and then resumes it.
	if lv := s.lastView; lv != nil { //nolint:nestif
		if lv.AltScreen {
			enableAltScreen(s, false, true)
		} else {
			// Go to the bottom of the screen.
			s.scr.MoveTo(0, s.cellbuf.Height()-1)
			_, _ = s.scr.WriteString(ansi.EraseScreenBelow)
		}
		if lv.Cursor == nil {
			enableTextCursor(s, true)
		}
		if !lv.DisableBracketedPasteMode {
			_, _ = s.scr.WriteString(ansi.ResetModeBracketedPaste)
		}
		if lv.ReportFocus {
			_, _ = s.scr.WriteString(ansi.ResetModeFocusEvent)
		}
		switch lv.MouseMode {
		case MouseModeNone:
		case MouseModeCellMotion, MouseModeAllMotion:
			_, _ = s.scr.WriteString(ansi.ResetModeMouseButtonEvent +
				ansi.ResetModeMouseAnyEvent +
				ansi.ResetModeMouseExtSgr)
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
		if lv.ProgressBar != nil && lv.ProgressBar.State != ProgressBarNone {
			_, _ = s.scr.WriteString(ansi.ResetProgressBar)
		}

		// NOTE: This needs to happen after we exit the alt screen.
		_, _ = s.scr.WriteString(ansi.KittyKeyboard(0, 1))
	}

	if err := s.scr.Flush(); err != nil {
		return fmt.Errorf("bubbletea: error closing screen writer: %w", err)
	}

	if s.buf.Len() > 0 {
		if s.logger != nil {
			s.logger.Printf("output: %q", s.buf.String())
		}
		if _, err := io.Copy(s.w, &s.buf); err != nil {
			return fmt.Errorf("bubbletea: error writing to screen: %w", err)
		}
		s.buf.Reset()
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

// flush implements renderer.
func (s *cursedRenderer) flush(closing bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	view := s.view
	frameArea := uv.Rect(0, 0, s.width, s.height)
	if view.Content == nil {
		// If the component is nil, we should clear the screen buffer.
		frameArea.Max.Y = 0
	}

	if !view.AltScreen {
		// We need to resizes the screen based on the frame height and
		// terminal width. This is because the frame height can change based on
		// the content of the frame. For example, if the frame contains a list
		// of items, the height of the frame will be the number of items in the
		// list. This is different from the alt screen buffer, which has a
		// fixed height and width.
		frameHeight := frameArea.Dy()
		switch l := view.Content.(type) {
		case interface{ Height() int }:
			// This covers [uv.StyledString] and [lipgloss.Canvas].
			frameHeight = l.Height()
		case interface{ Bounds() uv.Rectangle }:
			frameHeight = l.Bounds().Dy()
		}

		if frameHeight != frameArea.Dy() {
			frameArea.Max.Y = frameHeight
		}
	}

	if s.lastView != nil && *s.lastView == view && frameArea == s.cellbuf.Bounds() {
		// No changes, nothing to do.
		return nil
	}

	if frameArea != s.cellbuf.Bounds() {
		s.scr.Erase() // Force a full redraw to avoid artifacts.

		// We need to reset the touched lines buffer to match the new height.
		s.cellbuf.Touched = nil

		// Resize the screen buffer to match the frame area. This is necessary
		// to ensure that the screen buffer is the same size as the frame area
		// and to avoid rendering issues when the frame area is smaller than
		// the screen buffer.
		s.cellbuf.Resize(frameArea.Dx(), frameArea.Dy())
	}

	// Clear our screen buffer before copying the new frame into it to ensure
	// we erase any old content.
	s.cellbuf.Clear()
	if view.Content != nil {
		view.Content.Draw(s.cellbuf, s.cellbuf.Bounds())
	}

	// If the frame height is greater than the screen height, we drop the
	// lines from the top of the buffer.
	if frameHeight := frameArea.Dy(); frameHeight > s.height {
		s.cellbuf.Lines = s.cellbuf.Lines[frameHeight-s.height:]
	}

	// Alt screen mode.
	shouldUpdateAltScreen := (s.lastView == nil && view.AltScreen) || (s.lastView != nil && s.lastView.AltScreen != view.AltScreen)
	if shouldUpdateAltScreen {
		// We want to enter/exit altscreen mode but defer writing the actual
		// sequences until we flush the rest of the updates. This is because we
		// control the cursor visibility and we need to ensure that happens
		// after entering/exiting alt screen mode. Some terminals have
		// different cursor visibility states for main and alt screen modes and
		// this ensures we handle that correctly.
		enableAltScreen(s, view.AltScreen, false)
	}

	// Push prepended lines if any.
	if len(s.prependLines) > 0 {
		for _, line := range s.prependLines {
			prependLine(s, line)
		}
		s.prependLines = s.prependLines[:0]
	}

	// bracketed paste mode.
	if s.lastView == nil || view.DisableBracketedPasteMode != s.lastView.DisableBracketedPasteMode {
		if !view.DisableBracketedPasteMode {
			_, _ = s.scr.WriteString(ansi.SetModeBracketedPaste)
		} else if s.lastView != nil {
			_, _ = s.scr.WriteString(ansi.ResetModeBracketedPaste)
		}
	}

	// report focus events mode.
	if s.lastView == nil || s.lastView.ReportFocus != view.ReportFocus {
		if view.ReportFocus {
			_, _ = s.scr.WriteString(ansi.SetModeFocusEvent)
		} else if s.lastView != nil {
			_, _ = s.scr.WriteString(ansi.ResetModeFocusEvent)
		}
	}

	// mouse events mode.
	if s.lastView == nil || view.MouseMode != s.lastView.MouseMode {
		switch view.MouseMode {
		case MouseModeNone:
			if s.lastView != nil && s.lastView.MouseMode != MouseModeNone {
				_, _ = s.scr.WriteString(ansi.ResetModeMouseButtonEvent +
					ansi.ResetModeMouseAnyEvent +
					ansi.ResetModeMouseExtSgr)
			}
		case MouseModeCellMotion:
			if s.lastView != nil && s.lastView.MouseMode == MouseModeAllMotion {
				_, _ = s.scr.WriteString(ansi.ResetModeMouseAnyEvent)
			}
			_, _ = s.scr.WriteString(ansi.SetModeMouseButtonEvent + ansi.SetModeMouseExtSgr)
		case MouseModeAllMotion:
			if s.lastView != nil && s.lastView.MouseMode == MouseModeCellMotion {
				_, _ = s.scr.WriteString(ansi.ResetModeMouseButtonEvent)
			}
			_, _ = s.scr.WriteString(ansi.SetModeMouseAnyEvent + ansi.SetModeMouseExtSgr)
		}
	}

	// Set window title.
	if s.lastView == nil || view.WindowTitle != s.lastView.WindowTitle {
		if s.lastView != nil || view.WindowTitle != "" {
			_, _ = s.scr.WriteString(ansi.SetWindowTitle(view.WindowTitle))
		}
	}

	// kitty keyboard protocol
	if s.lastView == nil || view.KeyboardEnhancements != s.lastView.KeyboardEnhancements ||
		view.AltScreen != s.lastView.AltScreen {
		// NOTE: We need to reset the keyboard protocol when switching
		// between main and alt screen. This is because the specs specify
		// two different states for the main and alt screen.
		kittyFlags := ansi.KittyDisambiguateEscapeCodes // always enable basic key disambiguation
		if view.KeyboardEnhancements.ReportEventTypes {
			kittyFlags |= ansi.KittyReportEventTypes
		}
		_, _ = s.scr.WriteString(ansi.KittyKeyboard(kittyFlags, 1))
		if !closing {
			// Request keyboard enhancements when they change
			_, _ = s.scr.WriteString(ansi.RequestKittyKeyboard)
		}
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

	// Set cursor shape and blink if set.
	var ccStyle, lcStyle int
	var lcur *Cursor
	ccur := view.Cursor
	if lv := s.lastView; lv != nil {
		lcur = lv.Cursor
	}
	if ccur != nil {
		ccStyle = encodeCursorStyle(ccur.Shape, ccur.Blink)
	}
	if lcur != nil {
		lcStyle = encodeCursorStyle(lcur.Shape, lcur.Blink)
	}
	if ccStyle != lcStyle {
		_, _ = s.scr.WriteString(ansi.SetCursorStyle(ccStyle))
	}

	// Render progress bar if it's changed.
	if (s.lastView == nil && view.ProgressBar != nil && view.ProgressBar.State != ProgressBarNone) ||
		(s.lastView != nil && (s.lastView.ProgressBar == nil) != (view.ProgressBar == nil)) ||
		(s.lastView != nil && s.lastView.ProgressBar != nil && view.ProgressBar != nil && *s.lastView.ProgressBar != *view.ProgressBar) {
		// Render or clear the progress bar if it was added or removed.
		setProgressBar(s, view.ProgressBar)
	}

	// Render and queue changes to the screen buffer.
	s.scr.Render(s.cellbuf.Buffer)

	if cur := view.Cursor; cur != nil {
		// MoveTo must come after [uv.TerminalRenderer.Render] because the
		// cursor position might get updated during rendering.
		s.scr.MoveTo(view.Cursor.X, view.Cursor.Y)
	} else if !view.AltScreen {
		// We don't want the cursor to be dangling at the end of the line in
		// inline mode because it can cause unwanted line wraps in some
		// terminals. So we move it to the beginning of the next line if
		// necessary.
		// This is only needed when the cursor is hidden because when it's
		// visible, we already set its position above.
		x, y := s.scr.Position()
		if x >= s.width-1 {
			s.scr.MoveTo(0, y)
		}
	}

	if err := s.scr.Flush(); err != nil {
		return fmt.Errorf("bubbletea: error flushing screen writer: %w", err)
	}

	// Check if we have any render updates to flush.
	hasUpdates := s.buf.Len() > 0

	// Cursor visibility.
	didShowCursor := s.lastView != nil && s.lastView.Cursor != nil
	showCursor := view.Cursor != nil
	hideCursor := !showCursor
	shouldUpdateCursorVis := (s.lastView == nil || didShowCursor != showCursor) || shouldUpdateAltScreen

	// Build final output buffer with synchronized output or hide/show cursor
	// updates. But first, enter/exit alt screen mode if needed.
	//
	// Here, we have two scenarios:
	// 1. Synchronized output updates are supported. In this case, we want to
	//    wrap all updates, unless it's just a cursor visibility change, in
	//    synchronized output mode. This is because synchronized output mode
	//    takes care of rendering the updates atomically. In the case of
	//    just a cursor visibility change, we don't need to enter
	//    synchronized output mode because it's just a single sequence to
	//    flush out to the terminal.
	//
	// 2. We don't have synchronized output updates support. In this case, and
	//    if the cursor is visible or should be visible, we wrap the updates
	//    with hide/show cursor sequences to try and mitigate cursor
	//    flickering. This is terminal dependent and may still result in
	//    flickering in some terminals. It's the best effort we can do instead
	//    of showing the cursor flying around the screen during updates.

	var buf bytes.Buffer
	if shouldUpdateAltScreen {
		if view.AltScreen {
			// Entering alt screen mode.
			buf.WriteString(ansi.SetModeAltScreenSaveCursor)
		} else {
			// Exiting alt screen mode.
			buf.WriteString(ansi.ResetModeAltScreenSaveCursor)
		}
	}

	if s.syncdUpdates {
		if hasUpdates {
			// We have synchronized output updates enabled.
			buf.WriteString(ansi.SetModeSynchronizedOutput)
		}
		if shouldUpdateCursorVis && hideCursor {
			// Do we need to update the cursor visibility to hidden? If so, do
			// it here before writing any updates to the buffer.
			_, _ = buf.WriteString(ansi.ResetModeTextCursorEnable)
		}
	} else if (shouldUpdateCursorVis && hideCursor) || (hasUpdates && showCursor && didShowCursor) {
		_, _ = buf.WriteString(ansi.ResetModeTextCursorEnable)
	}

	if hasUpdates {
		buf.Write(s.buf.Bytes())
	}

	if s.syncdUpdates {
		if shouldUpdateCursorVis && showCursor {
			// Do we need to update the cursor visibility to visible? If so, do
			// it here after writing any updates to the buffer.
			_, _ = buf.WriteString(ansi.SetModeTextCursorEnable)
		}
		if hasUpdates {
			// Close synchronized output mode.
			buf.WriteString(ansi.ResetModeSynchronizedOutput)
		}
	} else if (shouldUpdateCursorVis && showCursor) || (hasUpdates && showCursor && didShowCursor) {
		_, _ = buf.WriteString(ansi.SetModeTextCursorEnable)
	}

	// Reset internal screen renderer buffer.
	s.buf.Reset()

	// If our updates flush buffer has content, write it to the output writer.
	if buf.Len() > 0 {
		if s.logger != nil {
			s.logger.Printf("output: %q", buf.String())
		}
		if _, err := io.Copy(s.w, &buf); err != nil {
			return fmt.Errorf("bubbletea: error flushing update to the writer: %w", err)
		}
	}

	s.lastView = &view

	return nil
}

// render implements renderer.
func (s *cursedRenderer) render(v View) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.view = v
}

// hit implements renderer.
func (s *cursedRenderer) hit(mouse MouseMsg) []Msg {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastView == nil {
		return nil
	}

	if l := s.lastView.Content; l != nil {
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
	s.buf.Reset()
	scr := uv.NewTerminalRenderer(&s.buf, s.env)
	scr.SetColorProfile(s.profile)
	scr.SetRelativeCursor(true) // Always start in inline mode
	scr.SetFullscreen(false)    // Always start in inline mode
	scr.SetTabStops(s.width)
	scr.SetBackspace(s.backspace)
	scr.SetMapNewline(s.mapnl)
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
	// We need to mark the screen for clear to force a redraw. However, we
	// only do so if we're using alt screen or the width has changed.
	// That's because redrawing is expensive and we can avoid it if the
	// width hasn't changed in inline mode. On the other hand, when using
	// alt screen mode, we always want to redraw because some terminals
	// would scroll the screen and our content would be lost.
	s.scr.Erase()
	s.width, s.height = w, h
	s.scr.Resize(s.width, s.height)
	s.mu.Unlock()
}

// clearScreen implements renderer.
func (s *cursedRenderer) clearScreen() {
	s.mu.Lock()
	// Move the cursor to the top left corner of the screen and trigger a full
	// screen redraw.
	s.scr.MoveTo(0, 0)
	s.scr.Erase()
	s.mu.Unlock()
}

// enableAltScreen sets the alt screen mode.
func enableAltScreen(s *cursedRenderer, enable bool, write bool) {
	if enable {
		enterAltScreen(s, write)
	} else {
		exitAltScreen(s, write)
	}
}

func enterAltScreen(s *cursedRenderer, write bool) {
	s.scr.SaveCursor()
	if write {
		s.buf.WriteString(ansi.SetModeAltScreenSaveCursor)
	}
	s.scr.SetFullscreen(true)
	s.scr.SetRelativeCursor(false)
	s.scr.Erase()
}

func exitAltScreen(s *cursedRenderer, write bool) {
	s.scr.Erase()
	s.scr.SetRelativeCursor(true)
	s.scr.SetFullscreen(false)
	if write {
		s.buf.WriteString(ansi.ResetModeAltScreenSaveCursor)
	}
	s.scr.RestoreCursor()
}

// enableTextCursor sets the text cursor mode.
func enableTextCursor(s *cursedRenderer, enable bool) {
	if enable {
		_, _ = s.scr.WriteString(ansi.SetModeTextCursorEnable)
	} else {
		_, _ = s.scr.WriteString(ansi.ResetModeTextCursorEnable)
	}
}

// setSyncdUpdates implements renderer.
func (s *cursedRenderer) setSyncdUpdates(syncd bool) {
	s.mu.Lock()
	s.syncdUpdates = syncd
	s.mu.Unlock()
}

// insertAbove implements renderer.
func (s *cursedRenderer) insertAbove(lines string) {
	s.mu.Lock()
	s.prependLines = append(s.prependLines, strings.Split(lines, "\n")...)
	s.mu.Unlock()
}

func prependLine(s *cursedRenderer, line string) {
	strLines := strings.Split(line, "\n")
	for i, line := range strLines {
		// Always erase to the right of the line to avoid possible artifacts.
		strLines[i] = line + ansi.EraseLineRight
	}
	s.scr.PrependString(s.cellbuf.Buffer, strings.Join(strLines, "\n"))
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
