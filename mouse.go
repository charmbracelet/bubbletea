package tea

import "strconv"

// MouseMsg contains information about a mouse event and are sent to a programs
// update function when mouse activity occurs. Note that the mouse must first
// be enabled in order for the mouse events to be received.
type MouseMsg MouseEvent

// String returns a string representation of a mouse event.
func (m MouseMsg) String() string {
	return MouseEvent(m).String()
}

// MouseEvent represents a mouse event, which could be a click, a scroll wheel
// movement, a cursor movement, or a combination.
type MouseEvent struct {
	X      int
	Y      int
	Shift  bool
	Alt    bool
	Ctrl   bool
	Action MouseAction
	Button MouseButton

	// Deprecated: Use MouseAction & MouseButton instead.
	Type MouseEventType
}

// IsWheel returns true if the mouse event is a wheel event.
func (m MouseEvent) IsWheel() bool {
	return m.Button == MouseButtonWheelUp || m.Button == MouseButtonWheelDown ||
		m.Button == MouseButtonWheelLeft || m.Button == MouseButtonWheelRight
}

// String returns a string representation of a mouse event.
func (m MouseEvent) String() (s string) {
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
			s += mouseActions[m.Action]
		} else {
			s += "unknown"
		}
	} else if m.IsWheel() {
		s += mouseButtons[m.Button]
	} else {
		btn := mouseButtons[m.Button]
		if btn != "" {
			s += btn
		}
		act := mouseActions[m.Action]
		if act != "" {
			s += " " + act
		}
	}

	return s
}

// MouseAction represents the action that occurred during a mouse event.
type MouseAction int

// Mouse event actions.
const (
	MouseActionPress MouseAction = iota
	MouseActionRelease
	MouseActionMotion
)

var mouseActions = map[MouseAction]string{
	MouseActionPress:   "press",
	MouseActionRelease: "release",
	MouseActionMotion:  "motion",
}

// MouseButton represents the button that was pressed during a mouse event.
type MouseButton int

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
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
	MouseButtonWheelUp
	MouseButtonWheelDown
	MouseButtonWheelLeft
	MouseButtonWheelRight
	MouseButtonBackward
	MouseButtonForward
	MouseButton10
	MouseButton11
)

var mouseButtons = map[MouseButton]string{
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
// Deprecated: Use MouseAction & MouseButton instead.
type MouseEventType int

// Mouse event types.
//
// Deprecated: Use MouseAction & MouseButton instead.
const (
	MouseUnknown MouseEventType = iota
	MouseLeft
	MouseRight
	MouseMiddle
	MouseRelease // mouse button release (X10 only)
	MouseWheelUp
	MouseWheelDown
	MouseWheelLeft
	MouseWheelRight
	MouseBackward
	MouseForward
	MouseMotion
)

// Parse SGR-encoded mouse events; SGR extended mouse events. SGR mouse events
// look like:
//
//	ESC [ < Cb ; Cx ; Cy (M or m)
//
// where:
//
//	Cb is the encoded button code
//	Cx is the x-coordinate of the mouse
//	Cy is the y-coordinate of the mouse
//	M is for button press, m is for button release
//
// https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Extended-coordinates
func parseSGRMouseEvent(buf []byte) MouseEvent {
	str := string(buf[3:])
	matches := mouseSGRRegex.FindStringSubmatch(str)
	if len(matches) != 5 { //nolint:gomnd
		// Unreachable, we already checked the regex in `detectOneMsg`.
		panic("invalid mouse event")
	}

	b, _ := strconv.Atoi(matches[1])
	px := matches[2]
	py := matches[3]
	release := matches[4] == "m"
	m := parseMouseButton(b, true)

	// Wheel buttons don't have release events
	// Motion can be reported as a release event in some terminals (Windows Terminal)
	if m.Action != MouseActionMotion && !m.IsWheel() && release {
		m.Action = MouseActionRelease
		m.Type = MouseRelease
	}

	x, _ := strconv.Atoi(px)
	y, _ := strconv.Atoi(py)

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	m.X = x - 1
	m.Y = y - 1

	return m
}

const x10MouseByteOffset = 32

// Parse X10-encoded mouse events; the simplest kind. The last release of X10
// was December 1986, by the way. The original X10 mouse protocol limits the Cx
// and Cy coordinates to 223 (=255-032).
//
// X10 mouse events look like:
//
//	ESC [M Cb Cx Cy
//
// See: http://www.xfree86.org/current/ctlseqs.html#Mouse%20Tracking
func parseX10MouseEvent(buf []byte) MouseEvent {
	v := buf[3:6]
	m := parseMouseButton(int(v[0]), false)

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	m.X = int(v[1]) - x10MouseByteOffset - 1
	m.Y = int(v[2]) - x10MouseByteOffset - 1

	return m
}

// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Extended-coordinates
func parseMouseButton(b int, isSGR bool) MouseEvent {
	var m MouseEvent
	e := b
	if !isSGR {
		e -= x10MouseByteOffset
	}

	const (
		bitShift  = 0b0000_0100
		bitAlt    = 0b0000_1000
		bitCtrl   = 0b0001_0000
		bitMotion = 0b0010_0000
		bitWheel  = 0b0100_0000
		bitAdd    = 0b1000_0000 // additional buttons 8-11

		bitsMask = 0b0000_0011
	)

	if e&bitAdd != 0 {
		m.Button = MouseButtonBackward + MouseButton(e&bitsMask)
	} else if e&bitWheel != 0 {
		m.Button = MouseButtonWheelUp + MouseButton(e&bitsMask)
	} else {
		m.Button = MouseButtonLeft + MouseButton(e&bitsMask)
		// X10 reports a button release as 0b0000_0011 (3)
		if e&bitsMask == bitsMask {
			m.Action = MouseActionRelease
			m.Button = MouseButtonNone
		}
	}

	// Motion bit doesn't get reported for wheel events.
	if e&bitMotion != 0 && !m.IsWheel() {
		m.Action = MouseActionMotion
	}

	// Modifiers
	m.Alt = e&bitAlt != 0
	m.Ctrl = e&bitCtrl != 0
	m.Shift = e&bitShift != 0

	// backward compatibility
	switch {
	case m.Button == MouseButtonLeft && m.Action == MouseActionPress:
		m.Type = MouseLeft
	case m.Button == MouseButtonMiddle && m.Action == MouseActionPress:
		m.Type = MouseMiddle
	case m.Button == MouseButtonRight && m.Action == MouseActionPress:
		m.Type = MouseRight
	case m.Button == MouseButtonNone && m.Action == MouseActionRelease:
		m.Type = MouseRelease
	case m.Button == MouseButtonWheelUp && m.Action == MouseActionPress:
		m.Type = MouseWheelUp
	case m.Button == MouseButtonWheelDown && m.Action == MouseActionPress:
		m.Type = MouseWheelDown
	case m.Button == MouseButtonWheelLeft && m.Action == MouseActionPress:
		m.Type = MouseWheelLeft
	case m.Button == MouseButtonWheelRight && m.Action == MouseActionPress:
		m.Type = MouseWheelRight
	case m.Button == MouseButtonBackward && m.Action == MouseActionPress:
		m.Type = MouseBackward
	case m.Button == MouseButtonForward && m.Action == MouseActionPress:
		m.Type = MouseForward
	case m.Action == MouseActionMotion:
		m.Type = MouseMotion
		switch m.Button { //nolint:exhaustive
		case MouseButtonLeft:
			m.Type = MouseLeft
		case MouseButtonMiddle:
			m.Type = MouseMiddle
		case MouseButtonRight:
			m.Type = MouseRight
		case MouseButtonBackward:
			m.Type = MouseBackward
		case MouseButtonForward:
			m.Type = MouseForward
		}
	default:
		m.Type = MouseUnknown
	}

	return m
}
