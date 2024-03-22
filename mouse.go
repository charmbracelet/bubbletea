package tea

import (
	"github.com/charmbracelet/x/input"
)

// MouseMsg contains information about a mouse event and are sent to a programs
// update function when mouse activity occurs. Note that the mouse must first
// be enabled in order for the mouse events to be received.
type (
	MouseEvent     = input.Mouse
	MouseMsg       = input.MouseDownEvent
	MouseDownMsg   = input.MouseDownEvent
	MouseUpMsg     = input.MouseUpEvent
	MouseWheelMsg  = input.MouseWheelEvent
	MouseMotionMsg = input.MouseMotionEvent
)

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
