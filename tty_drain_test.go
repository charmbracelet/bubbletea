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
	// Pretend input and output are a terminal; drainQueryReplies only runs
	// for TTYs.
	p.ttyInput = r
	p.ttyOutput = w
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
	events := p.drainQueryReplies(t.Context())
	if elapsed := time.Since(start); elapsed >= queryDrainTimeout {
		t.Fatalf("drain did not stop at the DA1 reply, took %v", elapsed)
	}
	if !bytes.Contains([]byte(ft.String()), []byte(ansi.RequestPrimaryDeviceAttributes)) {
		t.Fatalf("expected a DA1 request to be written, got %q", ft.String())
	}

	// The reply is returned, the DA1 reply is not.
	if len(events) != 1 {
		t.Fatalf("expected one reply, got %#v", events)
	}
	if ev, ok := events[0].(uv.ModeReportEvent); !ok || ev.Mode != ansi.ModeSynchronizedOutput {
		t.Fatalf("expected a synchronized output mode report, got %#v", events[0])
	}

	// Nothing but what is typed afterwards is left for the next reader.
	if _, err := io.WriteString(w, "abc"); err != nil {
		t.Fatal(err)
	}
	if got := readAvailable(t, p.input, time.Second); got != "abc" {
		t.Fatalf("expected next reader to get %q, got %q", "abc", got)
	}
}

// Releasing the terminal hands the replies it read to the program.
func TestReleaseTerminalDeliversQueryReplies(t *testing.T) {
	p, _, _ := newDrainTestProgram(t)

	p.executeQuery(ansi.RequestModeSynchronizedOutput)
	if err := p.releaseTerminal(false); err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-p.msgs:
		if ev, ok := msg.(uv.ModeReportEvent); !ok || ev.Mode != ansi.ModeSynchronizedOutput {
			t.Fatalf("expected a synchronized output mode report, got %#v", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("mode report was not delivered to the program")
	}
}

// A program that quits right after it starts sends its queries during
// shutdown. Their replies must be read before the terminal is handed back to
// the shell, or the shell reads them as typed input.
//
// Fixes: https://github.com/charmbracelet/bubbletea/issues/1590
func TestShutdownDrainsQueryReplies(t *testing.T) {
	p, w, ft := newDrainTestProgram(t)

	p.executeQuery(ansi.RequestModeSynchronizedOutput)
	p.shutdown(false)

	if !bytes.Contains([]byte(ft.String()), []byte(ansi.RequestPrimaryDeviceAttributes)) {
		t.Fatalf("expected a DA1 request to be written, got %q", ft.String())
	}
	if _, err := io.WriteString(w, "abc"); err != nil {
		t.Fatal(err)
	}
	if got := readAvailable(t, p.input, time.Second); got != "abc" {
		t.Fatalf("expected the shell to get %q, got %q", "abc", got)
	}
}

// Without outstanding queries there is nothing to drain, and no DA1 request
// is sent.
func TestDrainQueryRepliesNoPendingQuery(t *testing.T) {
	p, _, ft := newDrainTestProgram(t)

	p.drainQueryReplies(t.Context())
	if out := ft.String(); out != "" {
		t.Fatalf("expected no output, got %q", out)
	}
}

// The reply to a query sent long ago has been read by the input reader
// already, so there's nothing to drain.
func TestDrainQueryRepliesOldQuery(t *testing.T) {
	p, _, ft := newDrainTestProgram(t)

	p.lastQuery.Store(time.Now().Add(-2 * queryReplyWindow).UnixNano())
	p.drainQueryReplies(t.Context())
	if out := ft.String(); out != "" {
		t.Fatalf("expected no output, got %q", out)
	}
}

// The renderer's keyboard enhancements query counts as a query.
func TestCursedRendererReportsQueries(t *testing.T) {
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	var queries int
	r.onQuery = func() { queries++ }

	r.render(NewView("hello"))
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte(ansi.RequestKittyKeyboard)) {
		t.Fatalf("expected a keyboard enhancements query, got %q", out.String())
	}
	if queries != 1 {
		t.Fatalf("expected the query to be reported once, got %d", queries)
	}
}
