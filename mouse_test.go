package tea

import "testing"

func TestMouseEvent_String(t *testing.T) {
	tt := []struct {
		name     string
		event    MouseEvent
		expected string
	}{
		{
			name:     "unknown",
			event:    MouseEvent{Type: MouseUnknown},
			expected: "unknown",
		},
		{
			name:     "left",
			event:    MouseEvent{Type: MouseLeft},
			expected: "left",
		},
		{
			name:     "right",
			event:    MouseEvent{Type: MouseRight},
			expected: "right",
		},
		{
			name:     "middle",
			event:    MouseEvent{Type: MouseMiddle},
			expected: "middle",
		},
		{
			name:     "release",
			event:    MouseEvent{Type: MouseRelease},
			expected: "release",
		},
		{
			name:     "wheel up",
			event:    MouseEvent{Type: MouseWheelUp},
			expected: "wheel up",
		},
		{
			name:     "wheel down",
			event:    MouseEvent{Type: MouseWheelDown},
			expected: "wheel down",
		},
		{
			name:     "motion",
			event:    MouseEvent{Type: MouseMotion},
			expected: "motion",
		},
		{
			name: "alt+left",
			event: MouseEvent{
				Type: MouseLeft,
				Alt:  true,
			},
			expected: "alt+left",
		},
		{
			name: "ctrl+left",
			event: MouseEvent{
				Type: MouseLeft,
				Ctrl: true,
			},
			expected: "ctrl+left",
		},
		{
			name: "ctrl+alt+left",
			event: MouseEvent{
				Type: MouseLeft,
				Alt:  true,
				Ctrl: true,
			},
			expected: "ctrl+alt+left",
		},
		{
			name: "ignore coordinates",
			event: MouseEvent{
				X:    100,
				Y:    200,
				Type: MouseLeft,
			},
			expected: "left",
		},
		{
			name: "broken type",
			event: MouseEvent{
				Type: MouseEventType(-1000),
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
		expected []MouseEvent
	}{
		// Position.
		{
			name: "zero position",
			buf:  encode(0b0010_0000, 0, 0),
			expected: []MouseEvent{
				{
					X:    0,
					Y:    0,
					Type: MouseLeft,
				},
			},
		},
		{
			name: "max position",
			buf:  encode(0b0010_0000, 222, 222), // Because 255 (max int8) - 32 - 1.
			expected: []MouseEvent{
				{
					X:    222,
					Y:    222,
					Type: MouseLeft,
				},
			},
		},
		// Simple.
		{
			name: "left",
			buf:  encode(0b0000_0000, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseLeft,
				},
			},
		},
		{
			name: "left in motion",
			buf:  encode(0b0010_0000, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseLeft,
				},
			},
		},
		{
			name: "middle",
			buf:  encode(0b0000_0001, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseMiddle,
				},
			},
		},
		{
			name: "middle in motion",
			buf:  encode(0b0010_0001, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseMiddle,
				},
			},
		},
		{
			name: "right",
			buf:  encode(0b0000_0010, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseRight,
				},
			},
		},
		{
			name: "right in motion",
			buf:  encode(0b0010_0010, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseRight,
				},
			},
		},
		{
			name: "motion",
			buf:  encode(0b0010_0011, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseMotion,
				},
			},
		},
		{
			name: "wheel up",
			buf:  encode(0b0100_0000, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseWheelUp,
				},
			},
		},
		{
			name: "wheel down",
			buf:  encode(0b0100_0001, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseWheelDown,
				},
			},
		},
		{
			name: "release",
			buf:  encode(0b0000_0011, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseRelease,
				},
			},
		},
		// Combinations.
		{
			name: "alt+right",
			buf:  encode(0b0010_1010, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseRight,
					Alt:  true,
				},
			},
		},
		{
			name: "ctrl+right",
			buf:  encode(0b0011_0010, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseRight,
					Ctrl: true,
				},
			},
		},
		{
			name: "ctrl+alt+right",
			buf:  encode(0b0011_1010, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseRight,
					Alt:  true,
					Ctrl: true,
				},
			},
		},
		{
			name: "alt+wheel down",
			buf:  encode(0b0100_1001, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseWheelDown,
					Alt:  true,
				},
			},
		},
		{
			name: "ctrl+wheel down",
			buf:  encode(0b0101_0001, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseWheelDown,
					Ctrl: true,
				},
			},
		},
		{
			name: "ctrl+alt+wheel down",
			buf:  encode(0b0101_1001, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseWheelDown,
					Alt:  true,
					Ctrl: true,
				},
			},
		},
		// Unknown.
		{
			name: "wheel with unknown bit",
			buf:  encode(0b0100_0010, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseUnknown,
				},
			},
		},
		{
			name: "unknown with modifier",
			buf:  encode(0b0100_1010, 32, 16),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseUnknown,
					Alt:  true,
				},
			},
		},
		// Overflow position.
		{
			name: "overflow position",
			buf:  encode(0b0010_0000, 250, 223), // Because 255 (max int8) - 32 - 1.
			expected: []MouseEvent{
				{
					X:    -6,
					Y:    -33,
					Type: MouseLeft,
				},
			},
		},
		// Batched events.
		{
			name: "batched events",
			buf:  append(encode(0b0010_0000, 32, 16), encode(0b0000_0011, 64, 32)...),
			expected: []MouseEvent{
				{
					X:    32,
					Y:    16,
					Type: MouseLeft,
				},
				{
					X:    64,
					Y:    32,
					Type: MouseRelease,
				},
			},
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseX10MouseEvents(tc.buf)
			if err != nil {
				t.Fatalf("unexpected error for test: %v",
					err,
				)
			}

			for i := range tc.expected {
				if tc.expected[i] != actual[i] {
					t.Fatalf("expected %#v but got %#v",
						tc.expected[i],
						actual[i],
					)
				}
			}
		})
	}
}

func TestParseX10MouseEvent_error(t *testing.T) {
	tt := []struct {
		name string
		buf  []byte
	}{
		{
			name: "empty buf",
			buf:  nil,
		},
		{
			name: "wrong high bit",
			buf:  []byte("\x1a[M@A1"),
		},
		{
			name: "short buf",
			buf:  []byte("\x1b[M@A"),
		},
		{
			name: "long buf",
			buf:  []byte("\x1b[M@A11"),
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			_, err := parseX10MouseEvents(tc.buf)

			if err == nil {
				t.Fatalf("expected error but got nil")
			}
		})
	}
}
