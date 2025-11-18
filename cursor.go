package tea

// Position represents a position in the terminal.
type Position struct{ X, Y int }

// CursorPositionMsg is a message that represents the terminal cursor position.
type CursorPositionMsg struct {
	X, Y int
}

// CursorShape represents a terminal cursor shape.
type CursorShape int

// Cursor shapes.
const (
	CursorBlock CursorShape = iota
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
