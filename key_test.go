package tea

import (
	"bytes"
	"testing"
)

func TestKeyString(t *testing.T) {
	t.Run("alt+space", func(t *testing.T) {
		if got := KeyMsg(Key{
			Type: KeySpace,
			Alt:  true,
		}).String(); got != "alt+space" {
			t.Fatalf(`expected a "alt+space", got %q`, got)
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
	for out, in := range map[string][]byte{
		"a":         {'a'},
		"ctrl+a":    {byte(keySOH)},
		"alt+a":     {0x1b, 'a'},
		"abcd":      {'a', 'b', 'c', 'd'},
		"up":        []byte("\x1b[A"),
		"wheel up":  {'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
		"shift+tab": {'\x1b', '[', 'Z'},
	} {
		t.Run(out, func(t *testing.T) {
			msgs, err := readInputs(bytes.NewReader(in))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(msgs) == 0 {
				t.Fatalf("unexpected empty message list")
			}

			if m, ok := msgs[0].(KeyMsg); ok && m.String() != out {
				t.Fatalf(`expected a keymsg %q, got %q`, out, m)
			}
			if m, ok := msgs[0].(MouseMsg); ok && mouseEventTypes[m.Type] != out {
				t.Fatalf(`expected a mousemsg %q, got %q`, out, mouseEventTypes[m.Type])
			}
		})
	}
}
