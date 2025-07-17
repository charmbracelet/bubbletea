package tea

import (
	"fmt"

	uv "github.com/charmbracelet/ultraviolet"
)

// MouseButton represents the button that was pressed during a mouse message.
type MouseButton = uv.MouseButton

// Mouse event buttons
//
// This is based on X11 mouse button codes.
//
//	1 = left button
//	2 = middle button (pressing the scroll wheel)
//	3 = right button
//	4 = turn scroll wheel up
//	5 = turn scroll wheel down
//	6 = push scroll wheel left
//	7 = push scroll wheel right
//	8 = 4th button (aka browser backward button)
//	9 = 5th button (aka browser forward button)
//	10
//	11
//
// Other buttons are not supported.
const (
	MouseNone       = uv.MouseNone
	MouseLeft       = uv.MouseLeft
	MouseMiddle     = uv.MouseMiddle
	MouseRight      = uv.MouseRight
	MouseWheelUp    = uv.MouseWheelUp
	MouseWheelDown  = uv.MouseWheelDown
	MouseWheelLeft  = uv.MouseWheelLeft
	MouseWheelRight = uv.MouseWheelRight
	MouseBackward   = uv.MouseBackward
	MouseForward    = uv.MouseForward
	MouseButton10   = uv.MouseButton10
	MouseButton11
)

// MouseMsg represents a mouse message. This is a generic mouse message that
// can represent any kind of mouse event.
type MouseMsg interface {
	fmt.Stringer

	// Mouse returns the underlying mouse event.
	Mouse() Mouse
}

// Mouse represents a Mouse message. Use [MouseMsg] to represent all mouse
// messages.
//
// The X and Y coordinates are zero-based, with (0,0) being the upper left
// corner of the terminal.
//
//	// Catch all mouse events
//	switch msg := msg.(type) {
//	case MouseMsg:
//	    m := msg.Mouse()
//	    fmt.Println("Mouse event:", m.X, m.Y, m)
//	}
//
//	// Only catch mouse click events
//	switch msg := msg.(type) {
//	case MouseClickMsg:
//	    fmt.Println("Mouse click event:", msg.X, msg.Y, msg)
//	}
type Mouse struct {
	X, Y   int
	Button MouseButton
	Mod    KeyMod
}

// String returns a string representation of the mouse message.
func (m Mouse) String() (s string) {
	return uv.Mouse(m).String()
}

// MouseClickMsg represents a mouse button click message.
type MouseClickMsg Mouse

// String returns a string representation of the mouse click message.
func (e MouseClickMsg) String() string {
	return Mouse(e).String()
}

// Mouse returns the underlying mouse event. This is a convenience method and
// syntactic sugar to satisfy the [MouseMsg] interface, and cast the mouse
// event to [Mouse].
func (e MouseClickMsg) Mouse() Mouse {
	return Mouse(e)
}

// MouseReleaseMsg represents a mouse button release message.
type MouseReleaseMsg Mouse

// String returns a string representation of the mouse release message.
func (e MouseReleaseMsg) String() string {
	return Mouse(e).String()
}

// Mouse returns the underlying mouse event. This is a convenience method and
// syntactic sugar to satisfy the [MouseMsg] interface, and cast the mouse
// event to [Mouse].
func (e MouseReleaseMsg) Mouse() Mouse {
	return Mouse(e)
}

// MouseWheelMsg represents a mouse wheel message event.
type MouseWheelMsg Mouse

// String returns a string representation of the mouse wheel message.
func (e MouseWheelMsg) String() string {
	return Mouse(e).String()
}

// Mouse returns the underlying mouse event. This is a convenience method and
// syntactic sugar to satisfy the [MouseMsg] interface, and cast the mouse
// event to [Mouse].
func (e MouseWheelMsg) Mouse() Mouse {
	return Mouse(e)
}

// MouseMotionMsg represents a mouse motion message.
type MouseMotionMsg Mouse

// String returns a string representation of the mouse motion message.
func (e MouseMotionMsg) String() string {
	m := Mouse(e)
	if m.Button != 0 {
		return m.String() + "+motion"
	}
	return m.String() + "motion"
}

// Mouse returns the underlying mouse event. This is a convenience method and
// syntactic sugar to satisfy the [MouseMsg] interface, and cast the mouse
// event to [Mouse].
func (e MouseMotionMsg) Mouse() Mouse {
	return Mouse(e)
}
