package tea

import "github.com/charmbracelet/x/input"

// Mouse msgs contains information about mouse events and are sent to a
// programs update function when mouse activity occurs. Note that the mouse
// must first be enabled in order for the mouse events to be received.
type (
	Mouse           input.Mouse
	MouseClickMsg   input.MouseClickEvent
	MouseReleaseMsg input.MouseReleaseEvent
	MouseWheelMsg   input.MouseWheelEvent
	MouseMotionMsg  input.MouseMotionEvent

	// Deprecated: Use MouseClickMsg, MouseReleaseMsg, MouseWheelMsg, or
	// MouseMotionMsg instead.
	MouseMsg Mouse
)

// String returns a string representation of a mouse event.
//
// Deprecated: Use MouseClickMsg.String, MouseReleaseMsg.String,
// MouseWheelMsg.String, or MouseMotionMsg.String instead.
func (m MouseMsg) String() string {
	return input.Mouse(m).String()
}

// String returns a string representation of a mouse event.
func (m Mouse) String() string {
	return input.Mouse(m).String()
}

// String returns a string representation of a mouse event.
func (m MouseClickMsg) String() string {
	return input.MouseClickEvent(m).String()
}

// String returns a string representation of a mouse event.
func (m MouseReleaseMsg) String() string {
	return input.MouseReleaseEvent(m).String()
}

// String returns a string representation of a mouse event.
func (m MouseWheelMsg) String() string {
	return input.MouseWheelEvent(m).String()
}

// String returns a string representation of a mouse event.
func (m MouseMotionMsg) String() string {
	return input.MouseMotionEvent(m).String()
}
