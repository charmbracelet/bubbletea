package tea

import "image"

// CursorPositionMsg is a message that represents the terminal cursor position.
type CursorPositionMsg image.Point

// CursorStyle is a style that represents the terminal cursor.
type CursorStyle int

// Cursor styles.
const (
	CursorBlock CursorStyle = iota
	CursorUnderline
	CursorBar
)

// requestCursorPosMsg is a message that requests the cursor position.
type requestCursorPosMsg struct{}

// RequestCursorPosition is a command that requests the cursor position.
// The cursor position will be sent as a [CursorPositionMsg] message.
func RequestCursorPosition() Msg {
	return requestCursorPosMsg{}
}
