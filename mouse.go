package tea

import "errors"

type MouseMsg MouseEvent

type MouseEvent struct {
	X      int
	Y      int
	Button MouseButton
	Alt    bool
	Ctrl   bool
}

func (m MouseEvent) String() (s string) {
	if m.Ctrl {
		s += "ctrl+"
	}
	if m.Alt {
		s += "alt+"
	}
	s += mouseButtonNames[m.Button]
	return s
}

type MouseButton int

const (
	MouseLeft MouseButton = iota
	MouseRight
	MouseMiddle
	MouseRelease
	MouseWheelUp
	MouseWheelDown
	MouseMotion
)

var mouseButtonNames = map[MouseButton]string{
	MouseLeft:      "left",
	MouseRight:     "right",
	MouseMiddle:    "middle",
	MouseRelease:   "release",
	MouseWheelUp:   "wheel up",
	MouseWheelDown: "wheel down",
	MouseMotion:    "motion",
}

// Parse an X10-encoded mouse event. The simplest kind. The last release of
// X10 was December 1986, by the way.
func parseX10MouseEvent(buf []byte) (m MouseEvent, err error) {
	if len(buf) != 6 || string(buf[:3]) != "\x1b[M" {
		return m, errors.New("not an X10 mouse event")
	}

	e := buf[3] - 32

	switch e {
	case 35:
		m.Button = MouseMotion
	case 64:
		m.Button = MouseWheelUp
	case 65:
		m.Button = MouseWheelDown
	default:
		switch e & 3 {
		case 0:
			m.Button = MouseLeft
		case 1:
			m.Button = MouseMiddle
		case 2:
			m.Button = MouseRight
		case 3:
			m.Button = MouseRelease
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
