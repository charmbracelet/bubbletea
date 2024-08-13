package tea

// CursorPositionMsg is a message that represents the terminal cursor position.
type CursorPositionMsg struct {
	// Row is the row number.
	Row int

	// Column is the column number.
	Column int
}
