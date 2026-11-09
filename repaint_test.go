package tea

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *syncBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

func newTestCursedRenderer(w *bytes.Buffer, width, height int) *cursedRenderer {
	r := newCursedRenderer(w, []string{"TERM=xterm-256color"}, width, height)
	r.setColorProfile(colorprofile.ANSI256)
	return r
}

func altView(content string) View {
	v := NewView(content)
	v.AltScreen = true
	return v
}

func paint(t *testing.T, r *cursedRenderer, view View) {
	t.Helper()
	r.render(view)
	if err := r.flush(false); err != nil {
		t.Fatalf("flush: %v", err)
	}
}

// TestClearScreenSkipsUnchangedAltScreen is the neutralized-fix baseline:
// after an external alt-screen exit, ClearScreen plus an identical view
// must not re-enter the alternate screen. If this starts passing for
// 1049h, Repaint is no longer the distinct recovery API.
func TestClearScreenSkipsUnchangedAltScreen(t *testing.T) {
	var buf bytes.Buffer
	r := newTestCursedRenderer(&buf, 40, 8)
	r.start()
	paint(t, r, altView("hello"))
	if !strings.Contains(buf.String(), ansi.SetModeAltScreenSaveCursor) {
		t.Fatalf("initial paint missing alt-screen enter: %q", buf.String())
	}

	buf.Reset()
	paint(t, r, altView("hello"))
	if buf.Len() != 0 {
		t.Fatalf("unchanged flush wrote %q", buf.String())
	}

	buf.Reset()
	r.clearScreen()
	paint(t, r, altView("hello"))
	out := buf.String()
	if strings.Contains(out, ansi.SetModeAltScreenSaveCursor) {
		t.Fatalf("ClearScreen re-entered alt screen for an unchanged view: %q", out)
	}
}

func TestRepaintReentersAltScreenForUnchangedView(t *testing.T) {
	var buf bytes.Buffer
	r := newTestCursedRenderer(&buf, 40, 8)
	r.start()
	paint(t, r, altView("hello"))

	buf.Reset()
	paint(t, r, altView("hello"))
	if buf.Len() != 0 {
		t.Fatalf("unchanged flush wrote %q", buf.String())
	}

	buf.Reset()
	r.repaint()
	paint(t, r, altView("hello"))
	out := buf.String()
	if !strings.Contains(out, ansi.SetModeAltScreenSaveCursor) {
		t.Fatalf("Repaint did not re-enter alt screen: %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("Repaint did not redraw the unchanged view: %q", out)
	}
}

func TestRepaintNilLastViewDoesNotPanic(t *testing.T) {
	var buf bytes.Buffer
	r := newTestCursedRenderer(&buf, 40, 8)
	r.repaint()
	paint(t, r, altView("hello"))
	if !strings.Contains(buf.String(), "hello") {
		t.Fatalf("first paint after nil-lastView Repaint missing content: %q", buf.String())
	}
}

func waitForOutput(t *testing.T, buf *syncBuffer, needle string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), needle) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %q in output %q", needle, buf.String())
}

func TestRepaintMsgIsHandledWithoutQuit(t *testing.T) {
	// Send+Quit cannot prove recovery: flush(closing=true) bypasses the
	// unchanged-view skip. Drive a live ticker flush instead.
	var buf syncBuffer
	var in bytes.Buffer
	m := &testViewModel{testModel: &testModel{}, opts: testViewOpts{altScreen: true}}
	p := NewProgram(m,
		WithWindowSize(40, 8),
		WithColorProfile(colorprofile.ANSI256),
		WithEnvironment([]string{"TERM=xterm-256color"}),
		WithInput(&in),
		WithOutput(&buf),
		WithFPS(60),
	)

	done := make(chan error, 1)
	go func() { _, err := p.Run(); done <- err }()
	t.Cleanup(func() {
		p.Quit()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	})

	waitForOutput(t, &buf, "success", time.Second)

	before := buf.Len()
	p.Send(ClearScreen())
	time.Sleep(80 * time.Millisecond)
	clearDelta := buf.String()[before:]
	if strings.Contains(clearDelta, ansi.SetModeAltScreenSaveCursor) {
		t.Fatalf("live ClearScreen re-entered alt screen: %q", clearDelta)
	}

	before = buf.Len()
	p.Send(Repaint())
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		out := buf.String()
		if before < len(out) && strings.Contains(out[before:], ansi.SetModeAltScreenSaveCursor) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("live Repaint missing alt-screen enter; delta=%q full=%q", buf.String()[before:], buf.String())
}
