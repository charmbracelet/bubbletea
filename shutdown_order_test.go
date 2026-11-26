package tea

import (
	"bytes"
	"strings"
	"testing"
)

type rawQuitModel struct{}

func (rawQuitModel) Init() Cmd                 { return Sequence(Raw("RAWMARK"), Quit) }
func (m rawQuitModel) Update(Msg) (Model, Cmd) { return m, nil }
func (rawQuitModel) View() View                { v := NewView("hello"); v.AltScreen = true; return v }

// Raw output sent before Quit belongs to the program's screen, so it must
// reach the output before the alt screen is left.
func TestRawBeforeQuitPrecedesRestore(t *testing.T) {
	for range 20 {
		var buf, in bytes.Buffer
		if _, err := NewProgram(rawQuitModel{}, WithInput(&in), WithOutput(&buf)).Run(); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		m, r := strings.Index(out, "RAWMARK"), strings.LastIndex(out, "\x1b[?1049l")
		if m < 0 || r < 0 {
			t.Fatalf("mark at %d, alt screen exit at %d in %q", m, r, out)
		}
		if m > r {
			t.Fatalf("raw output written after the alt screen exit: %q", out[r:])
		}
	}
}
