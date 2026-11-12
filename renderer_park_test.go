package tea

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// These tests cover the renderer stopping its ticker while nothing changes.
// Once the ticker is stopped, a change only reaches the output if whatever
// made it woke the renderer, so each test parks the renderer first and then
// checks that the change is drawn.

// safeBuffer is an output that can be read while the program writes to it.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p) //nolint:wrapcheck
}

func (b *safeBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

// since returns everything written after the first n bytes.
func (b *safeBuffer) since(n int) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()[n:]
}

// tickRenderer records when the renderer goroutine flushes, which happens
// once per tick of the frame ticker.
type tickRenderer struct {
	renderer

	mu    sync.Mutex
	ticks int
}

func (r *tickRenderer) flush(closing bool) (bool, error) {
	if !closing {
		r.mu.Lock()
		r.ticks++
		r.mu.Unlock()
	}
	return r.renderer.flush(closing)
}

// count returns the number of ticks so far.
func (r *tickRenderer) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ticks
}

type setViewMsg func(*View)

type runCmdMsg struct{ cmd Cmd }

type parkModel struct{ view View }

func (m *parkModel) Init() Cmd { return nil }

func (m *parkModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case setViewMsg:
		msg(&m.view)
	case runCmdMsg:
		return m, msg.cmd
	}
	return m, nil
}

func (m *parkModel) View() View { return m.view }

type parkHarness struct {
	t     *testing.T
	p     *Program
	out   *safeBuffer
	r     *tickRenderer
	frame time.Duration
	done  chan struct{}
	err   error // valid once done is closed
}

func newParkHarness(t *testing.T, fps int, opts ...ProgramOption) *parkHarness {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	h := &parkHarness{
		t:     t,
		out:   &safeBuffer{},
		frame: time.Second / time.Duration(fps),
		done:  make(chan struct{}),
	}
	h.r = &tickRenderer{renderer: newCursedRenderer(h.out, nil, 80, 24)}
	opts = append([]ProgramOption{
		WithContext(ctx),
		WithInput(&bytes.Buffer{}),
		WithOutput(h.out),
		WithFPS(fps),
		WithWindowSize(80, 24),
	}, opts...)
	h.p = NewProgram(&parkModel{view: NewView("first")}, opts...)
	h.p.renderer = h.r

	go func() {
		defer close(h.done)
		_, h.err = h.p.Run()
	}()
	t.Cleanup(func() {
		h.p.Kill()
		<-h.done
	})
	return h
}

// waitParked waits until the renderer has stopped ticking.
func (h *parkHarness) waitParked() {
	h.t.Helper()

	// Parked means no tick for well over a frame. The ticker runs for about
	// half a second of unchanged frames before it stops.
	quiet := 6 * h.frame
	deadline := time.Now().Add(10 * time.Second)
	n := h.r.count()
	for time.Now().Before(deadline) {
		time.Sleep(quiet)
		m := h.r.count()
		if m == n && n > 0 {
			return
		}
		n = m
	}
	h.t.Fatal("the renderer never stopped ticking")
}

// waitOutput waits until everything in want is written after mark and
// returns how long that took.
func (h *parkHarness) waitOutput(mark int, within time.Duration, want ...string) time.Duration {
	h.t.Helper()

	start := time.Now()
	for {
		got := h.out.since(mark)
		missing := ""
		for _, w := range want {
			if !strings.Contains(got, w) {
				missing = w
				break
			}
		}
		if missing == "" {
			return time.Since(start)
		}
		if time.Since(start) > within {
			h.t.Fatalf("%q not written within %v after waking a parked renderer; got %q", missing, within, got)
		}
		time.Sleep(time.Millisecond)
	}
}

func (h *parkHarness) setView(fn func(*View)) {
	h.p.Send(setViewMsg(fn))
}

// waitDone waits for Run to return.
func (h *parkHarness) waitDone() error {
	h.t.Helper()

	const within = 2 * time.Second
	select {
	case <-h.done:
		return h.err
	case <-time.After(within):
		h.t.Fatalf("Run did not return within %v", within)
		return nil
	}
}

// nopExecCommand is an ExecCommand that does nothing.
type nopExecCommand struct{}

func (nopExecCommand) Run() error          { return nil }
func (nopExecCommand) SetStdin(io.Reader)  {}
func (nopExecCommand) SetStdout(io.Writer) {}
func (nopExecCommand) SetStderr(io.Writer) {}

func TestRendererStopsTickingWhenIdle(t *testing.T) {
	t.Parallel()

	h := newParkHarness(t, 60)
	h.waitOutput(0, 5*time.Second, "first")
	h.waitParked()

	n := h.r.count()
	time.Sleep(time.Second)
	if ticks := h.r.count() - n; ticks != 0 {
		t.Errorf("an idle program ticked %d times in a second, want 0", ticks)
	}
}

// Every way of giving the renderer work must wake it.
func TestRendererWakes(t *testing.T) {
	t.Parallel()

	cmd := func(c Cmd) func(h *parkHarness) {
		return func(h *parkHarness) { h.p.Send(runCmdMsg{c}) }
	}

	tests := []struct {
		name string
		do   func(h *parkHarness)
		want []string
	}{
		{"content", func(h *parkHarness) { h.setView(func(v *View) { v.Content = "second" }) }, []string{"second"}},
		{"Println", cmd(Println("printed")), []string{"printed"}},
		{"raw", cmd(Raw("\x1b]2;raw\x07")), []string{"\x1b]2;raw\x07"}},
		{"exec", cmd(Exec(nopExecCommand{}, nil)), []string{ansi.SetModeBracketedPaste, "first"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newParkHarness(t, 60)
			h.waitOutput(0, 5*time.Second, "first")
			h.waitParked()

			mark := h.out.Len()
			tc.do(h)
			took := h.waitOutput(mark, 2*time.Second, tc.want...)
			t.Logf("drawn %v after waking", took)
		})
	}
}

func TestRendererQuitWhileParked(t *testing.T) {
	t.Parallel()

	h := newParkHarness(t, 60)
	h.waitParked()

	mark := h.out.Len()
	h.setView(func(v *View) { v.Content = "last" })
	h.p.Quit()
	if err := h.waitDone(); err != nil {
		t.Errorf("Run returned %v, want nil", err)
	}
	if !strings.Contains(h.out.since(mark), "last") {
		t.Error("final frame not drawn")
	}
}
