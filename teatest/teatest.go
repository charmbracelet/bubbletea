package teatest

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type Sender interface {
	Send(tea.Msg)
}

func TestModel(
	tb testing.TB,
	m tea.Model,
	interact func(p Sender, in io.Writer),
	assert func(out io.Reader),
) {
	var in bytes.Buffer
	var out bytes.Buffer

	p := tea.NewProgram(m, tea.WithInput(&in), tea.WithOutput(&out))
	done := make(chan bool, 1)

	go func() {
		if err := p.Start(); err != nil {
			tb.Fatalf("app failed: %s", err)
		}
		done <- true
	}()

	interact(p, &in)
	<-done
	assert(&out)
}

var update = flag.Bool("update", false, "update .golden files")

func RequireEqualOutput(tb testing.TB, out io.Reader) {
	tb.Helper()

	bts, err := io.ReadAll(out)
	if err != nil {
		tb.Fatal(err)
	}

	golden := "testdata/" + tb.Name() + ".golden"
	if *update {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			tb.Fatal(err)
		}
		if err := os.WriteFile(golden, bts, 0o600); err != nil {
			tb.Fatal(err)
		}
	}

	gbts, err := os.ReadFile(golden)
	if err != nil {
		tb.Fatal(err)
	}

	sg := formatEscapes(string(gbts))
	so := formatEscapes(string(bts))
	if sg != so {
		tb.Fatalf("output do not match:\ngot:\n%s\n\nexpected:\n%s\n\n", so, sg)
	}
}

func formatEscapes(str string) string {
	return strings.ReplaceAll(str, "\x1b", "\\x1b")
}
