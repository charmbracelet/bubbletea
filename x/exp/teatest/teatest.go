// Package teatest provides helper functions to test tea.Model's.
package teatest

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/aymanbagabas/go-udiff"
	tea "github.com/rprtr258/bubbletea"
	"github.com/stretchr/testify/assert"
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
	if err := doWaitFor(r, condition, options...); err != nil {
		tb.Fatal(err)
	}
}

func doWaitFor(r io.Reader, condition func(bts []byte) bool, options ...WaitForOption) error {
	wf := WaitingForContext{
		Duration:      time.Second,
		CheckInterval: 50 * time.Millisecond, //nolint: gomnd
	}

	for _, opt := range options {
		opt(&wf)
	}

	var b bytes.Buffer
	start := time.Now()
	for time.Since(start) <= wf.Duration {
		if _, err := io.ReadAll(io.TeeReader(r, &b)); err != nil {
			return fmt.Errorf("WaitFor: %w", err)
		}
		if condition(b.Bytes()) {
			return nil
		}
		time.Sleep(wf.CheckInterval)
	}
	return fmt.Errorf("WaitFor: condition not met after %s", wf.Duration)
}

// TestModel is a model that is being tested.
type TestModel[M tea.Model[M]] struct {
	program *tea.Program[M]

	in  *bytes.Buffer
	out io.ReadWriter

	modelCh chan M
	model   M

	done   sync.Once
	doneCh chan bool
}

// NewTestModel makes a new TestModel which can be used for tests.
func NewTestModel[M tea.Model[M]](t *testing.T, m M, options ...TestOption) *TestModel[M] {
	tm := &TestModel[M]{
		in:      bytes.NewBuffer(nil),
		out:     safe(bytes.NewBuffer(nil)),
		modelCh: make(chan M, 1),
		doneCh:  make(chan bool, 1),
	}

	tm.program = tea.NewProgram(m).
		WithInput(tm.in).
		WithOutput(tm.out).
		WithoutSignals()

	interruptions := make(chan os.Signal, 1)
	signal.Notify(interruptions, syscall.SIGINT)
	go func() {
		m, err := tm.program.Run()
		assert.NoError(t, err)
		tm.doneCh <- true
		tm.modelCh <- m
	}()
	go func() {
		<-interruptions
		signal.Stop(interruptions)
		t.Log("interrupted")
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

func (tm *TestModel[M]) waitDone(tb testing.TB, opts []FinalOpt) {
	tm.done.Do(func() {
		fopts := FinalOpts{}
		for _, opt := range opts {
			opt(&fopts)
		}
		if fopts.timeout > 0 {
			select {
			case <-time.After(fopts.timeout):
				tb.Fatalf("timeout after %s", fopts.timeout)
			case <-tm.doneCh:
			}
		} else {
			<-tm.doneCh
		}
	})
}

// FinalOpts represents the options for FinalModel and FinalOutput.
type FinalOpts struct {
	timeout time.Duration
}

// FinalOpt changes FinalOpts.
type FinalOpt func(opts *FinalOpts)

// WithFinalTimeout allows to set a timeout for how long FinalModel and
// FinalOuput should wait for the program to complete.
func WithFinalTimeout(d time.Duration) FinalOpt {
	return func(opts *FinalOpts) {
		opts.timeout = d
	}
}

// WaitFinished waits for the app to finish.
// This method only returns once the program has finished running or when it
// times out.
func (tm *TestModel[M]) WaitFinished(tb testing.TB, opts ...FinalOpt) {
	tm.waitDone(tb, opts)
}

// FinalModel returns the resulting model, resulting from program.Run().
// This method only returns once the program has finished running or when it
// times out.
func (tm *TestModel[M]) FinalModel(tb testing.TB, opts ...FinalOpt) tea.Model[M] {
	tm.waitDone(tb, opts)
	select {
	case m := <-tm.modelCh:
		tm.model = m
		return tm.model
	default:
		return tm.model
	}
}

// FinalOutput returns the program's final output io.Reader.
// This method only returns once the program has finished running or when it
// times out.
func (tm *TestModel[M]) FinalOutput(tb testing.TB, opts ...FinalOpt) io.Reader {
	tm.waitDone(tb, opts)
	return tm.Output()
}

// Output returns the program's current output io.Reader.
func (tm *TestModel[M]) Output() io.Reader {
	return tm.out
}

// Send sends messages to the underlying program.
func (tm *TestModel[M]) Send(m tea.Msg) {
	tm.program.Send(m)
}

// Quit quits the program and releases the terminal.
func (tm *TestModel[M]) Quit() error {
	tm.program.Quit()
	return nil
}

// Type types the given text into the given program.
func (tm *TestModel[M]) Type(s string) {
	for _, c := range []byte(s) {
		tm.program.Send(tea.MsgKey{
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
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil { //nolint: gomnd
			tb.Fatal(err)
		}
		if err := os.WriteFile(golden, out, 0o600); err != nil { //nolint: gomnd
			tb.Fatal(err)
		}
	}

	path := filepath.Join(tb.TempDir(), tb.Name()+".out")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { //nolint: gomnd
		tb.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil { //nolint: gomnd
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

func safe(rw io.ReadWriter) io.ReadWriter {
	return &safeReadWriter{rw: rw}
}

// safeReadWriter implements io.ReadWriter, but locks reads and writes.
type safeReadWriter struct {
	rw io.ReadWriter
	m  sync.RWMutex
}

// Read implements io.ReadWriter.
func (s *safeReadWriter) Read(p []byte) (int, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.rw.Read(p) //nolint: wrapcheck
}

// Write implements io.ReadWriter.
func (s *safeReadWriter) Write(p []byte) (int, error) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.rw.Write(p) //nolint: wrapcheck
}
