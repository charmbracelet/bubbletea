package tea

// MouseMsg contains information about a mouse event and are sent to a programs
// update function when mouse activity occurs. Note that the mouse must first
// be enabled in order for the mouse events to be received.
//
// TODO(v2): Add a MouseMsg interface that incorporates all the mouse message
// types.
//
// Deprecated: in favor of MouseClickMsg, MouseReleaseMsg, MouseWheelMsg, and
// MouseMotionMsg.
type MouseMsg struct {
	X      int
	Y      int
	Shift  bool
	Alt    bool
	Ctrl   bool
	Action MouseAction
	Button MouseButton
	Type   MouseEventType
}

// MouseEvent represents a mouse event.
//
// Deprecated: Use Mouse.
type MouseEvent = MouseMsg

// IsWheel returns true if the mouse event is a wheel event.
func (m MouseMsg) IsWheel() bool {
	return m.Button == MouseButtonWheelUp || m.Button == MouseButtonWheelDown ||
		m.Button == MouseButtonWheelLeft || m.Button == MouseButtonWheelRight
}

// String returns a string representation of a mouse event.
func (m MouseMsg) String() (s string) {
	if m.Ctrl {
		s += "ctrl+"
	}
	if m.Alt {
		s += "alt+"
	}
	if m.Shift {
		s += "shift+"
	}

	if m.Button == MouseButtonNone { //nolint:nestif
		if m.Action == MouseActionMotion || m.Action == MouseActionRelease {
			s += mouseMsgActions[m.Action]
		} else {
			s += "unknown"
		}
	} else if m.IsWheel() {
		s += mouseMsgButtons[m.Button]
	} else {
		btn := mouseMsgButtons[m.Button]
		if btn != "" {
			s += btn
		}
		act := mouseMsgActions[m.Action]
		if act != "" {
			s += " " + act
		}
	}

	return s
}

// MouseAction represents the action that occurred during a mouse event.
//
// Deprecated: Use MouseClickMsg, MouseReleaseMsg, MouseWheelMsg, and
// MouseMotionMsg.
type MouseAction int

// Mouse event actions.
//
// Deprecated in favor of MouseClickMsg, MouseReleaseMsg, MouseWheelMsg, and
// MouseMotionMsg.
const (
	MouseActionPress MouseAction = iota
	MouseActionRelease
	MouseActionMotion
)

var mouseMsgActions = map[MouseAction]string{
	MouseActionPress:   "press",
	MouseActionRelease: "release",
	MouseActionMotion:  "motion",
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
//
// Deprecated: Use MouseNone, MouseLeft, etc.
const (
	MouseButtonNone       = MouseNone
	MouseButtonLeft       = MouseLeft
	MouseButtonMiddle     = MouseMiddle
	MouseButtonRight      = MouseRight
	MouseButtonWheelUp    = MouseWheelUp
	MouseButtonWheelDown  = MouseWheelDown
	MouseButtonWheelLeft  = MouseWheelLeft
	MouseButtonWheelRight = MouseWheelRight
	MouseButtonBackward   = MouseBackward
	MouseButtonForward    = MouseForward
	MouseButton10         = MouseExtra1
	MouseButton11         = MouseExtra2
)

// Deprecated: Use mouseButtons.
var mouseMsgButtons = map[MouseButton]string{
	MouseButtonNone:       "none",
	MouseButtonLeft:       "left",
	MouseButtonMiddle:     "middle",
	MouseButtonRight:      "right",
	MouseButtonWheelUp:    "wheel up",
	MouseButtonWheelDown:  "wheel down",
	MouseButtonWheelLeft:  "wheel left",
	MouseButtonWheelRight: "wheel right",
	MouseButtonBackward:   "backward",
	MouseButtonForward:    "forward",
	MouseButton10:         "button 10",
	MouseButton11:         "button 11",
}

// MouseEventType indicates the type of mouse event occurring.
//
// Deprecated: Use MouseButton.
type MouseEventType = MouseButton

// Mouse event types.
//
// Deprecated in favor of MouseReleaseMsg and MouseMotionMsg.
const (
	MouseUnknown = MouseNone

	MouseRelease MouseEventType = -iota // mouse button release (X10 only)
	MouseMotion
)

// toMouseMsg converts a mouse event to a mouse message.
func toMouseMsg(m Mouse) MouseMsg {
	return MouseMsg{
		X:      m.X,
		Y:      m.Y,
		Shift:  m.Mod.Contains(ModShift),
		Alt:    m.Mod.Contains(ModAlt),
		Ctrl:   m.Mod.Contains(ModCtrl),
		Button: m.Button,
	}
}
