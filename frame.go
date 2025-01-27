package tea

import "image/color"

// Cursor represents a cursor on the terminal screen.
type Cursor struct {
	// Position is a [Position] that determines the cursor's position on the
	// screen relative to the top left corner of the frame.
	Position Position

	// Color is a [color.Color] that determines the cursor's color.
	Color color.Color

	// Shape is a [CursorShape] that determines the cursor's shape.
	Shape CursorShape

	// Blink is a boolean that determines whether the cursor should blink.
	Blink bool
}

// NewCursor returns a new cursor with the default settings and the given
// position.
func NewCursor(x, y int) *Cursor {
	return &Cursor{
		Position: Position{X: x, Y: y},
		Color:    nil,
		Shape:    CursorBlock,
		Blink:    true,
	}
}

// Frame represents a single frame of the program's output.
type Frame struct {
	// Content contains the frame's content. This is the only required field.
	// It should be a string of text and ANSI escape codes.
	Content string

	// Cursor contains cursor settings for the frame. If nil, the cursor will
	// be hidden.
	Cursor *Cursor
}

// NewFrame creates a new frame with the given content.
func NewFrame(content string) Frame {
	return Frame{Content: content}
}

// String implements the fmt.Stringer interface for [Frame].
func (f Frame) String() string {
	return f.Content
}

// View represents a simple [fmt.Stringer] interface for a view.
type View string

// String implements the fmt.Stringer interface for [View].
func (v View) String() string {
	return string(v)
}
