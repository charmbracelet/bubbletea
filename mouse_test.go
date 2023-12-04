package tea

import (
	"fmt"
	"testing"
)

func TestMouseEvent_String(t *testing.T) {
	tt := []struct {
		name     string
		event    MouseEvent
		expected string
	}{
		{
			name: "unknown",
			event: MouseEvent{
				Action: MouseActionPress,
				Button: MouseButtonNone,
				Type:   MouseUnknown,
			},
			expected: "unknown",
		},
		{
			name: "left",
			event: MouseEvent{
				Action: MouseActionPress,
				Button: MouseButtonLeft,
				Type:   MouseLeft,
			},
			expected: "left press",
		},
		{
			name: "right",
			event: MouseEvent{
				Action: MouseActionPress,
				Button: MouseButtonRight,
				Type:   MouseRight,
			},
			expected: "right press",
		},
		{
			name: "middle",
			event: MouseEvent{
				Action: MouseActionPress,
				Button: MouseButtonMiddle,
				Type:   MouseMiddle,
			},
			expected: "middle press",
		},
		{
			name: "release",
			event: MouseEvent{
				Action: MouseActionRelease,
				Button: MouseButtonNone,
				Type:   MouseRelease,
			},
			expected: "release",
		},
		{
			name: "wheel up",
			event: MouseEvent{
				Action: MouseActionPress,
				Button: MouseButtonWheelUp,
				Type:   MouseWheelUp,
			},
			expected: "wheel up",
		},
		{
			name: "wheel down",
			event: MouseEvent{
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
				Type:   MouseWheelDown,
			},
			expected: "wheel down",
		},
		{
			name: "wheel left",
			event: MouseEvent{
				Action: MouseActionPress,
				Button: MouseButtonWheelLeft,
				Type:   MouseWheelLeft,
			},
			expected: "wheel left",
		},
		{
			name: "wheel right",
			event: MouseEvent{
				Action: MouseActionPress,
				Button: MouseButtonWheelRight,
				Type:   MouseWheelRight,
			},
			expected: "wheel right",
		},
		{
			name: "motion",
			event: MouseEvent{
				Action: MouseActionMotion,
				Button: MouseButtonNone,
				Type:   MouseMotion,
			},
			expected: "motion",
		},
		{
			name: "shift+left release",
			event: MouseEvent{
				Type:   MouseLeft,
				Action: MouseActionRelease,
				Button: MouseButtonLeft,
				Shift:  true,
			},
			expected: "shift+left release",
		},
		{
			name: "shift+left",
			event: MouseEvent{
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
				Shift:  true,
			},
			expected: "shift+left press",
		},
		{
			name: "ctrl+shift+left",
			event: MouseEvent{
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
				Shift:  true,
				Ctrl:   true,
			},
			expected: "ctrl+shift+left press",
		},
		{
			name: "alt+left",
			event: MouseEvent{
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
				Alt:    true,
			},
			expected: "alt+left press",
		},
		{
			name: "ctrl+left",
			event: MouseEvent{
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
				Ctrl:   true,
			},
			expected: "ctrl+left press",
		},
		{
			name: "ctrl+alt+left",
			event: MouseEvent{
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
				Alt:    true,
				Ctrl:   true,
			},
			expected: "ctrl+alt+left press",
		},
		{
			name: "ctrl+alt+shift+left",
			event: MouseEvent{
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
				Alt:    true,
				Ctrl:   true,
				Shift:  true,
			},
			expected: "ctrl+alt+shift+left press",
		},
		{
			name: "ignore coordinates",
			event: MouseEvent{
				X:      100,
				Y:      200,
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
			},
			expected: "left press",
		},
		{
			name: "broken type",
			event: MouseEvent{
				Type:   MouseEventType(-100),
				Action: MouseAction(-110),
				Button: MouseButton(-120),
			},
			expected: "",
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := tc.event.String()

			if tc.expected != actual {
				t.Fatalf("expected %q but got %q",
					tc.expected,
					actual,
				)
			}
		})
	}
}

func TestParseX10MouseEvent(t *testing.T) {
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
		expected MouseEvent
	}{
		// Position.
		{
			name: "zero position",
			buf:  encode(0b0000_0000, 0, 0),
			expected: MouseEvent{
				X:      0,
				Y:      0,
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
			},
		},
		{
			name: "max position",
			buf:  encode(0b0000_0000, 222, 222), // Because 255 (max int8) - 32 - 1.
			expected: MouseEvent{
				X:      222,
				Y:      222,
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
			},
		},
		// Simple.
		{
			name: "left",
			buf:  encode(0b0000_0000, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
			},
		},
		{
			name: "left in motion",
			buf:  encode(0b0010_0000, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseLeft,
				Action: MouseActionMotion,
				Button: MouseButtonLeft,
			},
		},
		{
			name: "middle",
			buf:  encode(0b0000_0001, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseMiddle,
				Action: MouseActionPress,
				Button: MouseButtonMiddle,
			},
		},
		{
			name: "middle in motion",
			buf:  encode(0b0010_0001, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseMiddle,
				Action: MouseActionMotion,
				Button: MouseButtonMiddle,
			},
		},
		{
			name: "right",
			buf:  encode(0b0000_0010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseRight,
				Action: MouseActionPress,
				Button: MouseButtonRight,
			},
		},
		{
			name: "right in motion",
			buf:  encode(0b0010_0010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseRight,
				Action: MouseActionMotion,
				Button: MouseButtonRight,
			},
		},
		{
			name: "motion",
			buf:  encode(0b0010_0011, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseMotion,
				Action: MouseActionMotion,
				Button: MouseButtonNone,
			},
		},
		{
			name: "wheel up",
			buf:  encode(0b0100_0000, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseWheelUp,
				Action: MouseActionPress,
				Button: MouseButtonWheelUp,
			},
		},
		{
			name: "wheel down",
			buf:  encode(0b0100_0001, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseWheelDown,
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
			},
		},
		{
			name: "wheel left",
			buf:  encode(0b0100_0010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseWheelLeft,
				Action: MouseActionPress,
				Button: MouseButtonWheelLeft,
			},
		},
		{
			name: "wheel right",
			buf:  encode(0b0100_0011, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseWheelRight,
				Action: MouseActionPress,
				Button: MouseButtonWheelRight,
			},
		},
		{
			name: "release",
			buf:  encode(0b0000_0011, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseRelease,
				Action: MouseActionRelease,
				Button: MouseButtonNone,
			},
		},
		{
			name: "backward",
			buf:  encode(0b1000_0000, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseBackward,
				Action: MouseActionPress,
				Button: MouseButtonBackward,
			},
		},
		{
			name: "forward",
			buf:  encode(0b1000_0001, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseForward,
				Action: MouseActionPress,
				Button: MouseButtonForward,
			},
		},
		{
			name: "button 10",
			buf:  encode(0b1000_0010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseUnknown,
				Action: MouseActionPress,
				Button: MouseButton10,
			},
		},
		{
			name: "button 11",
			buf:  encode(0b1000_0011, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseUnknown,
				Action: MouseActionPress,
				Button: MouseButton11,
			},
		},
		// Combinations.
		{
			name: "alt+right",
			buf:  encode(0b0000_1010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Type:   MouseRight,
				Action: MouseActionPress,
				Button: MouseButtonRight,
			},
		},
		{
			name: "ctrl+right",
			buf:  encode(0b0001_0010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Ctrl:   true,
				Type:   MouseRight,
				Action: MouseActionPress,
				Button: MouseButtonRight,
			},
		},
		{
			name: "left in motion",
			buf:  encode(0b0010_0000, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    false,
				Type:   MouseLeft,
				Action: MouseActionMotion,
				Button: MouseButtonLeft,
			},
		},
		{
			name: "alt+right in motion",
			buf:  encode(0b0010_1010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Type:   MouseRight,
				Action: MouseActionMotion,
				Button: MouseButtonRight,
			},
		},
		{
			name: "ctrl+right in motion",
			buf:  encode(0b0011_0010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Ctrl:   true,
				Type:   MouseRight,
				Action: MouseActionMotion,
				Button: MouseButtonRight,
			},
		},
		{
			name: "ctrl+alt+right",
			buf:  encode(0b0001_1010, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Ctrl:   true,
				Type:   MouseRight,
				Action: MouseActionPress,
				Button: MouseButtonRight,
			},
		},
		{
			name: "ctrl+wheel up",
			buf:  encode(0b0101_0000, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Ctrl:   true,
				Type:   MouseWheelUp,
				Action: MouseActionPress,
				Button: MouseButtonWheelUp,
			},
		},
		{
			name: "alt+wheel down",
			buf:  encode(0b0100_1001, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Type:   MouseWheelDown,
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
			},
		},
		{
			name: "ctrl+alt+wheel down",
			buf:  encode(0b0101_1001, 32, 16),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Ctrl:   true,
				Type:   MouseWheelDown,
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
			},
		},
		// Overflow position.
		{
			name: "overflow position",
			buf:  encode(0b0010_0000, 250, 223), // Because 255 (max int8) - 32 - 1.
			expected: MouseEvent{
				X:      -6,
				Y:      -33,
				Type:   MouseLeft,
				Action: MouseActionMotion,
				Button: MouseButtonLeft,
			},
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

// func TestParseX10MouseEvent_error(t *testing.T) {
// 	tt := []struct {
// 		name string
// 		buf  []byte
// 	}{
// 		{
// 			name: "empty buf",
// 			buf:  nil,
// 		},
// 		{
// 			name: "wrong high bit",
// 			buf:  []byte("\x1a[M@A1"),
// 		},
// 		{
// 			name: "short buf",
// 			buf:  []byte("\x1b[M@A"),
// 		},
// 		{
// 			name: "long buf",
// 			buf:  []byte("\x1b[M@A11"),
// 		},
// 	}
//
// 	for i := range tt {
// 		tc := tt[i]
//
// 		t.Run(tc.name, func(t *testing.T) {
// 			_, err := parseX10MouseEvent(tc.buf)
//
// 			if err == nil {
// 				t.Fatalf("expected error but got nil")
// 			}
// 		})
// 	}
// }

func TestParseSGRMouseEvent(t *testing.T) {
	encode := func(b, x, y int, r bool) []byte {
		re := 'M'
		if r {
			re = 'm'
		}
		return []byte(fmt.Sprintf("\x1b[<%d;%d;%d%c", b, x+1, y+1, re))
	}

	tt := []struct {
		name     string
		buf      []byte
		expected MouseEvent
	}{
		// Position.
		{
			name: "zero position",
			buf:  encode(0, 0, 0, false),
			expected: MouseEvent{
				X:      0,
				Y:      0,
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
			},
		},
		{
			name: "225 position",
			buf:  encode(0, 225, 225, false),
			expected: MouseEvent{
				X:      225,
				Y:      225,
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
			},
		},
		// Simple.
		{
			name: "left",
			buf:  encode(0, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseLeft,
				Action: MouseActionPress,
				Button: MouseButtonLeft,
			},
		},
		{
			name: "left in motion",
			buf:  encode(32, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseLeft,
				Action: MouseActionMotion,
				Button: MouseButtonLeft,
			},
		},
		{
			name: "left release",
			buf:  encode(0, 32, 16, true),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseRelease,
				Action: MouseActionRelease,
				Button: MouseButtonLeft,
			},
		},
		{
			name: "middle",
			buf:  encode(1, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseMiddle,
				Action: MouseActionPress,
				Button: MouseButtonMiddle,
			},
		},
		{
			name: "middle in motion",
			buf:  encode(33, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseMiddle,
				Action: MouseActionMotion,
				Button: MouseButtonMiddle,
			},
		},
		{
			name: "middle release",
			buf:  encode(1, 32, 16, true),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseRelease,
				Action: MouseActionRelease,
				Button: MouseButtonMiddle,
			},
		},
		{
			name: "right",
			buf:  encode(2, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseRight,
				Action: MouseActionPress,
				Button: MouseButtonRight,
			},
		},
		{
			name: "right release",
			buf:  encode(2, 32, 16, true),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseRelease,
				Action: MouseActionRelease,
				Button: MouseButtonRight,
			},
		},
		{
			name: "motion",
			buf:  encode(35, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseMotion,
				Action: MouseActionMotion,
				Button: MouseButtonNone,
			},
		},
		{
			name: "wheel up",
			buf:  encode(64, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseWheelUp,
				Action: MouseActionPress,
				Button: MouseButtonWheelUp,
			},
		},
		{
			name: "wheel down",
			buf:  encode(65, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseWheelDown,
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
			},
		},
		{
			name: "wheel left",
			buf:  encode(66, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseWheelLeft,
				Action: MouseActionPress,
				Button: MouseButtonWheelLeft,
			},
		},
		{
			name: "wheel right",
			buf:  encode(67, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseWheelRight,
				Action: MouseActionPress,
				Button: MouseButtonWheelRight,
			},
		},
		{
			name: "backward",
			buf:  encode(128, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseBackward,
				Action: MouseActionPress,
				Button: MouseButtonBackward,
			},
		},
		{
			name: "backward in motion",
			buf:  encode(160, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseBackward,
				Action: MouseActionMotion,
				Button: MouseButtonBackward,
			},
		},
		{
			name: "forward",
			buf:  encode(129, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseForward,
				Action: MouseActionPress,
				Button: MouseButtonForward,
			},
		},
		{
			name: "forward in motion",
			buf:  encode(161, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Type:   MouseForward,
				Action: MouseActionMotion,
				Button: MouseButtonForward,
			},
		},
		// Combinations.
		{
			name: "alt+right",
			buf:  encode(10, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Type:   MouseRight,
				Action: MouseActionPress,
				Button: MouseButtonRight,
			},
		},
		{
			name: "ctrl+right",
			buf:  encode(18, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Ctrl:   true,
				Type:   MouseRight,
				Action: MouseActionPress,
				Button: MouseButtonRight,
			},
		},
		{
			name: "ctrl+alt+right",
			buf:  encode(26, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Ctrl:   true,
				Type:   MouseRight,
				Action: MouseActionPress,
				Button: MouseButtonRight,
			},
		},
		{
			name: "alt+wheel press",
			buf:  encode(73, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Type:   MouseWheelDown,
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
			},
		},
		{
			name: "ctrl+wheel press",
			buf:  encode(81, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Ctrl:   true,
				Type:   MouseWheelDown,
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
			},
		},
		{
			name: "ctrl+alt+wheel press",
			buf:  encode(89, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Alt:    true,
				Ctrl:   true,
				Type:   MouseWheelDown,
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
			},
		},
		{
			name: "ctrl+alt+shift+wheel press",
			buf:  encode(93, 32, 16, false),
			expected: MouseEvent{
				X:      32,
				Y:      16,
				Shift:  true,
				Alt:    true,
				Ctrl:   true,
				Type:   MouseWheelDown,
				Action: MouseActionPress,
				Button: MouseButtonWheelDown,
			},
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
