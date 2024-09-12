package tea

import "github.com/charmbracelet/x/ansi"

// WindowSizeMsg is used to report the terminal size. It's sent to Update once
// initially and then on every terminal resize. Note that Windows does not
// have support for reporting when resizes occur as it does not support the
// SIGWINCH signal.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// ClearScreen is a special command that tells the program to clear the screen
// before the next update. This can be used to move the cursor to the top left
// of the screen and clear visual clutter when the alt screen is not in use.
//
// Note that it should never be necessary to call ClearScreen() for regular
// redraws.
func ClearScreen() Msg {
	return clearScreenMsg{}
}

// clearScreenMsg is an internal message that signals to clear the screen.
// You can send a clearScreenMsg with ClearScreen.
type clearScreenMsg struct{}

// EnterAltScreen is a special command that tells the Bubble Tea program to
// enter the alternate screen buffer.
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. To initialize your program with the altscreen enabled
// use the WithAltScreen ProgramOption instead.
func EnterAltScreen() Msg {
	return enableMode(ansi.AltScreenBufferMode)
}

// ExitAltScreen is a special command that tells the Bubble Tea program to exit
// the alternate screen buffer. This command should be used to exit the
// alternate screen buffer while the program is running.
//
// Note that the alternate screen buffer will be automatically exited when the
// program quits.
func ExitAltScreen() Msg {
	return disableMode(ansi.AltScreenBufferMode)
}

// EnableMouseCellMotion is a special command that enables mouse click,
// release, and wheel events. Mouse movement events are also captured if
// a mouse button is pressed (i.e., drag events).
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. Use the WithMouseCellMotion ProgramOption instead.
func EnableMouseCellMotion() Msg {
	return sequenceMsg{
		func() Msg { return enableMode(ansi.MouseCellMotionMode) },
		func() Msg { return enableMode(ansi.MouseSgrExtMode) },
	}
}

// EnableMouseAllMotion is a special command that enables mouse click, release,
// wheel, and motion events, which are delivered regardless of whether a mouse
// button is pressed, effectively enabling support for hover interactions.
//
// Many modern terminals support this, but not all. If in doubt, use
// EnableMouseCellMotion instead.
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. Use the WithMouseAllMotion ProgramOption instead.
func EnableMouseAllMotion() Msg {
	return sequenceMsg{
		func() Msg { return enableMode(ansi.MouseAllMotionMode) },
		func() Msg { return enableMode(ansi.MouseSgrExtMode) },
	}
}

// DisableMouse is a special command that stops listening for mouse events.
func DisableMouse() Msg {
	return sequenceMsg{
		func() Msg { return disableMode(ansi.MouseCellMotionMode) },
		func() Msg { return disableMode(ansi.MouseAllMotionMode) },
		func() Msg { return disableMode(ansi.MouseSgrExtMode) },
	}
}

// HideCursor is a special command for manually instructing Bubble Tea to hide
// the cursor. In some rare cases, certain operations will cause the terminal
// to show the cursor, which is normally hidden for the duration of a Bubble
// Tea program's lifetime. You will most likely not need to use this command.
func HideCursor() Msg {
	return disableMode(ansi.CursorVisibilityMode)
}

// ShowCursor is a special command for manually instructing Bubble Tea to show
// the cursor.
func ShowCursor() Msg {
	return enableMode(ansi.CursorVisibilityMode)
}

// EnableBracketedPaste is a special command that tells the Bubble Tea program
// to accept bracketed paste input.
//
// Note that bracketed paste will be automatically disabled when the
// program quits.
func EnableBracketedPaste() Msg {
	return enableMode(ansi.BracketedPasteMode)
}

// DisableBracketedPaste is a special command that tells the Bubble Tea program
// to accept bracketed paste input.
//
// Note that bracketed paste will be automatically disabled when the
// program quits.
func DisableBracketedPaste() Msg {
	return disableMode(ansi.BracketedPasteMode)
}

// EnableGraphemeClustering is a special command that tells the Bubble Tea
// program to enable grapheme clustering. This is enabled by default.
func EnableGraphemeClustering() Msg {
	return enableMode(ansi.GraphemeClusteringMode)
}

// DisableGraphemeClustering is a special command that tells the Bubble Tea
// program to disable grapheme clustering. This mode will be disabled
// automatically when the program quits.
func DisableGraphemeClustering() Msg {
	return disableMode(ansi.GraphemeClusteringMode)
}

// EnabledReportFocus is a special command that tells the Bubble Tea program
// to enable focus reporting.
func EnabledReportFocus() Msg { return enableMode(ansi.ReportFocusMode) }

// DisabledReportFocus is a special command that tells the Bubble Tea program
// to disable focus reporting.
func DisabledReportFocus() Msg { return disableMode(ansi.ReportFocusMode) }

// enableModeMsg is an internal message that signals to set a terminal mode.
type enableModeMsg string

// enableMode is an internal command that signals to set a terminal mode.
func enableMode(mode string) Msg {
	return enableModeMsg(mode)
}

// disableModeMsg is an internal message that signals to unset a terminal mode.
type disableModeMsg string

// disableMode is an internal command that signals to unset a terminal mode.
func disableMode(mode string) Msg {
	return disableModeMsg(mode)
}

// EnterAltScreen enters the alternate screen buffer, which consumes the entire
// terminal window. ExitAltScreen will return the terminal to its former state.
//
// Deprecated: Use the WithAltScreen ProgramOption instead.
func (p *Program) EnterAltScreen() {
	if p.renderer != nil {
		_ = p.renderer.update(enableMode(ansi.AltScreenBufferMode))
	} else {
		p.startupOptions |= withAltScreen
	}
}

// ExitAltScreen exits the alternate screen buffer.
//
// Deprecated: The altscreen will exited automatically when the program exits.
func (p *Program) ExitAltScreen() {
	if p.renderer != nil {
		_ = p.renderer.update(disableMode(ansi.AltScreenBufferMode))
	} else {
		p.startupOptions &^= withAltScreen
	}
}

// EnableMouseCellMotion enables mouse click, release, wheel and motion events
// if a mouse button is pressed (i.e., drag events).
//
// Deprecated: Use the WithMouseCellMotion ProgramOption instead.
func (p *Program) EnableMouseCellMotion() {
	p.execute(ansi.EnableMouseCellMotion)
}

// DisableMouseCellMotion disables Mouse Cell Motion tracking. This will be
// called automatically when exiting a Bubble Tea program.
//
// Deprecated: The mouse will automatically be disabled when the program exits.
func (p *Program) DisableMouseCellMotion() {
	p.execute(ansi.DisableMouseCellMotion)
}

// EnableMouseAllMotion enables mouse click, release, wheel and motion events,
// regardless of whether a mouse button is pressed. Many modern terminals
// support this, but not all.
//
// Deprecated: Use the WithMouseAllMotion ProgramOption instead.
func (p *Program) EnableMouseAllMotion() {
	p.execute(ansi.EnableMouseAllMotion)
}

// DisableMouseAllMotion disables All Motion mouse tracking. This will be
// called automatically when exiting a Bubble Tea program.
//
// Deprecated: The mouse will automatically be disabled when the program exits.
func (p *Program) DisableMouseAllMotion() {
	p.execute(ansi.DisableMouseAllMotion)
}

// SetWindowTitle sets the terminal window title.
//
// Deprecated: Use the SetWindowTitle command instead.
func (p *Program) SetWindowTitle(title string) {
	p.execute(ansi.SetWindowTitle(title))
}
