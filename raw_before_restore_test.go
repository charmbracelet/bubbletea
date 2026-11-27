package tea

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

// TestRawBeforeQuitWritesBeforeTerminalRestore guards the ordering of output
// that is still queued when the program quits: it must reach the terminal
// before the renderer leaves the alternate screen, otherwise a raw sequence
// meant for the program's screen lands on the primary one.
func TestRawBeforeQuitWritesBeforeTerminalRestore(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	p := NewProgram(rawQuitModel{},
		WithFPS(1),
		WithWindowSize(80, 24),
		WithColorProfile(colorprofile.ANSI256),
		WithEnvironment([]string{"TERM=xterm-256color"}),
		WithInput(&in),
		WithOutput(&buf),
	)
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	out := buf.Bytes()
	raw := bytes.Index(out, []byte("RAWMARK"))
	exit := bytes.Index(out, []byte("\x1b[?1049l"))
	if raw < 0 {
		t.Fatalf("raw output missing from %q", out)
	}
	if exit < 0 {
		t.Fatalf("alternate screen exit missing from %q", out)
	}
	if raw > exit {
		t.Fatalf("raw output at %d came after the alternate screen exit at %d: %q", raw, exit, out)
	}
}

type rawQuitModel struct{}

func (rawQuitModel) Init() Cmd {
	return Sequence(Raw("\x1b_RAWMARK\x1b\\"), Quit)
}

func (rawQuitModel) Update(Msg) (Model, Cmd) { return rawQuitModel{}, nil }

func (rawQuitModel) View() View {
	v := NewView("hello")
	v.AltScreen = true
	return v
}

var _ = strings.TrimSpace
