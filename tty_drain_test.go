//go:build !windows
// +build !windows

package tea

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// fakeTerminal answers terminal queries written to it by writing the replies
// to the program's input, like a real terminal would.
type fakeTerminal struct {
	mu    sync.Mutex
	out   bytes.Buffer
	input io.Writer
}

func (f *fakeTerminal) Write(b []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.out.Write(b)
	if bytes.Contains(b, []byte(ansi.RequestModeSynchronizedOutput)) {
		_, _ = io.WriteString(f.input, "\x1b[?2026;2$y")
	}
	if bytes.Contains(b, []byte(ansi.RequestPrimaryDeviceAttributes)) {
		_, _ = io.WriteString(f.input, "\x1b[?62;22c")
	}
	return len(b), nil
}

func (f *fakeTerminal) String() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.out.String()
}

func newDrainTestProgram(t *testing.T) (*Program, *os.File, *fakeTerminal) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})
	ft := &fakeTerminal{input: w}
	p := NewProgram(nil, WithInput(r), WithOutput(ft))
	// Pretend the input is a terminal; drainQueryReplies only runs for TTYs.
	p.ttyInput = r
	return p, w, ft
}

// readAvailable returns what can be read from r within d.
func readAvailable(t *testing.T, r io.Reader, d time.Duration) string {
	t.Helper()
	ch := make(chan string, 1)
	go func() {
		buf := make([]byte, 256)
		n, _ := r.Read(buf)
		ch <- string(buf[:n])
	}()
	select {
	case s := <-ch:
		return s
	case <-time.After(d):
		return ""
	}
}

// Replies to queries that arrive after the input reader is canceled must not
// be left in the tty input for the next reader, such as an exec'd process.
func TestDrainQueryReplies(t *testing.T) {
	p, w, ft := newDrainTestProgram(t)

	// The query is still buffered, as it is when ExecProcess is returned from
	// Init before the renderer has flushed.
	p.executeQuery(ansi.RequestModeSynchronizedOutput)

	start := time.Now()
	p.drainQueryReplies()
	if elapsed := time.Since(start); elapsed >= queryDrainTimeout {
		t.Fatalf("drain did not stop at the DA1 reply, took %v", elapsed)
	}
	if !bytes.Contains([]byte(ft.String()), []byte(ansi.RequestPrimaryDeviceAttributes)) {
		t.Fatalf("expected a DA1 request to be written, got %q", ft.String())
	}

	// The reply is delivered to the program, the DA1 reply is not.
	select {
	case msg := <-p.msgs:
		if ev, ok := msg.(uv.ModeReportEvent); !ok || ev.Mode != ansi.ModeSynchronizedOutput {
			t.Fatalf("expected a synchronized output mode report, got %#v", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("mode report was not delivered to the program")
	}

	// Nothing but what is typed afterwards is left for the next reader.
	if _, err := io.WriteString(w, "abc"); err != nil {
		t.Fatal(err)
	}
	if got := readAvailable(t, p.input, time.Second); got != "abc" {
		t.Fatalf("expected next reader to get %q, got %q", "abc", got)
	}
}

// Without outstanding queries there is nothing to drain, and no DA1 request
// is sent.
func TestDrainQueryRepliesNoPendingQuery(t *testing.T) {
	p, _, ft := newDrainTestProgram(t)

	p.drainQueryReplies()
	if out := ft.String(); out != "" {
		t.Fatalf("expected no output, got %q", out)
	}
}
