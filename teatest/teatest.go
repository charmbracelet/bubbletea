// Package teatest provides helper functions to test tea.Model's.
package teatest

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
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
	if err := p.ReleaseTerminal(); err != nil {
		tb.Fatalf("could not restore terminal: %v", err)
	}

	// wait for the program to quit and assert
	<-done
	assert(out.Bytes())
}

// TypeText types the given text into the given program.
func TypeText(p Program, s string) {
	for _, c := range []byte(s) {
		p.Send(tea.KeyMsg{
			Runes: []rune{rune(c)},
			Type:  tea.KeyRunes,
		})
	}
}

var update = flag.Bool("update", false, "update .golden files")

// RequireEqualOutput is a helper function to assert the given output is
// the expected from the golden files, printing its diff in case it is not.
//
// Important: this uses the system `diff` tool.
//
// You can update the golden files by running your tests with the -update flag.
func RequireEqualOutput(tb testing.TB, out []byte) {
	tb.Helper()

	golden := filepath.Join("testdata", tb.Name()+".golden")
	if *update {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil { // nolint: gomnd
			tb.Fatal(err)
		}
		if err := os.WriteFile(golden, out, 0o600); err != nil { // nolint: gomnd
			tb.Fatal(err)
		}
	}

	path := filepath.Join(tb.TempDir(), tb.Name()+".out")
	if err := os.WriteFile(path, out, 0o600); err != nil { // nolint: gomnd
		tb.Fatal(err)
	}

	// inspired by https://cs.opensource.google/go/go/+/refs/tags/go1.18.3:src/cmd/internal/diff/diff.go;l=18
	diff, err := exec.Command("diff", path, golden).CombinedOutput()
	if err != nil {
		tb.Fatalf("output does not match, diff:\n\n%s", string(diff))
	}
}
