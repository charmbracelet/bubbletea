package tea

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

type echoModel struct{ typed string }

func (m *echoModel) Init() Cmd { return nil }

func (m *echoModel) Update(msg Msg) (Model, Cmd) {
	if k, ok := msg.(KeyPressMsg); ok {
		m.typed += k.Text
	}
	return m, nil
}

func (m *echoModel) View() View { return NewView("typed: " + m.typed) }

// waitWriter records everything written and wakes waiters on each write.
type waitWriter struct {
	mu   sync.Mutex
	buf  bytes.Buffer
	wake chan struct{}
}

func (w *waitWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	select {
	case w.wake <- struct{}{}:
	default:
	}
	return len(p), nil
}

func (w *waitWriter) waitFor(t *testing.T, s string, d time.Duration) time.Duration {
	t.Helper()
	t0 := time.Now()
	deadline := time.After(d)
	for {
		w.mu.Lock()
		ok := strings.Contains(w.buf.String(), s)
		w.mu.Unlock()
		if ok {
			return time.Since(t0)
		}
		select {
		case <-w.wake:
		case <-deadline:
			t.Fatalf("%q not written within %v", s, d)
		}
	}
}

// TestInputRendersBeforeNextFrame: a key's effect reaches the terminal right
// after Update and View, without waiting for the next frame tick. At 1 fps a
// key typed just after a frame would otherwise wait about a second.
func TestInputRendersBeforeNextFrame(t *testing.T) {
	in, keys := io.Pipe()
	w := &waitWriter{wake: make(chan struct{}, 1)}
	p := NewProgram(&echoModel{}, WithInput(in), WithOutput(w), WithFPS(1),
		WithWindowSize(40, 5), WithoutSignals())
	done := make(chan struct{})
	go func() { _, _ = p.Run(); close(done) }()
	defer func() { p.Quit(); _ = keys.Close(); <-done }()

	w.waitFor(t, "typed:", 5*time.Second) // the first frame, just flushed by a tick
	if _, err := keys.Write([]byte("a")); err != nil {
		t.Fatal(err)
	}
	if d := w.waitFor(t, "typed: a", 5*time.Second); d > 250*time.Millisecond {
		t.Errorf("the key's frame took %v; the next frame tick is a second away", d)
	}
}
