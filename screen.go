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
	return enableMode(ansi.AltScreenBufferMode.String())
}

// ExitAltScreen is a special command that tells the Bubble Tea program to exit
// the alternate screen buffer. This command should be used to exit the
// alternate screen buffer while the program is running.
//
// Note that the alternate screen buffer will be automatically exited when the
// program quits.
func ExitAltScreen() Msg {
	return disableMode(ansi.AltScreenBufferMode.String())
}

// EnableMouseCellMotion is a special command that enables mouse click,
// release, and wheel events. Mouse movement events are also captured if
// a mouse button is pressed (i.e., drag events).
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. Use the WithMouseCellMotion ProgramOption instead.
func EnableMouseCellMotion() Msg {
	return sequenceMsg{
		func() Msg { return enableMode(ansi.MouseCellMotionMode.String()) },
		func() Msg { return enableMode(ansi.MouseSgrExtMode.String()) },
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
		func() Msg { return enableMode(ansi.MouseAllMotionMode.String()) },
		func() Msg { return enableMode(ansi.MouseSgrExtMode.String()) },
	}
}

// DisableMouse is a special command that stops listening for mouse events.
func DisableMouse() Msg {
	return sequenceMsg{
		func() Msg { return disableMode(ansi.MouseCellMotionMode.String()) },
		func() Msg { return disableMode(ansi.MouseAllMotionMode.String()) },
		func() Msg { return disableMode(ansi.MouseSgrExtMode.String()) },
	}
}

// HideCursor is a special command for manually instructing Bubble Tea to hide
// the cursor. In some rare cases, certain operations will cause the terminal
// to show the cursor, which is normally hidden for the duration of a Bubble
// Tea program's lifetime. You will most likely not need to use this command.
func HideCursor() Msg {
	return disableMode(ansi.CursorEnableMode.String())
}

// ShowCursor is a special command for manually instructing Bubble Tea to show
// the cursor.
func ShowCursor() Msg {
	return enableMode(ansi.CursorEnableMode.String())
}

// EnableBracketedPaste is a special command that tells the Bubble Tea program
// to accept bracketed paste input.
//
// Note that bracketed paste will be automatically disabled when the
// program quits.
func EnableBracketedPaste() Msg {
	return enableMode(ansi.BracketedPasteMode.String())
}

// DisableBracketedPaste is a special command that tells the Bubble Tea program
// to accept bracketed paste input.
//
// Note that bracketed paste will be automatically disabled when the
// program quits.
func DisableBracketedPaste() Msg {
	return disableMode(ansi.BracketedPasteMode.String())
}

// EnableGraphemeClustering is a special command that tells the Bubble Tea
// program to enable grapheme clustering. This is enabled by default.
func EnableGraphemeClustering() Msg {
	return enableMode(ansi.GraphemeClusteringMode.String())
}

// DisableGraphemeClustering is a special command that tells the Bubble Tea
// program to disable grapheme clustering. This mode will be disabled
// automatically when the program quits.
func DisableGraphemeClustering() Msg {
	return disableMode(ansi.GraphemeClusteringMode.String())
}

// EnabledReportFocus is a special command that tells the Bubble Tea program
// to enable focus reporting.
func EnabledReportFocus() Msg { return enableMode(ansi.ReportFocusMode.String()) }

// DisabledReportFocus is a special command that tells the Bubble Tea program
// to disable focus reporting.
func DisabledReportFocus() Msg { return disableMode(ansi.ReportFocusMode.String()) }

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
