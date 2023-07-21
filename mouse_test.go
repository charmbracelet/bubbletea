package tea

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMouseEvent_String(t *testing.T) {
	for name, test := range map[string]struct {
		event    MouseEvent
		expected string
	}{
		"unknown": {
			event:    MouseEvent{Type: MouseUnknown},
			expected: "unknown",
		},
		"left": {
			event:    MouseEvent{Type: MouseLeft},
			expected: "left",
		},
		"right": {
			event:    MouseEvent{Type: MouseRight},
			expected: "right",
		},
		"middle": {
			event:    MouseEvent{Type: MouseMiddle},
			expected: "middle",
		},
		"release": {
			event:    MouseEvent{Type: MouseRelease},
			expected: "release",
		},
		"wheel up": {
			event:    MouseEvent{Type: MouseWheelUp},
			expected: "wheel up",
		},
		"wheel down": {
			event:    MouseEvent{Type: MouseWheelDown},
			expected: "wheel down",
		},
		"motion": {
			event:    MouseEvent{Type: MouseMotion},
			expected: "motion",
		},
		"alt+left": {
			event: MouseEvent{
				Type: MouseLeft,
				Alt:  true,
			},
			expected: "alt+left",
		},
		"ctrl+left": {
			event: MouseEvent{
				Type: MouseLeft,
				Ctrl: true,
			},
			expected: "ctrl+left",
		},
		"ctrl+alt+left": {
			event: MouseEvent{
				Type: MouseLeft,
				Alt:  true,
				Ctrl: true,
			},
			expected: "ctrl+alt+left",
		},
		"ignore coordinates": {
			event: MouseEvent{
				X:    100,
				Y:    200,
				Type: MouseLeft,
			},
			expected: "left",
		},
		"broken type": {
			event: MouseEvent{
				Type: MouseEventType(-1000),
			},
			expected: "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.event.String())
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

	for name, test := range map[string]struct {
		buf      []byte
		expected MouseEvent
	}{
		// Position.
		"zero position": {
			buf: encode(0b0010_0000, 0, 0),
			expected: MouseEvent{
				X:    0,
				Y:    0,
				Type: MouseLeft,
			},
		},
		"max position": {
			buf: encode(0b0010_0000, 222, 222), // Because 255 (max int8) - 32 - 1.
			expected: MouseEvent{
				X:    222,
				Y:    222,
				Type: MouseLeft,
			},
		},
		// Simple.
		"left": {
			buf: encode(0b0000_0000, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseLeft,
			},
		},
		"left in motion": {
			buf: encode(0b0010_0000, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseLeft,
			},
		},
		"middle": {
			buf: encode(0b0000_0001, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseMiddle,
			},
		},
		"middle in motion": {
			buf: encode(0b0010_0001, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseMiddle,
			},
		},
		"right": {
			buf: encode(0b0000_0010, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseRight,
			},
		},
		"right in motion": {
			buf: encode(0b0010_0010, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseRight,
			},
		},
		"motion": {
			buf: encode(0b0010_0011, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseMotion,
			},
		},
		"wheel up": {
			buf: encode(0b0100_0000, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseWheelUp,
			},
		},
		"wheel down": {
			buf: encode(0b0100_0001, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseWheelDown,
			},
		},
		"release": {
			buf: encode(0b0000_0011, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseRelease,
			},
		},
		// Combinations.
		"alt+right": {
			buf: encode(0b0010_1010, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseRight,
				Alt:  true,
			},
		},
		"ctrl+right": {
			buf: encode(0b0011_0010, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseRight,
				Ctrl: true,
			},
		},
		"ctrl+alt+right": {
			buf: encode(0b0011_1010, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseRight,
				Alt:  true,
				Ctrl: true,
			},
		},
		"alt+wheel down": {
			buf: encode(0b0100_1001, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseWheelDown,
				Alt:  true,
			},
		},
		"ctrl+wheel down": {
			buf: encode(0b0101_0001, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseWheelDown,
				Ctrl: true,
			},
		},
		"ctrl+alt+wheel down": {
			buf: encode(0b0101_1001, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseWheelDown,
				Alt:  true,
				Ctrl: true,
			},
		},
		// Unknown.
		"wheel with unknown bit": {
			buf: encode(0b0100_0010, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseUnknown,
			},
		},
		"unknown with modifier": {
			buf: encode(0b0100_1010, 32, 16),
			expected: MouseEvent{
				X:    32,
				Y:    16,
				Type: MouseUnknown,
				Alt:  true,
			},
		},
		// Overflow position.
		"overflow position": {
			buf: encode(0b0010_0000, 250, 223), // Because 255 (max int8) - 32 - 1.
			expected: MouseEvent{
				X:    -6,
				Y:    -33,
				Type: MouseLeft,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, parseX10MouseEvent(test.buf))
		})
	}
}
