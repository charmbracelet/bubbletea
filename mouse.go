package tea

import "errors"

type MouseMsg MouseEvent

// MouseEvent represents a mouse event, which could be a click, a scroll wheel
// movement, a cursor movement, or a combination.
type MouseEvent struct {
	X    int
	Y    int
	Type MouseEventType
	Alt  bool
	Ctrl bool
}

// String returns a string representation of a mouse event.
func (m MouseEvent) String() (s string) {
	if m.Ctrl {
		s += "ctrl+"
	}
	if m.Alt {
		s += "alt+"
	}
	s += mouseEventTypes[m.Type]
	return s
}

// MouseEventType indicates the type of mouse event occurring.
type MouseEventType int

const (
	MouseUnknown MouseEventType = iota
	MouseLeft
	MouseRight
	MouseMiddle
	MouseRelease
	MouseWheelUp
	MouseWheelDown
	MouseMotion
)

var mouseEventTypes = map[MouseEventType]string{
	MouseUnknown:   "unknown",
	MouseLeft:      "left",
	MouseRight:     "right",
	MouseMiddle:    "middle",
	MouseRelease:   "release",
	MouseWheelUp:   "wheel up",
	MouseWheelDown: "wheel down",
	MouseMotion:    "motion",
}

// Parse an X10-encoded mouse event; the simplest kind. The last release of
// X10 was December 1986, by the way.
//
// X10 mouse events look like:
//
//     ESC [M Cb Cx Cy
//
func parseX10MouseEvent(buf []byte) (m MouseEvent, err error) {
	if len(buf) != 6 || string(buf[:3]) != "\x1b[M" {
		return m, errors.New("not an X10 mouse event")
	}

	e := buf[3] - 32

	switch e {
	case 35:
		m.Type = MouseMotion
	case 64:
		m.Type = MouseWheelUp
	case 65:
		m.Type = MouseWheelDown
	default:
		switch e & 3 {
		case 0:
			if e&64 != 0 {
				m.Type = MouseWheelUp
			} else {
				m.Type = MouseLeft
			}
		case 1:
			if e&64 != 0 {
				m.Type = MouseWheelDown
			} else {
				m.Type = MouseMiddle
			}
		case 2:
			m.Type = MouseRight
		case 3:
			m.Type = MouseRelease
		}
	}

	if e&8 != 0 {
		m.Alt = true
	}
	if e&16 != 0 {
		m.Ctrl = true
	}

	// (1,1) is the upper left. We subtract 1 to normalize it to (0,0).
	m.X = int(buf[4]) - 32 - 1
	m.Y = int(buf[5]) - 32 - 1

	return m, nil
}
