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
	MouseNone       = input.MouseNone
	MouseLeft       = input.MouseLeft
	MouseMiddle     = input.MouseMiddle
	MouseRight      = input.MouseRight
	MouseWheelUp    = input.MouseWheelUp
	MouseWheelDown  = input.MouseWheelDown
	MouseWheelLeft  = input.MouseWheelLeft
	MouseWheelRight = input.MouseWheelRight
	MouseBackward   = input.MouseBackward
	MouseForward    = input.MouseForward
	MouseExtra1     = input.MouseExtra1
	MouseExtra2     = input.MouseExtra2
)
