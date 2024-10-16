package tea

import (
	"fmt"

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

// String returns a string representation of the mouse button.
func (b MouseButton) String() string {
	return mouseButtons[b]
}

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
	if m.Mod.Contains(ModCtrl) {
		s += "ctrl+"
	}
	if m.Mod.Contains(ModAlt) {
		s += "alt+"
	}
	if m.Mod.Contains(ModShift) {
		s += "shift+"
	}

	str, ok := mouseButtons[m.Button]
	if !ok {
		s += "unknown"
	} else if str != "none" { // motion events don't have a button
		s += str
	}

	return s
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

	m := Mouse{X: x, Y: y, Button: btn, Mod: mod}

	// Wheel buttons don't have release events
	// Motion can be reported as a release event in some terminals (Windows Terminal)
	if isWheel(m.Button) {
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

	m := Mouse{X: x, Y: y, Button: btn, Mod: mod}
	if isWheel(m.Button) {
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
