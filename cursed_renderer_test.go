package tea

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"
)

type mouseRaceModel struct {
	i int
}

func (m *mouseRaceModel) Init() Cmd { return nil }

func (m *mouseRaceModel) Update(msg Msg) (Model, Cmd) {
	switch msg.(type) {
	case MouseClickMsg, MouseMotionMsg, MouseWheelMsg:
		m.i++
	}
	return m, nil
}

func (m *mouseRaceModel) View() View {
	return View{
		Content:   fmt.Sprintf("tick-%d\n", m.i),
		MouseMode: MouseModeCellMotion,
	}
}

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1690
func TestCursedRenderer_mouseVsFlush(t *testing.T) {
	t.Parallel()

	pr, pw := io.Pipe()
	defer func() { _ = pw.Close() }()

	m := &mouseRaceModel{}
	p := NewProgram(
		m,
		WithContext(t.Context()),
		WithInput(pr),
		WithOutput(io.Discard),
		WithEnvironment([]string{
			"TERM=xterm-256color",
			"TERM_PROGRAM=Apple_Terminal",
		}),
		WithoutSignals(),
		WithWindowSize(80, 24),
	)

	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		_, _ = p.Run()
	}()

	time.Sleep(150 * time.Millisecond)

	const iterations = 100
	for i := range iterations {
		switch i % 4 {
		case 0:
			p.Send(MouseClickMsg{X: i % 80, Y: i % 24, Button: MouseLeft})
		case 1:
			p.Send(MouseMotionMsg{X: i % 80, Y: i % 24})
		case 2:
			p.Send(MouseWheelMsg{X: 0, Y: 0, Button: MouseWheelUp})
		default:
			p.Send(MouseReleaseMsg{X: i % 80, Y: i % 24, Button: MouseLeft})
		}
	}

	p.Quit()
	select {
	case <-runDone:
	case <-time.After(5 * time.Second):
		t.Fatal("program did not exit after Quit")
	}
}

// cursorDownRe matches ANSI cursor-down sequences (CSI n B).
var cursorDownRe = regexp.MustCompile(`\x1b\[(\d+)B`)

// maxCursorDown returns the largest cursor-down (CSI n B) distance found in s,
// or 0 if there are none.
func maxCursorDown(s string) int {
	max := 0
	for _, m := range cursorDownRe.FindAllStringSubmatch(s, -1) {
		if n, err := strconv.Atoi(m[1]); err == nil && n > max {
			max = n
		}
	}
	return max
}

// TestCursedRenderer_insertAboveBeforeFlush reproduces
// https://github.com/charmbracelet/bubbletea/issues/1740.
//
// A tea.Println/tea.Printf issued before the renderer's first flush (e.g. from
// Model.Init) went through insertAbove, which computed the cursor-down distance
// from the cellbuf height. Since the cellbuf is created at the full terminal
// height and only resized to the frame height during flush, the cursor was
// moved down a full screen and the whole viewport scrolled. This drives the
// renderer directly to keep the assertion deterministic (no ticker race).
func TestCursedRenderer_insertAboveBeforeFlush(t *testing.T) {
	t.Parallel()

	const width, height = 80, 24
	var buf bytes.Buffer
	r := newCursedRenderer(&buf, []string{"TERM=xterm-256color"}, width, height)

	// Program.Run stashes the initial view via render() before the event loop
	// starts, but the first flush only happens later on the render ticker.
	r.render(NewView("hello world"))

	// A tea.Println that runs before that first flush.
	if err := r.insertAbove("printed from Init"); err != nil {
		t.Fatalf("insertAbove returned error: %v", err)
	}

	out := buf.String()
	if got := maxCursorDown(out); got >= height-1 {
		t.Fatalf("insertAbove before first flush scrolled the viewport: "+
			"emitted cursor-down of %d (frame is 1 line tall), output=%q", got, out)
	}
}

// printInInitModel issues a tea.Println from Init and renders a single line.
type printInInitModel struct{}

func (printInInitModel) Init() Cmd { return Println("printed from Init") }

func (printInInitModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(KeyPressMsg); ok {
		return printInInitModel{}, Quit
	}
	return printInInitModel{}, nil
}

func (printInInitModel) View() View { return NewView("hello world") }

// TestCursedRenderer_printlnFromInit drives a full Program with a fixed window
// size and asserts that a tea.Println from Init does not emit a large cursor
// down (which would scroll the whole terminal). Regression test for
// https://github.com/charmbracelet/bubbletea/issues/1740.
func TestCursedRenderer_printlnFromInit(t *testing.T) {
	t.Parallel()

	const width, height = 80, 24
	out := &safeBuffer{}

	p := NewProgram(
		printInInitModel{},
		WithContext(t.Context()),
		WithInput(&bytes.Buffer{}),
		WithOutput(out),
		WithEnvironment([]string{
			"TERM=xterm-256color",
			"TERM_PROGRAM=Apple_Terminal",
		}),
		WithoutSignals(),
		WithWindowSize(width, height),
	)

	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		_, _ = p.Run()
	}()

	// Give Init's Println time to hit insertAbove and let at least one frame
	// flush.
	time.Sleep(200 * time.Millisecond)

	p.Quit()
	select {
	case <-runDone:
	case <-time.After(5 * time.Second):
		t.Fatal("program did not exit after Quit")
	}

	if got := maxCursorDown(out.String()); got >= height-1 {
		t.Fatalf("Println from Init scrolled the viewport: emitted cursor-down "+
			"of %d (frame is 1 line tall), output=%q", got, out.String())
	}
}

// safeBuffer is a goroutine-safe bytes.Buffer.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
