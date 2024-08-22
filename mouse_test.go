package tea

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

func TestMouseEvent_String(t *testing.T) {
	tt := []struct {
		name     string
		event    Msg
		expected string
	}{
		{
			name:     "unknown",
			event:    MouseClickMsg{button: MouseButton(0xff)},
			expected: "unknown",
		},
		{
			name:     "left",
			event:    MouseClickMsg{button: MouseLeft},
			expected: "left",
		},
		{
			name:     "right",
			event:    MouseClickMsg{button: MouseRight},
			expected: "right",
		},
		{
			name:     "middle",
			event:    MouseClickMsg{button: MouseMiddle},
			expected: "middle",
		},
		{
			name:     "release",
			event:    MouseReleaseMsg{button: MouseNone},
			expected: "",
		},
		{
			name:     "wheelup",
			event:    MouseWheelMsg{button: MouseWheelUp},
			expected: "wheelup",
		},
		{
			name:     "wheeldown",
			event:    MouseWheelMsg{button: MouseWheelDown},
			expected: "wheeldown",
		},
		{
			name:     "wheelleft",
			event:    MouseWheelMsg{button: MouseWheelLeft},
			expected: "wheelleft",
		},
		{
			name:     "wheelright",
			event:    MouseWheelMsg{button: MouseWheelRight},
			expected: "wheelright",
		},
		{
			name:     "motion",
			event:    MouseMotionMsg{button: MouseNone},
			expected: "motion",
		},
		{
			name:     "shift+left",
			event:    MouseReleaseMsg{button: MouseLeft, mod: ModShift},
			expected: "shift+left",
		},
		{
			name: "shift+left", event: MouseClickMsg{button: MouseLeft, mod: ModShift},
			expected: "shift+left",
		},
		{
			name:     "ctrl+shift+left",
			event:    MouseClickMsg{button: MouseLeft, mod: ModCtrl | ModShift},
			expected: "ctrl+shift+left",
		},
		{
			name:     "alt+left",
			event:    MouseClickMsg{button: MouseLeft, mod: ModAlt},
			expected: "alt+left",
		},
		{
			name:     "ctrl+left",
			event:    MouseClickMsg{button: MouseLeft, mod: ModCtrl},
			expected: "ctrl+left",
		},
		{
			name:     "ctrl+alt+left",
			event:    MouseClickMsg{button: MouseLeft, mod: ModAlt | ModCtrl},
			expected: "ctrl+alt+left",
		},
		{
			name:     "ctrl+alt+shift+left",
			event:    MouseClickMsg{button: MouseLeft, mod: ModAlt | ModCtrl | ModShift},
			expected: "ctrl+alt+shift+left",
		},
		{
			name:     "ignore coordinates",
			event:    MouseClickMsg{x: 100, y: 200, button: MouseLeft},
			expected: "left",
		},
		{
			name:     "broken type",
			event:    MouseClickMsg{button: MouseButton(120)},
			expected: "unknown",
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := fmt.Sprint(tc.event)

			if tc.expected != actual {
				t.Fatalf("expected %q but got %q",
					tc.expected,
					actual,
				)
			}
		})
	}
}

func TestParseX10MouseDownEvent(t *testing.T) {
	encode := func(b byte, x, y int) []byte {
		return []byte{
			'\x1b',
			'[',
			'M',
			byte(32) + b,
			byte(x + 32 + 1),
			byte(y + 32 + 1),
		}
	}

	tt := []struct {
		name     string
		buf      []byte
		expected Msg
	}{
		// Position.
		{
			name:     "zero position",
			buf:      encode(0b0000_0000, 0, 0),
			expected: MouseClickMsg{x: 0, y: 0, button: MouseLeft},
		},
		{
			name:     "max position",
			buf:      encode(0b0000_0000, 222, 222), // Because 255 (max int8) - 32 - 1.
			expected: MouseClickMsg{x: 222, y: 222, button: MouseLeft},
		},
		// Simple.
		{
			name:     "left",
			buf:      encode(0b0000_0000, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseLeft},
		},
		{
			name:     "left in motion",
			buf:      encode(0b0010_0000, 32, 16),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseLeft},
		},
		{
			name:     "middle",
			buf:      encode(0b0000_0001, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseMiddle},
		},
		{
			name:     "middle in motion",
			buf:      encode(0b0010_0001, 32, 16),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseMiddle},
		},
		{
			name:     "right",
			buf:      encode(0b0000_0010, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseRight},
		},
		{
			name:     "right in motion",
			buf:      encode(0b0010_0010, 32, 16),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseRight},
		},
		{
			name:     "motion",
			buf:      encode(0b0010_0011, 32, 16),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseNone},
		},
		{
			name:     "wheel up",
			buf:      encode(0b0100_0000, 32, 16),
			expected: MouseWheelMsg{x: 32, y: 16, button: MouseWheelUp},
		},
		{
			name:     "wheel down",
			buf:      encode(0b0100_0001, 32, 16),
			expected: MouseWheelMsg{x: 32, y: 16, button: MouseWheelDown},
		},
		{
			name:     "wheel left",
			buf:      encode(0b0100_0010, 32, 16),
			expected: MouseWheelMsg{x: 32, y: 16, button: MouseWheelLeft},
		},
		{
			name:     "wheel right",
			buf:      encode(0b0100_0011, 32, 16),
			expected: MouseWheelMsg{x: 32, y: 16, button: MouseWheelRight},
		},
		{
			name:     "release",
			buf:      encode(0b0000_0011, 32, 16),
			expected: MouseReleaseMsg{x: 32, y: 16, button: MouseNone},
		},
		{
			name:     "backward",
			buf:      encode(0b1000_0000, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseBackward},
		},
		{
			name:     "forward",
			buf:      encode(0b1000_0001, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseForward},
		},
		{
			name:     "button 10",
			buf:      encode(0b1000_0010, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseExtra1},
		},
		{
			name:     "button 11",
			buf:      encode(0b1000_0011, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseExtra2},
		},
		// Combinations.
		{
			name:     "alt+right",
			buf:      encode(0b0000_1010, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, mod: ModAlt, button: MouseRight},
		},
		{
			name:     "ctrl+right",
			buf:      encode(0b0001_0010, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, mod: ModCtrl, button: MouseRight},
		},
		{
			name:     "left in motion",
			buf:      encode(0b0010_0000, 32, 16),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseLeft},
		},
		{
			name:     "alt+right in motion",
			buf:      encode(0b0010_1010, 32, 16),
			expected: MouseMotionMsg{x: 32, y: 16, mod: ModAlt, button: MouseRight},
		},
		{
			name:     "ctrl+right in motion",
			buf:      encode(0b0011_0010, 32, 16),
			expected: MouseMotionMsg{x: 32, y: 16, mod: ModCtrl, button: MouseRight},
		},
		{
			name:     "ctrl+alt+right",
			buf:      encode(0b0001_1010, 32, 16),
			expected: MouseClickMsg{x: 32, y: 16, mod: ModAlt | ModCtrl, button: MouseRight},
		},
		{
			name:     "ctrl+wheel up",
			buf:      encode(0b0101_0000, 32, 16),
			expected: MouseWheelMsg{x: 32, y: 16, mod: ModCtrl, button: MouseWheelUp},
		},
		{
			name:     "alt+wheel down",
			buf:      encode(0b0100_1001, 32, 16),
			expected: MouseWheelMsg{x: 32, y: 16, mod: ModAlt, button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+wheel down",
			buf:      encode(0b0101_1001, 32, 16),
			expected: MouseWheelMsg{x: 32, y: 16, mod: ModAlt | ModCtrl, button: MouseWheelDown},
		},
		// Overflow position.
		{
			name:     "overflow position",
			buf:      encode(0b0010_0000, 250, 223), // Because 255 (max int8) - 32 - 1.
			expected: MouseMotionMsg{x: -6, y: -33, button: MouseLeft},
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := parseX10MouseEvent(tc.buf)

			if tc.expected != actual {
				t.Fatalf("expected %#v but got %#v",
					tc.expected,
					actual,
				)
			}
		})
	}
}

func TestParseSGRMouseEvent(t *testing.T) {
	encode := func(b, x, y int, r bool) *ansi.CsiSequence {
		re := 'M'
		if r {
			re = 'm'
		}
		return &ansi.CsiSequence{
			Params: []int{b, x + 1, y + 1},
			Cmd:    int(re) | ('<' << parser.MarkerShift),
		}
	}

	tt := []struct {
		name     string
		buf      *ansi.CsiSequence
		expected Msg
	}{
		// Position.
		{
			name:     "zero position",
			buf:      encode(0, 0, 0, false),
			expected: MouseClickMsg{x: 0, y: 0, button: MouseLeft},
		},
		{
			name:     "225 position",
			buf:      encode(0, 225, 225, false),
			expected: MouseClickMsg{x: 225, y: 225, button: MouseLeft},
		},
		// Simple.
		{
			name:     "left",
			buf:      encode(0, 32, 16, false),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseLeft},
		},
		{
			name:     "left in motion",
			buf:      encode(32, 32, 16, false),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseLeft},
		},
		{
			name:     "left",
			buf:      encode(0, 32, 16, true),
			expected: MouseReleaseMsg{x: 32, y: 16, button: MouseLeft},
		},
		{
			name:     "middle",
			buf:      encode(1, 32, 16, false),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseMiddle},
		},
		{
			name:     "middle in motion",
			buf:      encode(33, 32, 16, false),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseMiddle},
		},
		{
			name:     "middle",
			buf:      encode(1, 32, 16, true),
			expected: MouseReleaseMsg{x: 32, y: 16, button: MouseMiddle},
		},
		{
			name:     "right",
			buf:      encode(2, 32, 16, false),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseRight},
		},
		{
			name:     "right",
			buf:      encode(2, 32, 16, true),
			expected: MouseReleaseMsg{x: 32, y: 16, button: MouseRight},
		},
		{
			name:     "motion",
			buf:      encode(35, 32, 16, false),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseNone},
		},
		{
			name:     "wheel up",
			buf:      encode(64, 32, 16, false),
			expected: MouseWheelMsg{x: 32, y: 16, button: MouseWheelUp},
		},
		{
			name:     "wheel down",
			buf:      encode(65, 32, 16, false),
			expected: MouseWheelMsg{x: 32, y: 16, button: MouseWheelDown},
		},
		{
			name:     "wheel left",
			buf:      encode(66, 32, 16, false),
			expected: MouseWheelMsg{x: 32, y: 16, button: MouseWheelLeft},
		},
		{
			name:     "wheel right",
			buf:      encode(67, 32, 16, false),
			expected: MouseWheelMsg{x: 32, y: 16, button: MouseWheelRight},
		},
		{
			name:     "backward",
			buf:      encode(128, 32, 16, false),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseBackward},
		},
		{
			name:     "backward in motion",
			buf:      encode(160, 32, 16, false),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseBackward},
		},
		{
			name:     "forward",
			buf:      encode(129, 32, 16, false),
			expected: MouseClickMsg{x: 32, y: 16, button: MouseForward},
		},
		{
			name:     "forward in motion",
			buf:      encode(161, 32, 16, false),
			expected: MouseMotionMsg{x: 32, y: 16, button: MouseForward},
		},
		// Combinations.
		{
			name:     "alt+right",
			buf:      encode(10, 32, 16, false),
			expected: MouseClickMsg{x: 32, y: 16, mod: ModAlt, button: MouseRight},
		},
		{
			name:     "ctrl+right",
			buf:      encode(18, 32, 16, false),
			expected: MouseClickMsg{x: 32, y: 16, mod: ModCtrl, button: MouseRight},
		},
		{
			name:     "ctrl+alt+right",
			buf:      encode(26, 32, 16, false),
			expected: MouseClickMsg{x: 32, y: 16, mod: ModAlt | ModCtrl, button: MouseRight},
		},
		{
			name:     "alt+wheel",
			buf:      encode(73, 32, 16, false),
			expected: MouseWheelMsg{x: 32, y: 16, mod: ModAlt, button: MouseWheelDown},
		},
		{
			name:     "ctrl+wheel",
			buf:      encode(81, 32, 16, false),
			expected: MouseWheelMsg{x: 32, y: 16, mod: ModCtrl, button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+wheel",
			buf:      encode(89, 32, 16, false),
			expected: MouseWheelMsg{x: 32, y: 16, mod: ModAlt | ModCtrl, button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+shift+wheel",
			buf:      encode(93, 32, 16, false),
			expected: MouseWheelMsg{x: 32, y: 16, mod: ModAlt | ModShift | ModCtrl, button: MouseWheelDown},
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := parseSGRMouseEvent(tc.buf)
			if tc.expected != actual {
				t.Fatalf("expected %#v but got %#v",
					tc.expected,
					actual,
				)
			}
		})
	}
}
