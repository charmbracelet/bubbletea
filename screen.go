package tea

// WindowSizeMsg is used to report the terminal size. It's sent to Update once
// initially and then on every terminal resize. Note that Windows does not
// have support for reporting when resizes occur as it does not support the
// SIGWINCH signal.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// clearScreenMsg is an internal message that signals to clear the screen.
// You can send a clearScreenMsg with ClearScreen.
type clearScreenMsg struct{}

// ClearScreen is a special command that tells the program to clear the screen
// before the next update. This can be used to move the cursor to the top left
// of the screen and clear visual clutter when the alt screen is not in use.
//
// Note that it should never be necessary to call ClearScreen() for regular
// redraws.
func ClearScreen() Msg {
	return clearScreenMsg{}
}

// enterAltScreenMsg in an internal message signals that the program should
// enter alternate screen buffer. You can send a enterAltScreenMsg with
// EnterAltScreen.
type enterAltScreenMsg struct{}

// EnterAltScreen is a special command that tells the Bubble Tea program to
// enter the alternate screen buffer.
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. To initialize your program with the altscreen enabled
// use the WithAltScreen ProgramOption instead.
func EnterAltScreen() Msg {
	return enterAltScreenMsg{}
}

// exitAltScreenMsg in an internal message signals that the program should exit
// alternate screen buffer. You can send a exitAltScreenMsg with ExitAltScreen.
type exitAltScreenMsg struct{}

// ExitAltScreen is a special command that tells the Bubble Tea program to exit
// the alternate screen buffer. This command should be used to exit the
// alternate screen buffer while the program is running.
//
// Note that the alternate screen buffer will be automatically exited when the
// program quits.
func ExitAltScreen() Msg {
	return exitAltScreenMsg{}
}

// enableMouseCellMotionMsg is a special command that signals to start
// listening for "cell motion" type mouse events (ESC[?1002l). To send an
// enableMouseCellMotionMsg, use the EnableMouseCellMotion command.
type enableMouseCellMotionMsg struct{}

// EnableMouseCellMotion is a special command that enables mouse click,
// release, and wheel events. Mouse movement events are also captured if
// a mouse button is pressed (i.e., drag events).
//
// Because commands run asynchronously, this command should not be used in your
// model's Init function. Use the WithMouseCellMotion ProgramOption instead.
func EnableMouseCellMotion() Msg {
	return enableMouseCellMotionMsg{}
}

// enableMouseAllMotionMsg is a special command that signals to start listening
// for "all motion" type mouse events (ESC[?1003l). To send an
// enableMouseAllMotionMsg, use the EnableMouseAllMotion command.
type enableMouseAllMotionMsg struct{}

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
	return enableMouseAllMotionMsg{}
}

// disableMouseMsg is an internal message that signals to stop listening
// for mouse events. To send a disableMouseMsg, use the DisableMouse command.
type disableMouseMsg struct{}

// DisableMouse is a special command that stops listening for mouse events.
func DisableMouse() Msg {
	return disableMouseMsg{}
}

// hideCursorMsg is an internal command used to hide the cursor. You can send
// this message with HideCursor.
type hideCursorMsg struct{}

// HideCursor is a special command for manually instructing Bubble Tea to hide
// the cursor. In some rare cases, certain operations will cause the terminal
// to show the cursor, which is normally hidden for the duration of a Bubble
// Tea program's lifetime. You will most likely not need to use this command.
func HideCursor() Msg {
	return hideCursorMsg{}
}

// showCursorMsg is an internal command used to show the cursor. You can send
// this message with ShowCursor.
type showCursorMsg struct{}

// ShowCursor is a special command for manually instructing Bubble Tea to show
// the cursor.
func ShowCursor() Msg {
	return showCursorMsg{}
}

// EnterAltScreen enters the alternate screen buffer, which consumes the entire
// terminal window. ExitAltScreen will return the terminal to its former state.
//
// Deprecated: Use the WithAltScreen ProgramOption instead.
func (p *Program[M]) EnterAltScreen() {
	if p.renderer != nil {
		p.renderer.enterAltScreen()
	}
}

// ExitAltScreen exits the alternate screen buffer.
//
// Deprecated: The altscreen will exited automatically when the program exits.
func (p *Program[M]) ExitAltScreen() {
	if p.renderer != nil {
		p.renderer.exitAltScreen()
	}
}

// EnableMouseCellMotion enables mouse click, release, wheel and motion events
// if a mouse button is pressed (i.e., drag events).
//
// Deprecated: Use the WithMouseCellMotion ProgramOption instead.
func (p *Program[M]) EnableMouseCellMotion() {
	p.renderer.enableMouseCellMotion()
}

// DisableMouseCellMotion disables Mouse Cell Motion tracking. This will be
// called automatically when exiting a Bubble Tea program.
//
// Deprecated: The mouse will automatically be disabled when the program exits.
func (p *Program[M]) DisableMouseCellMotion() {
	p.renderer.disableMouseCellMotion()
}

// EnableMouseAllMotion enables mouse click, release, wheel and motion events,
// regardless of whether a mouse button is pressed. Many modern terminals
// support this, but not all.
//
// Deprecated: Use the WithMouseAllMotion ProgramOption instead.
func (p *Program[M]) EnableMouseAllMotion() {
	p.renderer.enableMouseAllMotion()
}

// DisableMouseAllMotion disables All Motion mouse tracking. This will be
// called automatically when exiting a Bubble Tea program.
//
// Deprecated: The mouse will automatically be disabled when the program exits.
func (p *Program[M]) DisableMouseAllMotion() {
	p.renderer.disableMouseAllMotion()
}
