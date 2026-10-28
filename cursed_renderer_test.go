package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
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

func assertInOrder(t *testing.T, got string, wants ...string) {
	t.Helper()
	rest := got
	for _, want := range wants {
		idx := strings.Index(rest, want)
		if idx < 0 {
			t.Fatalf("expected %q to appear after the previous sequences in %q", want, got)
		}
		rest = rest[idx+len(want):]
	}
}

func TestCursedRenderer_restoresKittyKeyboardStack(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.start()

	view := NewView("hello")
	view.KeyboardEnhancements.ReportEventTypes = true
	pushMain := ansi.PushKittyKeyboard(keyboardEnhancementsFlags(view.KeyboardEnhancements))
	pop := ansi.PopKittyKeyboard(1)

	render := func(v View) {
		t.Helper()
		r.render(v)
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}

	render(view)

	// Stop the renderer (as on suspend or ExecProcess) and start it again:
	// close pops the stack entry, start pushes it back.
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	r.start()

	// Enter and leave the alt screen. The terminal keeps a separate Kitty
	// keyboard stack per screen, so each screen gets its own push and pop.
	view.AltScreen = true
	render(view)
	view.AltScreen = false
	render(view)

	if err := r.close(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	// The flags are pushed once per screen activation: the first flush,
	// the flush after the renderer was restarted, and on each screen
	// switch. start() itself does not write to [out].
	if n := strings.Count(got, pushMain); n != 4 {
		t.Fatalf("expected kitty keyboard protocol to be pushed 4 times with %q (%d times), got %q", pushMain, n, got)
	}
	// One pop per stop/start cycle and per screen switch: closing pops the
	// current screen's entry, and switching screens pops the entry of the
	// screen being left.
	if n := strings.Count(got, pop); n != 4 {
		t.Fatalf("expected kitty keyboard protocol to be popped 4 times with %q (%d times), got %q", pop, n, got)
	}
	// Every pop must come after a push: the stack is balanced when pushes
	// and pops alternate. The resumed flush pushes twice in a row (once in
	// start(), once in flush()), and both entries are popped afterwards.
	assertInOrder(t, got,
		pushMain, pop, // close pops the entry pushed by the first flush
		pop,           // entering the alt screen pops the resumed entry
		pushMain, pop, // leaving the alt screen
		pushMain, pop, // the resumed main screen entry and the final close
	)
	if strings.Contains(got, ansi.KittyKeyboard(0, 1)) {
		t.Fatalf("expected kitty keyboard protocol not to be reset in-place with %q, got %q", ansi.KittyKeyboard(0, 1), got)
	}
}

func TestCursedRenderer_updatesKittyKeyboardFlagsInPlace(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)

	render := func(v View) {
		t.Helper()
		r.render(v)
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}

	view := NewView("hello")
	render(view)

	// Changing the enhancement flags without switching screens updates the
	// current stack entry in place instead of pushing a new one.
	changed := view
	changed.KeyboardEnhancements.ReportEventTypes = true
	render(changed)

	wantUpdate := ansi.KittyKeyboard(keyboardEnhancementsFlags(changed.KeyboardEnhancements), 1)
	got := out.String()
	if !strings.Contains(got, wantUpdate) {
		t.Fatalf("expected kitty keyboard flags to be updated in place with %q, got %q", wantUpdate, got)
	}
	assertInOrder(t, got,
		ansi.PushKittyKeyboard(keyboardEnhancementsFlags(view.KeyboardEnhancements)),
		wantUpdate,
	)
	if strings.Contains(got, ansi.PopKittyKeyboard(1)) {
		t.Fatalf("expected kitty keyboard protocol not to be popped with %q, got %q", ansi.PopKittyKeyboard(1), got)
	}
	if n := strings.Count(got, ansi.PushKittyKeyboard(0)); n > 1 {
		t.Fatalf("expected kitty keyboard protocol to be pushed once, got %d pushes in %q", n, got)
	}
}

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1838
func TestCursedRenderer_refreshTabStopsOnResize(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.setOptimizations(true, false, false)
	r.start()

	render := func(v View) {
		t.Helper()
		r.render(v)
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}

	view := NewView("hello")
	render(view)

	// Startup should have emitted SetTabEvery8Columns once.
	if count := strings.Count(out.String(), ansi.SetTabEvery8Columns); count != 1 {
		t.Fatalf("expected 1 tab stops reset at startup, got %d in %q", count, out.String())
	}

	// Another render without resize should not re-emit SetTabEvery8Columns.
	render(NewView("hello world"))
	if count := strings.Count(out.String(), ansi.SetTabEvery8Columns); count != 1 {
		t.Fatalf("expected no additional tab stops reset without resize, got %d", count)
	}

	// Height-only resize should not re-emit SetTabEvery8Columns.
	r.resize(80, 40)
	render(NewView("hello height resize"))
	if count := strings.Count(out.String(), ansi.SetTabEvery8Columns); count != 1 {
		t.Fatalf("expected no tab stops reset on height-only resize, got %d", count)
	}

	// Identical size resize should not re-emit SetTabEvery8Columns.
	r.resize(80, 40)
	render(NewView("hello same size"))
	if count := strings.Count(out.String(), ansi.SetTabEvery8Columns); count != 1 {
		t.Fatalf("expected no tab stops reset on identical resize, got %d", count)
	}

	// Widening resize should re-emit SetTabEvery8Columns.
	r.resize(272, 40)
	render(NewView("hello widened"))
	if count := strings.Count(out.String(), ansi.SetTabEvery8Columns); count != 2 {
		t.Fatalf("expected 2 tab stops resets after widening resize, got %d", count)
	}

	// Narrowing resize should also re-emit SetTabEvery8Columns.
	r.resize(120, 40)
	render(NewView("hello narrowed"))
	if count := strings.Count(out.String(), ansi.SetTabEvery8Columns); count != 3 {
		t.Fatalf("expected 3 tab stops resets after narrowing resize, got %d", count)
	}

	// When hard tabs are disabled, resizing should not emit SetTabEvery8Columns.
	r.setOptimizations(false, false, false)
	r.resize(200, 40)
	render(NewView("hello no hard tabs"))
	if count := strings.Count(out.String(), ansi.SetTabEvery8Columns); count != 3 {
		t.Fatalf("expected no tab stops reset when hard tabs are disabled, got %d", count)
	}

	// Re-enabling hard tabs should reset tab stops on the next flush.
	r.setOptimizations(true, false, false)
	render(NewView("hello hard tabs re-enabled"))
	if count := strings.Count(out.String(), ansi.SetTabEvery8Columns); count != 4 {
		t.Fatalf("expected 4 tab stops resets after re-enabling hard tabs, got %d", count)
	}
}

func TestCursedRenderer_refreshTabStopsOrdering(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.setOptimizations(true, false, false)
	r.start()

	render := func(v View) {
		t.Helper()
		r.render(v)
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}

	render(NewView("initial frame"))
	out.Reset()

	// Widen window and check that SetTabEvery8Columns is output before the rendered content.
	r.resize(272, 24)
	render(NewView("widened content"))

	got := out.String()
	tabIdx := strings.Index(got, ansi.SetTabEvery8Columns)
	contentIdx := strings.Index(got, "widened content")
	if tabIdx < 0 {
		t.Fatalf("expected SetTabEvery8Columns in output, got %q", got)
	}
	if contentIdx < 0 {
		t.Fatalf("expected widened content in output, got %q", got)
	}
	if tabIdx > contentIdx {
		t.Fatalf("expected SetTabEvery8Columns (idx %d) before frame content (idx %d)", tabIdx, contentIdx)
	}
}

