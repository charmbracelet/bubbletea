package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// countModel counts the messages it handles and the views it renders.
type countModel struct {
	count    int          // only touched by the event loop
	views    atomic.Int64 // calls to View
	rendered atomic.Int64 // the count the last view showed
	viewedAt atomic.Int64 // when View was last called, in Unix nanoseconds
}

func (m *countModel) Init() Cmd { return nil }

func (m *countModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case incrementMsg, MouseMotionMsg:
		m.count++
	case KeyPressMsg:
		if msg.String() == "q" {
			return m, Quit
		}
	}
	return m, nil
}

func (m *countModel) View() View {
	m.views.Add(1)
	m.rendered.Store(int64(m.count))
	m.viewedAt.Store(time.Now().UnixNano())
	return NewView(fmt.Sprintf("count: %d", m.count))
}

// recordingRenderer remembers the last view it was given and the last one
// it flushed.
type recordingRenderer struct {
	nilRenderer
	mu      sync.Mutex
	view    View
	flushed View
}

func (r *recordingRenderer) render(v View) {
	r.mu.Lock()
	r.view = v
	r.mu.Unlock()
}

func (r *recordingRenderer) flush(bool) error {
	r.mu.Lock()
	r.flushed = r.view
	r.mu.Unlock()
	return nil
}

func (r *recordingRenderer) lastFlushed() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.flushed.Content
}

// runCountModel runs a countModel. The returned function quits the program
// and returns what Run returned.
func runCountModel(t *testing.T, input io.Reader, opts ...ProgramOption) (*Program, *countModel, func() (Model, error)) {
	t.Helper()
	if input == nil {
		input = &bytes.Buffer{}
	}
	m := &countModel{}
	p := NewProgram(m, append([]ProgramOption{
		WithInput(input),
		WithOutput(io.Discard),
		WithoutSignalHandler(),
	}, opts...)...)
	type result struct {
		model Model
		err   error
	}
	done := make(chan result, 1)
	go func() {
		model, err := p.Run()
		done <- result{model, err}
	}()
	quit := func() (Model, error) {
		p.Quit()
		select {
		case r := <-done:
			return r.model, r.err
		case <-time.After(5 * time.Second):
			t.Fatal("program didn't quit")
			return nil, nil
		}
	}
	t.Cleanup(func() { p.Kill() })
	return p, m, quit
}

// waitFor waits until cond is true.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

// waitIdle waits for the startup messages to be rendered and a frame to
// pass without a render, so the next message is an isolated one.
func waitIdle(t *testing.T, m *countModel, frame time.Duration) {
	t.Helper()
	waitFor(t, "the first views", func() bool { return m.views.Load() >= 2 })
	waitFor(t, "a frame without views", func() bool {
		return time.Since(time.Unix(0, m.viewedAt.Load())) > frame+frame/4
	})
}

func TestRenderCoalescesBurst(t *testing.T) {
	t.Parallel()
	const n = 1000
	p, m, quit := runCountModel(t, nil)
	waitIdle(t, m, time.Second/defaultFPS)

	before := m.views.Load()
	for range n {
		p.Send(incrementMsg{})
	}
	waitFor(t, "the last message to be rendered", func() bool { return m.rendered.Load() == n })
	views := m.views.Load() - before
	t.Logf("%d views for a burst of %d messages", views, n)
	if views > n/10 {
		t.Errorf("rendered %d views for a burst of %d messages", views, n)
	}
	if _, err := quit(); err != nil {
		t.Fatal(err)
	}
}

func TestRenderCoalescesInputBurst(t *testing.T) {
	t.Parallel()
	const n = 1000
	r, w := io.Pipe()
	defer r.Close() //nolint:errcheck
	_, m, quit := runCountModel(t, r)
	waitIdle(t, m, time.Second/defaultFPS)

	// Mouse drag motions, written at once like a terminal does when the
	// program can't keep up.
	var seqs strings.Builder
	for i := range n {
		fmt.Fprintf(&seqs, "\x1b[<32;%d;%dM", 1+i%80, 1+i%24)
	}
	before := m.views.Load()
	if _, err := io.WriteString(w, seqs.String()); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the last motion to be rendered", func() bool { return m.rendered.Load() == n })
	views := m.views.Load() - before
	t.Logf("%d views for a burst of %d motions", views, n)
	if views > n/10 {
		t.Errorf("rendered %d views for a burst of %d motions", views, n)
	}
	if _, err := quit(); err != nil {
		t.Fatal(err)
	}
}

func TestRenderIsolatedMessagePromptly(t *testing.T) {
	t.Parallel()
	const fps = 5
	frame := time.Second / fps
	p, m, quit := runCountModel(t, nil, WithFPS(fps))
	waitIdle(t, m, frame)

	start := time.Now()
	p.Send(incrementMsg{})
	waitFor(t, "the message to be rendered", func() bool { return m.rendered.Load() == 1 })
	// A held back render would wait for the frame to pass.
	if took := time.Since(start); took > frame/2 {
		t.Errorf("an isolated message was rendered after %v, want right away", took)
	}
	if _, err := quit(); err != nil {
		t.Fatal(err)
	}
}

func TestRenderContinuousStream(t *testing.T) {
	t.Parallel()
	const fps = 20
	frame := time.Second / fps
	p, m, quit := runCountModel(t, nil, WithFPS(fps))
	waitIdle(t, m, frame)

	// Send messages as fast as they are taken for a while: the view must
	// still be rendered about once a frame, not after every message and not
	// only at the end.
	const streamFor = 500 * time.Millisecond
	before := m.views.Load()
	sent := 0
	var shown []int64
	for start := time.Now(); time.Since(start) < streamFor; {
		p.Send(incrementMsg{})
		sent++
		if r := m.rendered.Load(); len(shown) == 0 || shown[len(shown)-1] != r {
			shown = append(shown, r)
		}
	}
	views := m.views.Load() - before
	frames := int64(streamFor / frame)
	t.Logf("%d views and %d distinct states for %d messages in %d frames", views, len(shown), sent, frames)
	if views < frames/2 || views > 2*frames+2 {
		t.Errorf("rendered %d views for %d messages in %d frames", views, sent, frames)
	}
	if len(shown) < int(frames/2) {
		t.Errorf("only %d distinct states were rendered while messages kept coming", len(shown))
	}
	waitFor(t, "the last message to be rendered", func() bool { return m.rendered.Load() == int64(sent) })
	if _, err := quit(); err != nil {
		t.Fatal(err)
	}
}

func TestRenderQuitAfterBurst(t *testing.T) {
	t.Parallel()
	const n = 500
	rr := &recordingRenderer{}
	m := &countModel{}
	p := NewProgram(m,
		WithInput(&bytes.Buffer{}),
		WithOutput(io.Discard),
		WithoutSignalHandler(),
	)
	p.renderer = rr
	go func() {
		for range n {
			p.Send(incrementMsg{})
		}
		p.Quit()
	}()
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if got := m.rendered.Load(); got != n {
		t.Errorf("the last view showed %d, want %d", got, n)
	}
	if got, want := rr.lastFlushed(), fmt.Sprintf("count: %d", n); got != want {
		t.Errorf("the last frame drawn was %q, want %q", got, want)
	}
	if views := m.views.Load(); views > n/10 {
		t.Errorf("rendered %d views for a burst of %d messages", views, n)
	}
}

// TestRenderBeforeExec checks that the terminal is handed over with the
// latest view drawn, even when its render was held back.
func TestRenderBeforeExec(t *testing.T) {
	t.Parallel()
	rr := &recordingRenderer{}
	m := &countModel{}
	p := NewProgram(m,
		WithInput(&bytes.Buffer{}),
		WithOutput(io.Discard),
		WithoutSignalHandler(),
		WithFPS(1), // so renders right after the first ones are held back
	)
	p.renderer = rr
	var atExec string
	go func() {
		for m.views.Load() < 2 {
			time.Sleep(time.Millisecond)
		}
		for range 10 {
			p.Send(incrementMsg{})
		}
		p.Send(execMsg{cmd: &testExecCommand{run: func() { atExec = rr.lastFlushed() }}})
		p.Quit()
	}()
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if want := "count: 10"; atExec != want {
		t.Errorf("the frame drawn when the terminal was handed over was %q, want %q", atExec, want)
	}
}

// testExecCommand is an ExecCommand that calls run.
type testExecCommand struct{ run func() }

func (c *testExecCommand) Run() error          { c.run(); return nil }
func (c *testExecCommand) SetStdin(io.Reader)  {}
func (c *testExecCommand) SetStdout(io.Writer) {}
func (c *testExecCommand) SetStderr(io.Writer) {}
