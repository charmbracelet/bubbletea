package tea

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

// lockedBuf records what the renderer emitted, safely across goroutines.
type lockedBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (w *lockedBuf) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

func (w *lockedBuf) len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Len()
}

func (w *lockedBuf) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

type idleModel struct{ n int }

type idleBumpMsg struct{}

func (m idleModel) Init() Cmd { return nil }

func (m idleModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(idleBumpMsg); ok {
		m.n++
	}
	return m, nil
}

// Each frame differs in content and length, so the cell diff can never be empty.
func (m idleModel) View() View {
	var b strings.Builder
	for i := 0; i <= m.n; i++ {
		b.WriteString("row")
		b.WriteRune(rune('A' + i))
		b.WriteByte('\n')
	}
	return NewView(b.String())
}

func idleProgram(t *testing.T, w *lockedBuf, fps int) (*Program, chan struct{}) {
	t.Helper()
	var in bytes.Buffer
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)
	p := NewProgram(idleModel{}, WithContext(ctx), WithInput(&in),
		WithOutput(w), WithFPS(fps), WithWindowSize(80, 24))
	done := make(chan struct{})
	go func() { defer close(done); _, _ = p.Run() }()
	return p, done
}

// The renderer's ticker stops when nothing is pending. This proves it comes back:
// every update lands long after the ticker has gone idle, and each one still has to
// reach the output.
func TestRendererWakesAfterGoingIdle(t *testing.T) {
	t.Parallel()

	w := &lockedBuf{}
	p, done := idleProgram(t, w, 30)

	time.Sleep(300 * time.Millisecond) // first frame drawn, ticker now idle
	sizes := []int{w.len()}
	for range 4 {
		p.Send(idleBumpMsg{})
		time.Sleep(300 * time.Millisecond) // far longer than one 33ms interval
		sizes = append(sizes, w.len())
	}
	p.Quit()
	<-done

	if sizes[0] == 0 {
		t.Fatal("the first frame was never drawn")
	}
	for i := 1; i < len(sizes); i++ {
		if sizes[i] <= sizes[i-1] {
			t.Errorf("update %d produced no output (%d then %d bytes): the ticker did not re-arm",
				i, sizes[i-1], sizes[i])
		}
	}
	for _, want := range []string{"rowA", "rowB", "rowC", "rowD", "rowE"} {
		if !strings.Contains(w.String(), want) {
			t.Errorf("frame %q was never drawn", want)
		}
	}
}

// fps bounds how long an update waits to be flushed, so at a low rate the leading
// edge is what keeps the program responsive: the first frame after an idle stretch
// must not wait out an interval.
func TestLowFPSStaysResponsiveAfterIdle(t *testing.T) {
	t.Parallel()

	w := &lockedBuf{}
	p, done := idleProgram(t, w, 1)

	// Two intervals, not one: the ticker stops on the first tick that finds nothing
	// pending, so reaching the idle state costs one tick to draw and one to notice.
	time.Sleep(2500 * time.Millisecond)
	before := w.len()
	start := time.Now()
	p.Send(idleBumpMsg{})
	for time.Since(start) < 3*time.Second {
		if w.len() > before {
			break
		}
		time.Sleep(time.Millisecond)
	}
	took := time.Since(start)
	p.Quit()
	<-done

	if w.len() <= before {
		t.Fatal("the update never reached the output")
	}
	// Without the leading edge this is a full second at fps=1.
	if took > 200*time.Millisecond {
		t.Errorf("took %v to draw after idle at fps=1, want well under an interval", took)
	}
}
