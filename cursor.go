package tea

import "image"

// CursorPositionMsg is a message that represents the terminal cursor position.
type CursorPositionMsg struct {
	// Row is the row number.
	Row int

	// Column is the column number.
	Column int
}

// CursorStyle is a style that represents the terminal cursor.
type CursorStyle int

// Cursor styles.
const (
	CursorBlock CursorStyle = iota
	CursorUnderline
	CursorBar
)

// setCursorStyle is an internal message that sets the cursor style. This matches the
// ANSI escape sequence values for cursor styles. This includes:
//
//	0: Blinking block
//	1: Blinking block (default)
//	2: Steady block
//	3: Blinking underline
//	4: Steady underline
//	5: Blinking bar (xterm)
//	6: Steady bar (xterm)
type setCursorStyle int

// SetCursorStyle is a command that sets the terminal cursor style. Steady
// determines if the cursor should blink or not.
func SetCursorStyle(style CursorStyle, steady bool) Cmd {
	// We're using the ANSI escape sequence values for cursor styles.
	// We need to map both [style] and [steady] to the correct value.
	style = (style * 2) + 1
	if steady {
		style++
	}
	return func() Msg {
		return setCursorStyle(style)
	}
}

// setCursorPosMsg represents a message to set the cursor position.
type setCursorPosMsg image.Point

// SetCursorPosition sets the cursor position to the specified relative
// coordinates. Using -1 for either x or y will not change the cursor position
// for that axis.
func SetCursorPosition(x, y int) Cmd {
	return func() Msg {
		return setCursorPosMsg{x, y}
	}
}
