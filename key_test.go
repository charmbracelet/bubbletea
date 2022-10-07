package tea

import (
	"bytes"
	"fmt"
	"testing"
)

func TestKeyString(t *testing.T) {
	t.Run("alt+space", func(t *testing.T) {
		if got := KeyMsg(Key{
			Type: KeySpace,
			Alt:  true,
		}).String(); got != "alt+ " {
			t.Fatalf(`expected a "alt+ ", got %q`, got)
		}
	})

	t.Run("runes", func(t *testing.T) {
		if got := KeyMsg(Key{
			Type:  KeyRunes,
			Runes: []rune{'a'},
		}).String(); got != "a" {
			t.Fatalf(`expected an "a", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if got := KeyMsg(Key{
			Type: KeyType(99999),
		}).String(); got != "" {
			t.Fatalf(`expected a "", got %q`, got)
		}
	})
}

func TestKeyTypeString(t *testing.T) {
	t.Run("space", func(t *testing.T) {
		if got := KeySpace.String(); got != " " {
			t.Fatalf(`expected a " ", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if got := KeyType(99999).String(); got != "" {
			t.Fatalf(`expected a "", got %q`, got)
		}
	})
}

func TestReadInput(t *testing.T) {
	type test struct {
		keyname string
		in      []byte
		out     []Msg
	}
	for i, td := range []test{
		{"a",
			[]byte{'a'},
			[]Msg{
				KeyMsg{
					Type:  KeyRunes,
					Runes: []rune{'a'},
				},
			},
		},
		{" ",
			[]byte{' '},
			[]Msg{
				KeyMsg{
					Type:  KeySpace,
					Runes: []rune{' '},
				},
			},
		},
		{"ctrl+a",
			[]byte{byte(keySOH)},
			[]Msg{
				KeyMsg{
					Type: KeyCtrlA,
				},
			},
		},
		{"alt+a",
			[]byte{byte(0x1b), 'a'},
			[]Msg{
				KeyMsg{
					Type:  KeyRunes,
					Alt:   true,
					Runes: []rune{'a'},
				},
			},
		},
		{"abcd",
			[]byte{'a', 'b', 'c', 'd'},
			[]Msg{
				KeyMsg{
					Type:  KeyRunes,
					Runes: []rune{'a'},
				},
				KeyMsg{
					Type:  KeyRunes,
					Runes: []rune{'b'},
				},
				KeyMsg{
					Type:  KeyRunes,
					Runes: []rune{'c'},
				},
				KeyMsg{
					Type:  KeyRunes,
					Runes: []rune{'d'},
				},
			},
		},
		{"up",
			[]byte("\x1b[A"),
			[]Msg{
				KeyMsg{
					Type: KeyUp,
				},
			},
		},
		{"wheel up",
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Msg{
				MouseMsg{
					Type: MouseWheelUp,
				},
			},
		},
		{"shift+tab",
			[]byte{'\x1b', '[', 'Z'},
			[]Msg{
				KeyMsg{
					Type: KeyShiftTab,
				},
			},
		},
		{"alt+enter",
			[]byte{'\x1b', '\r'},
			[]Msg{
				KeyMsg{
					Type: KeyEnter,
					Alt:  true,
				},
			},
		},
		{"insert",
			[]byte{'\x1b', '[', '2', '~'},
			[]Msg{
				KeyMsg{
					Type: KeyInsert,
				},
			},
		},
		{"alt+ctrl+a",
			[]byte{'\x1b', byte(keySOH)},
			[]Msg{
				KeyMsg{
					Type: KeyCtrlA,
					Alt:  true,
				},
			},
		},
		{"unrecognized CSI",
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Msg{},
		},
		// Powershell sequences.
		{"up",
			[]byte{'\x1b', 'O', 'A'},
			[]Msg{KeyMsg{Type: KeyUp}},
		},
		{"down",
			[]byte{'\x1b', 'O', 'B'},
			[]Msg{KeyMsg{Type: KeyDown}},
		},
		{"right",
			[]byte{'\x1b', 'O', 'C'},
			[]Msg{KeyMsg{Type: KeyRight}},
		},
		{"left",
			[]byte{'\x1b', 'O', 'D'},
			[]Msg{KeyMsg{Type: KeyLeft}},
		},
		{"alt+enter",
			[]byte{'\x1b', '\x0d'},
			[]Msg{KeyMsg{Type: KeyEnter, Alt: true}},
		},
		{"alt+backspace",
			[]byte{'\x1b', '\x7f'},
			[]Msg{KeyMsg{Type: KeyBackspace, Alt: true}},
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, td.keyname), func(t *testing.T) {
			msgs, err := readInputs(bytes.NewReader(td.in))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(msgs) != len(td.out) {
				t.Fatalf("unexpected message list length")
			}

			if len(msgs) == 1 {
				if m, ok := msgs[0].(KeyMsg); ok && m.String() != td.keyname {
					t.Fatalf(`expected a keymsg %q, got %q`, td.keyname, m)
				}
			}

			for i, v := range msgs {
				if m, ok := v.(KeyMsg); ok &&
					m.String() != td.out[i].(KeyMsg).String() {
					t.Fatalf(`expected a keymsg %q, got %q`, td.out[i].(KeyMsg), m)
				}
				if m, ok := v.(MouseMsg); ok &&
					(mouseEventTypes[m.Type] != td.keyname || m.Type != td.out[i].(MouseMsg).Type) {
					t.Fatalf(`expected a mousemsg %q, got %q`,
						td.keyname,
						mouseEventTypes[td.out[i].(MouseMsg).Type])
				}
			}
		})
	}
}
