package tea

import (
	"github.com/charmbracelet/x/ansi"
)

// MouseButton represents the button that was pressed during a mouse message.
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
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
	MouseWheelLeft
	MouseWheelRight
	MouseBackward
	MouseForward
	MouseExtra1
	MouseExtra2
)

var mouseButtons = map[MouseButton]string{
	MouseNone:       "none",
	MouseLeft:       "left",
	MouseMiddle:     "middle",
	MouseRight:      "right",
	MouseWheelUp:    "wheelup",
	MouseWheelDown:  "wheeldown",
	MouseWheelLeft:  "wheelleft",
	MouseWheelRight: "wheelright",
	MouseBackward:   "backward",
	MouseForward:    "forward",
	MouseExtra1:     "button10",
	MouseExtra2:     "button11",
}

// mouse represents a mouse message.
type mouse struct {
	x, y   int
	button MouseButton
	mod    KeyMod
}

var _ MouseMsg = mouse{}

// Button implements MouseMsg.
func (m mouse) Button() MouseButton {
	return m.button
}

// Mod implements MouseMsg.
func (m mouse) Mod() KeyMod {
	return m.mod
}

// X implements MouseMsg.
func (m mouse) X() int {
	return m.x
}

// Y implements MouseMsg.
func (m mouse) Y() int {
	return m.y
}

// String returns a string representation of the mouse message.
func (m mouse) String() (s string) {
	if m.mod.Contains(ModCtrl) {
		s += "ctrl+"
	}
	if m.mod.Contains(ModAlt) {
		s += "alt+"
	}
	if m.mod.Contains(ModShift) {
		s += "shift+"
	}

	str, ok := mouseButtons[m.button]
	if !ok {
		s += "unknown"
	} else if str != "none" { // motion events don't have a button
		s += str
	}

	return s
}

// MouseMsg contains information about a mouse event and are sent to a programs
// update function when mouse activity occurs. Note that the mouse must first
// be enabled in order for the mouse events to be received.
type MouseMsg interface {
	// String returns a string representation of the mouse event.
	String() string

	// X returns the x-coordinate of the mouse event.
	X() int

	// Y returns the y-coordinate of the mouse event.
	Y() int

	// Button returns the button that was pressed during the mouse event.
	Button() MouseButton

	// Mod returns any modifier keys that were pressed during the mouse event.
	Mod() KeyMod
}

// MouseClickMsg represents a mouse button click message.
type MouseClickMsg mouse

var _ MouseMsg = MouseClickMsg{}

// Button implements MouseMsg.
func (e MouseClickMsg) Button() MouseButton {
	return mouse(e).Button()
}

// Mod implements MouseMsg.
func (e MouseClickMsg) Mod() KeyMod {
	return mouse(e).Mod()
}

// X implements MouseMsg.
func (e MouseClickMsg) X() int {
	return mouse(e).X()
}

// Y implements MouseMsg.
func (e MouseClickMsg) Y() int {
	return mouse(e).Y()
}

// String returns a string representation of the mouse click message.
func (e MouseClickMsg) String() string {
	return mouse(e).String()
}

var _ MouseMsg = MouseReleaseMsg{}

// MouseReleaseMsg represents a mouse button release message.
type MouseReleaseMsg mouse

// Button implements MouseMsg.
func (e MouseReleaseMsg) Button() MouseButton {
	return mouse(e).Button()
}

// Mod implements MouseMsg.
func (e MouseReleaseMsg) Mod() KeyMod {
	return mouse(e).Mod()
}

// X implements MouseMsg.
func (e MouseReleaseMsg) X() int {
	return mouse(e).X()
}

// Y implements MouseMsg.
func (e MouseReleaseMsg) Y() int {
	return mouse(e).Y()
}

// String returns a string representation of the mouse release message.
func (e MouseReleaseMsg) String() string {
	return mouse(e).String()
}

var _ MouseMsg = MouseWheelMsg{}

// MouseWheelMsg represents a mouse wheel message event.
type MouseWheelMsg mouse

// Button implements MouseMsg.
func (e MouseWheelMsg) Button() MouseButton {
	return mouse(e).Button()
}

// Mod implements MouseMsg.
func (e MouseWheelMsg) Mod() KeyMod {
	return mouse(e).Mod()
}

// X implements MouseMsg.
func (e MouseWheelMsg) X() int {
	return mouse(e).X()
}

// Y implements MouseMsg.
func (e MouseWheelMsg) Y() int {
	return mouse(e).Y()
}

// String returns a string representation of the mouse wheel message.
func (e MouseWheelMsg) String() string {
	return mouse(e).String()
}

// MouseMotionMsg represents a mouse motion message.
type MouseMotionMsg mouse

var _ MouseMsg = MouseMotionMsg{}

// Button implements MouseMsg.
func (e MouseMotionMsg) Button() MouseButton {
	return mouse(e).Button()
}

// Mod implements MouseMsg.
func (e MouseMotionMsg) Mod() KeyMod {
	return mouse(e).Mod()
}

// X implements MouseMsg.
func (e MouseMotionMsg) X() int {
	return mouse(e).X()
}

// Y implements MouseMsg.
func (e MouseMotionMsg) Y() int {
	return mouse(e).Y()
}

// String returns a string representation of the mouse motion message.
func (e MouseMotionMsg) String() string {
	m := mouse(e)
	if m.button != 0 {
		return m.String() + "+motion"
	}
	return m.String() + "motion"
}

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
func parseSGRMouseEvent(csi *ansi.CsiSequence) Msg {
	x := csi.Param(1)
	y := csi.Param(2)
	release := csi.Command() == 'm'
	mod, btn, _, isMotion := parseMouseButton(csi.Param(0))

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	x--
	y--

	m := mouse{x: x, y: y, button: btn, mod: mod}

	// Wheel buttons don't have release events
	// Motion can be reported as a release event in some terminals (Windows Terminal)
	if isWheel(m.button) {
		return MouseWheelMsg(m)
	} else if !isMotion && release {
		return MouseReleaseMsg(m)
	} else if isMotion {
		return MouseMotionMsg(m)
	}
	return MouseClickMsg(m)
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
func parseX10MouseEvent(buf []byte) Msg {
	v := buf[3:6]
	b := int(v[0])
	if b >= x10MouseByteOffset {
		// XXX: b < 32 should be impossible, but we're being defensive.
		b -= x10MouseByteOffset
	}

	mod, btn, isRelease, isMotion := parseMouseButton(b)

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	x := int(v[1]) - x10MouseByteOffset - 1
	y := int(v[2]) - x10MouseByteOffset - 1

	m := mouse{x: x, y: y, button: btn, mod: mod}
	if isWheel(m.button) {
		return MouseWheelMsg(m)
	} else if isMotion {
		return MouseMotionMsg(m)
	} else if isRelease {
		return MouseReleaseMsg(m)
	}
	return MouseClickMsg(m)
}

// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Extended-coordinates
func parseMouseButton(b int) (mod KeyMod, btn MouseButton, isRelease bool, isMotion bool) {
	// mouse bit shifts
	const (
		bitShift  = 0b0000_0100
		bitAlt    = 0b0000_1000
		bitCtrl   = 0b0001_0000
		bitMotion = 0b0010_0000
		bitWheel  = 0b0100_0000
		bitAdd    = 0b1000_0000 // additional buttons 8-11

		bitsMask = 0b0000_0011
	)

	// Modifiers
	if b&bitAlt != 0 {
		mod |= ModAlt
	}
	if b&bitCtrl != 0 {
		mod |= ModCtrl
	}
	if b&bitShift != 0 {
		mod |= ModShift
	}

	if b&bitAdd != 0 {
		btn = MouseBackward + MouseButton(b&bitsMask)
	} else if b&bitWheel != 0 {
		btn = MouseWheelUp + MouseButton(b&bitsMask)
	} else {
		btn = MouseLeft + MouseButton(b&bitsMask)
		// X10 reports a button release as 0b0000_0011 (3)
		if b&bitsMask == bitsMask {
			btn = MouseNone
			isRelease = true
		}
	}

	// Motion bit doesn't get reported for wheel events.
	if b&bitMotion != 0 && !isWheel(btn) {
		isMotion = true
	}

	return
}

// isWheel returns true if the mouse event is a wheel event.
func isWheel(btn MouseButton) bool {
	return btn >= MouseWheelUp && btn <= MouseWheelRight
}
