package tea

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"runtime"
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
	starting      bool // indicates whether the renderer is starting after being stopped
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

	// Mark that we're starting. This is used to restore some state when
	// starting the renderer again after it was stopped.
	s.starting = true

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
	// Enable modifyOtherKeys and Kitty keyboard protocol.
	// Both can coexist; terminals ignore what they don't support.
	_, _ = s.scr.WriteString(ansi.SetModifyOtherKeys2)

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
		// NOTE: The Kitty keyboard specs specify that the terminal should have
		// two registries for the main and alt screens. We disable keyboard
		// enhancements whenever we enter/exit alt screen mode in
		// [cursedRenderer.flush].
		// Here, we reset the keyboard protocol of the last screen used
		// assuming the other screen is already reset when we switched screens.
		_, _ = s.buf.WriteString(ansi.ResetModifyOtherKeys)
		_, _ = s.buf.WriteString(ansi.KittyKeyboard(0, 1))

		// Go to the bottom of the screen.
		// We need to go to the bottom of the screen regardless of whether
		// we're in alt screen mode or not to avoid leaving the cursor in the
		// middle in terminals that don't support alt screen mode.
		s.scr.MoveTo(0, s.cellbuf.Height()-1)
		_ = s.scr.Flush() // we need to flush to write the cursor movement
		if lv.AltScreen {
			enableAltScreen(s, false, true)
		} else {
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
	}

	if s.cellbuf.Method == ansi.GraphemeWidth {
		// Make sure to turn off Unicode mode (2027)
		_, _ = s.scr.WriteString(ansi.ResetModeUnicodeCore)
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
	if len(view.Content) == 0 {
		// If the component is nil, we should clear the screen buffer.
		frameArea.Max.Y = 0
	}

	content := uv.NewStyledString(view.Content)
	if !view.AltScreen {
		// We need to resizes the screen based on the frame height and
		// terminal width. This is because the frame height can change based on
		// the content of the frame. For example, if the frame contains a list
		// of items, the height of the frame will be the number of items in the
		// list. This is different from the alt screen buffer, which has a
		// fixed height and width.
		frameHeight := content.Height()
		if frameHeight != frameArea.Dy() {
			frameArea.Max.Y = frameHeight
		}
	}

	if !s.starting && !closing && s.lastView != nil && viewEquals(s.lastView, &view) && frameArea == s.cellbuf.Bounds() {
		// No changes, nothing to do.
		return nil
	}

	// We're no longer starting.
	s.starting = false

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
	content.Draw(s.cellbuf, s.cellbuf.Bounds())

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

		// Enable modifyOtherKeys and Kitty keyboard protocol.
		_, _ = s.scr.WriteString(ansi.SetModifyOtherKeys2)

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
	s.scr.Render(s.cellbuf.RenderBuffer)

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
		// We always disable keyboard enhancements when switching screens
		// because the terminal is expected to have two different keyboard
		// registries for main and alt screens.
		_, _ = buf.WriteString(ansi.ResetModifyOtherKeys)
		_, _ = buf.WriteString(ansi.KittyKeyboard(0, 1))
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
	scr.SetScrollOptim(runtime.GOOS != "windows") // disable scroll optimization on Windows due to bugs in some terminals
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
// Note that this writes to the buffer directly if write is true.
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

// setWidthMethod implements renderer.
func (s *cursedRenderer) setWidthMethod(method ansi.Method) {
	s.mu.Lock()
	if method == ansi.GraphemeWidth {
		// Turn on Unicode mode (2027) for accurate grapheme width calculation.
		// This is needed for proper rendering of wide characters and emojis.
		_, _ = s.scr.WriteString(ansi.SetModeUnicodeCore)
	} else if s.cellbuf.Method == ansi.GraphemeWidth {
		// Turn off Unicode mode if we're switching away from grapheme width
		// calculation to avoid issues with some terminals that might still be
		// in Unicode mode and render characters incorrectly.
		_, _ = s.scr.WriteString(ansi.ResetModeUnicodeCore)
	}
	s.cellbuf.Method = method
	s.mu.Unlock()
}

// insertAbove implements renderer.
func (s *cursedRenderer) insertAbove(str string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(str) == 0 {
		return nil
	}

	var sb strings.Builder
	w, h := s.cellbuf.Width(), s.cellbuf.Height()
	_, y := s.scr.Position()

	// We need to scroll the screen up by the number of lines in the queue.
	sb.WriteByte('\r')
	down := h - y - 1
	if down > 0 {
		sb.WriteString(ansi.CursorDown(down))
	}

	lines := strings.Split(str, "\n")
	offset := len(lines)
	for _, line := range lines {
		lineWidth := ansi.StringWidth(line)
		if w > 0 && lineWidth > w {
			offset += (lineWidth / w)
		}
	}

	// Scroll the screen up by the offset to make room for the new lines.
	sb.WriteString(strings.Repeat("\n", offset))

	// XXX: Now go to the top of the screen, insert new lines, and write
	// the queued strings. It is important to use [Screen.moveCursor]
	// instead of [Screen.move] because we don't want to perform any checks
	// on the cursor position.
	up := offset + h - 1
	sb.WriteString(ansi.CursorUp(up))
	sb.WriteString(ansi.InsertLine(offset))
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString(ansi.EraseLineRight)
		sb.WriteString("\r\n")
	}

	s.scr.SetPosition(0, 0)

	if s.logger != nil {
		s.logger.Printf("insert above: %q", sb.String())
	}

	_, err := io.WriteString(s.w, sb.String())
	return err
}

// onMouse implements renderer.
func (s *cursedRenderer) onMouse(m MouseMsg) Cmd {
	if s.lastView != nil && s.lastView.OnMouse != nil {
		return s.lastView.OnMouse(m)
	}
	return nil
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

func viewEquals(a, b *View) bool {
	if a == nil || b == nil {
		return false
	}

	if a.Content != b.Content ||
		a.AltScreen != b.AltScreen ||
		a.DisableBracketedPasteMode != b.DisableBracketedPasteMode ||
		a.ReportFocus != b.ReportFocus ||
		a.MouseMode != b.MouseMode ||
		a.WindowTitle != b.WindowTitle ||
		a.ForegroundColor != b.ForegroundColor ||
		a.BackgroundColor != b.BackgroundColor ||
		a.KeyboardEnhancements != b.KeyboardEnhancements {
		return false
	}

	if (a.Cursor == nil) != (b.Cursor == nil) {
		return false
	}
	if a.Cursor != nil && b.Cursor != nil {
		if a.Cursor.X != b.Cursor.X ||
			a.Cursor.Y != b.Cursor.Y ||
			a.Cursor.Shape != b.Cursor.Shape ||
			a.Cursor.Blink != b.Cursor.Blink ||
			a.Cursor.Color != b.Cursor.Color {
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

	return true
}
