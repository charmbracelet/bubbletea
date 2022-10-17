// Package teatest provides helper functions to test tea.Model's.
package teatest

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Program defines the subset of the tea.Program API we need for testing.
type Program interface {
	Send(tea.Msg)
}

// TestModelOptions defines all options available to the test function.
type TestModelOptions struct {
	interact      func(p Program, in io.Writer)
	assert        func(out []byte)
	validateModel func(m tea.Model) error
}

// TestOption is a functional option.
type TestOption func(opts *TestModelOptions)

// WithProgramInteractions ...
func WithProgramInteractions(fn func(p Program, in io.Writer)) TestOption {
	return func(opts *TestModelOptions) {
		opts.interact = fn
	}
}

// WithRequiredOutputChecker ...
func WithRequiredOutputChecker(fn func(out []byte)) TestOption {
	return func(opts *TestModelOptions) {
		opts.assert = fn
	}
}

// WithValidateFinalModel ...
func WithValidateFinalModel(fn func(m tea.Model) error) TestOption {
	return func(opts *TestModelOptions) {
		opts.validateModel = fn
	}
}

// TestModel tests a given model with the given interactions and assertions.
func TestModel(tb testing.TB, m tea.Model, options ...TestOption) {
	var in bytes.Buffer
	var out bytes.Buffer

	p := tea.NewProgram(m, tea.WithInput(&in), tea.WithOutput(safe(&out)), tea.WithoutSignals())

	ints := make(chan os.Signal, 1)
	signal.Notify(ints, syscall.SIGINT)
	done := make(chan bool, 1)
	go func() {
		if _, err := p.Run(); err != nil {
			tb.Fatalf("app failed: %s", err)
		}
		done <- true
	}()
	go func() {
		<-ints
		signal.Stop(ints)
		tb.Log("interrupted")
		p.Quit()
	}()

	var opts TestModelOptions
	for _, opt := range options {
		opt(&opts)
	}

	// run the user interactions and then force a quit
	if opts.interact != nil {
		opts.interact(p, safe(&in))
	}
	p.Quit()
	if err := p.ReleaseTerminal(); err != nil {
		tb.Fatalf("could not restore terminal: %v", err)
	}

	// wait for the program to quit and assert
	<-done

	if opts.validateModel != nil {
		if err := opts.validateModel(m); err != nil {
			tb.Fatalf("failed to validate model: %v", err)
		}
	}

	if opts.assert != nil {
		opts.assert(out.Bytes())
	}
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { // nolint: gomnd
		tb.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil { // nolint: gomnd
		tb.Fatal(err)
	}

	// inspired by https://cs.opensource.google/go/go/+/refs/tags/go1.18.3:src/cmd/internal/diff/diff.go;l=18
	diff, err := exec.Command("diff", path, golden).CombinedOutput()
	if err != nil {
		tb.Fatalf("output does not match, diff:\n\n%s", string(diff))
	}
}

func safe(w io.Writer) io.Writer {
	return &safeWriter{w: w}
}

type safeWriter struct {
	w io.Writer
	m sync.Mutex
}

func (s *safeWriter) Write(p []byte) (int, error) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.w.Write(p)
}
