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

// repaintTestRenderer returns a renderer wired to a buffer together with a
// render function that pushes a view and flushes it.
func repaintTestRenderer(t *testing.T) (*cursedRenderer, *bytes.Buffer, func(View)) {
	t.Helper()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	render := func(v View) {
		t.Helper()
		r.render(v)
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}
	return r, &out, render
}

// After external tty damage Repaint must replay the alt screen and repaint
// the whole view on a normal (non-closing) flush.
func TestCursedRenderer_repaintRestoresAltScreen(t *testing.T) {
	t.Parallel()

	r, out, render := repaintTestRenderer(t)

	view := NewView("hello")
	view.AltScreen = true
	render(view)

	out.Reset()
	r.repaint()
	// A normal flush, not a closing one, must not skip the redraw.
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	assertInOrder(t, got, ansi.SetModeAltScreenSaveCursor, "hello")
}

// Baseline verification: ClearScreen alone does not re-enter the alternate
// screen, whereas Repaint is the recovery operation that does.
func TestCursedRenderer_clearScreenDoesNotRestoreAltScreen(t *testing.T) {
	t.Parallel()

	r, out, render := repaintTestRenderer(t)

	view := NewView("hello")
	view.AltScreen = true
	render(view)

	out.Reset()
	r.clearScreen()
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if strings.Contains(got, ansi.SetModeAltScreenSaveCursor) {
		t.Fatalf("expected ClearScreen not to replay alt screen, got %q", got)
	}
}

// Inline mode must keep working: Repaint must not enter the alt screen.
func TestCursedRenderer_repaintInline(t *testing.T) {
	t.Parallel()

	r, out, render := repaintTestRenderer(t)

	view := NewView("hello")
	render(view)

	out.Reset()
	r.repaint()
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if strings.Contains(got, ansi.SetModeAltScreenSaveCursor) {
		t.Fatalf("expected inline mode not to enter the alt screen, got %q", got)
	}
	if !strings.Contains(got, "hello") {
		t.Fatalf("expected the full frame to be repainted, got %q", got)
	}
}

// Repaint must re-emit the modes carried by the view, e.g. mouse tracking,
// bracketed paste, and focus reporting, which an external writer may have reset.
func TestCursedRenderer_repaintRestoresModes(t *testing.T) {
	t.Parallel()

	r, out, render := repaintTestRenderer(t)

	view := NewView("hello")
	view.MouseMode = MouseModeCellMotion
	view.ReportFocus = true
	render(view)

	out.Reset()
	r.repaint()
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{
		ansi.SetModeMouseButtonEvent + ansi.SetModeMouseExtSgr,
		ansi.SetModeBracketedPaste,
		ansi.SetModeFocusEvent,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q to be replayed by Repaint, got %q", want, got)
		}
	}
}

// Calling Repaint repeatedly must update the Kitty keyboard flags in place
// rather than pushing a new stack entry every time.
func TestCursedRenderer_repaintDoesNotGrowKittyStack(t *testing.T) {
	t.Parallel()

	r, out, render := repaintTestRenderer(t)

	view := NewView("hello")
	view.KeyboardEnhancements.ReportEventTypes = true
	flags := keyboardEnhancementsFlags(view.KeyboardEnhancements)
	render(view)

	const repaints = 3
	out.Reset()
	for range repaints {
		r.repaint()
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}

	got := out.String()
	if n := strings.Count(got, ansi.PushKittyKeyboard(flags)); n != 0 {
		t.Fatalf("expected no kitty keyboard pushes, got %d in %q", n, got)
	}
	if n := strings.Count(got, ansi.KittyKeyboard(flags, 1)); n != repaints {
		t.Fatalf("expected kitty keyboard flags to be set in place %d times, got %d in %q", repaints, n, got)
	}
}

// Calling Repaint when lastView is nil must be a no-op and not panic.
func TestCursedRenderer_repaintNilLastView(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.repaint()

	view := NewView("hello")
	r.render(view)
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("expected view to be rendered after repaint, got %q", out.String())
	}
}
