// Package teatest provides helper functions to test tea.Model's.
package teatest

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/aymanbagabas/go-udiff"
	tea "github.com/charmbracelet/bubbletea"
)

// Program defines the subset of the tea.Program API we need for testing.
type Program interface {
	Send(tea.Msg)
}

// TestModelOptions defines all options available to the test function.
type TestModelOptions struct {
	size tea.WindowSizeMsg
}

// TestOption is a functional option.
type TestOption func(opts *TestModelOptions)

// WithInitialTermSize ...
func WithInitialTermSize(x, y int) TestOption {
	return func(opts *TestModelOptions) {
		opts.size = tea.WindowSizeMsg{
			Width:  x,
			Height: y,
		}
	}
}

// WaitingForContext is the context for a WaitFor.
type WaitingForContext struct {
	Duration      time.Duration
	CheckInterval time.Duration
}

// WaitForOption changes how a WaitFor will behave.
type WaitForOption func(*WaitingForContext)

// WithCheckInterval sets how much time a WaitFor should sleep between every
// check.
func WithCheckInterval(d time.Duration) WaitForOption {
	return func(wf *WaitingForContext) {
		wf.CheckInterval = d
	}
}

// WithDuration sets how much time a WaitFor will wait for the condition.
func WithDuration(d time.Duration) WaitForOption {
	return func(wf *WaitingForContext) {
		wf.Duration = d
	}
}

// WaitFor keeps reading from r until the condition matches.
// Default duration is 1s, default check interval is 50ms.
// These defaults can be changed with WithDuration and WithCheckInterval.
func WaitFor(
	tb testing.TB,
	r io.Reader,
	condition func(bts []byte) bool,
	options ...WaitForOption,
) {
	tb.Helper()

	wf := WaitingForContext{
		Duration:      time.Second,
		CheckInterval: 50 * time.Millisecond,
	}

	for _, opt := range options {
		opt(&wf)
	}

	var b bytes.Buffer
	start := time.Now()
	for time.Since(start) <= wf.Duration {
		if _, err := io.ReadAll(io.TeeReader(r, &b)); err != nil {
			tb.Fatal("WaitFor:", err)
		}
		if condition(b.Bytes()) {
			return
		}
		time.Sleep(wf.CheckInterval)
	}
	tb.Fatalf("WaitFor: condition not met after %s", wf.Duration)
}

// TestModel is a model that is being tested.
type TestModel struct {
	program *tea.Program

	in  *bytes.Buffer
	out *bytes.Buffer

	modelCh chan tea.Model
	model   tea.Model

	done   sync.Once
	doneCh chan bool
}

// NewTestModel makes a new TestModel which can be used for tests.
func NewTestModel(tb testing.TB, m tea.Model, options ...TestOption) *TestModel {
	tm := &TestModel{
		in:      bytes.NewBuffer(nil),
		out:     bytes.NewBuffer(nil),
		modelCh: make(chan tea.Model, 1),
		doneCh:  make(chan bool, 1),
	}

	tm.program = tea.NewProgram(
		m,
		tea.WithInput(tm.in),
		tea.WithOutput(safe(tm.out)),
		tea.WithoutSignals(),
	)

	interruptions := make(chan os.Signal, 1)
	signal.Notify(interruptions, syscall.SIGINT)
	go func() {
		m, err := tm.program.Run()
		if err != nil {
			tb.Fatalf("app failed: %s", err)
		}
		tm.doneCh <- true
		tm.modelCh <- m
	}()
	go func() {
		<-interruptions
		signal.Stop(interruptions)
		tb.Log("interrupted")
		tm.program.Quit()
	}()

	var opts TestModelOptions
	for _, opt := range options {
		opt(&opts)
	}

	if opts.size.Width != 0 {
		tm.program.Send(opts.size)
	}
	return tm
}

func (tm *TestModel) waitDone() {
	tm.done.Do(func() {
		<-tm.doneCh
	})
}

// FinalModel returns the resulting model, resulting from program.Run().
// This method only returns once the program has finished running.
func (tm *TestModel) FinalModel() tea.Model {
	tm.waitDone()
	select {
	case m := <-tm.modelCh:
		tm.model = m
		return tm.model
	default:
		return tm.model
	}
}

// Output returns the program's output io.Reader.
func (tm *TestModel) Output() io.Reader {
	return tm.out
}

// Send sends messages to the underlying program.
func (tm *TestModel) Send(m tea.Msg) {
	tm.program.Send(m)
}

// Quit quits the program and releases the terminal.
func (tm *TestModel) Quit() error {
	tm.program.Quit()
	tm.program.Wait()
	return tm.program.ReleaseTerminal()
}

// Type types the given text into the given program.
func (tm *TestModel) Type(s string) {
	for _, c := range []byte(s) {
		tm.program.Send(tea.KeyMsg{
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

	goldenBts, err := os.ReadFile(golden)
	if err != nil {
		tb.Fatal(err)
	}
	diff := udiff.Unified("golden", "run", string(goldenBts), string(out))
	if diff != "" {
		tb.Fatalf("output does not match, diff:\n\n%s", diff)
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
