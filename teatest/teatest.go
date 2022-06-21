// Package teatest provides helper functions to test tea.Model's.
package teatest

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Program defines the subset of the tea.Program API we need for testing.
type Program interface {
	Send(tea.Msg)
}

// TestModel tests a given model with the given interactions and assertions.
func TestModel(
	tb testing.TB,
	m tea.Model,
	interact func(p Program, in io.Writer),
	assert func(out []byte),
) {
	var in bytes.Buffer
	var out bytes.Buffer

	p := tea.NewProgram(m, tea.WithInput(&in), tea.WithOutput(&out), tea.WithoutSignals())

	done := make(chan bool, 1)
	go func() {
		if err := p.Start(); err != nil {
			tb.Fatalf("app failed: %s", err)
		}
		done <- true
	}()

	// run the user interactions and then force a quit
	interact(p, &in)
	p.Quit()

	// wait for the program to quit and assert
	<-done
	assert(out.Bytes())
}

var update = flag.Bool("update", false, "update .golden files")

// RequireEqualOutput is a helper function to assert the given output is the
// the expected from the golden files.
//
// You can update the golden files by running your tests with the -update flag.
func RequireEqualOutput(tb testing.TB, out []byte) {
	tb.Helper()

	golden := filepath.Join("testdata", tb.Name()+".golden")
	if *update {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			tb.Fatal(err)
		}
		if err := os.WriteFile(golden, out, 0o600); err != nil {
			tb.Fatal(err)
		}
	}

	gbts, err := os.ReadFile(golden)
	if err != nil {
		tb.Fatal(err)
	}

	if bytes.Equal(gbts, out) {
		tb.Fatalf("output do not match:\ngot:\n%s\n\nexpected:\n%s\n\n", out, gbts)
	}
}
