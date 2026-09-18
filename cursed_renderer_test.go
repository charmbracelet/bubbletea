package tea

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
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

var kittyStackSeqRe = regexp.MustCompile(`\x1b\[(\?1049[hl]|>[0-9;]*u|<[0-9]*u)`)

// countKittyPushes counts the Kitty keyboard push sequences in a terminal
// output stream, regardless of the flags pushed.
func countKittyPushes(out string) (n int) {
	for _, m := range kittyStackSeqRe.FindAllStringSubmatch(out, -1) {
		if m[1][0] == '>' {
			n++
		}
	}
	return n
}

// simulateKittyStacks replays the Kitty keyboard push/pop and alt screen
// switch sequences of a terminal output stream and returns the final depth of
// the main and alt screen stacks. A pop on an empty stack fails the test: a
// real terminal would remove an entry pushed by the host application.
func simulateKittyStacks(t *testing.T, out string) (mainDepth, altDepth int) {
	t.Helper()
	depth := make(map[bool]int)
	var alt bool
	for _, m := range kittyStackSeqRe.FindAllStringSubmatch(out, -1) {
		seq := m[1]
		switch {
		case seq == "?1049h":
			alt = true
		case seq == "?1049l":
			alt = false
		case seq[0] == '>':
			depth[alt]++
		case seq[0] == '<':
			n := 1
			if digits := seq[1 : len(seq)-1]; digits != "" {
				var err error
				if n, err = strconv.Atoi(digits); err != nil {
					t.Fatalf("bad pop sequence %q in %q", seq, out)
				}
			}
			depth[alt] -= n
			if depth[alt] < 0 {
				t.Errorf("pop of %d entries on a stack of %d (alt=%v) in %q", n, depth[alt]+n, alt, out)
				depth[alt] = 0
			}
		}
	}
	return depth[false], depth[true]
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
	// close pops the stack entry; the first flush after the restart pushes a
	// fresh one.
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	r.start()

	// The first flush after the restart enters the alt screen, then a later
	// flush leaves it. The terminal keeps a separate Kitty keyboard stack per
	// screen: nothing was pushed since the restart, so entering the alt
	// screen must not pop, and each screen activation gets its own push.
	view.AltScreen = true
	render(view)
	view.AltScreen = false
	render(view)

	if err := r.close(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	// One push per screen activation: the first flush, the alt screen
	// entered right after the restart, and the main screen on the way back.
	if n := strings.Count(got, pushMain); n != 3 {
		t.Fatalf("expected kitty keyboard protocol to be pushed 3 times with %q (%d times), got %q", pushMain, n, got)
	}
	// One pop per entry: the first close, leaving the alt screen, and the
	// final close.
	if n := strings.Count(got, pop); n != 3 {
		t.Fatalf("expected kitty keyboard protocol to be popped 3 times with %q (%d times), got %q", pop, n, got)
	}
	// Pushes and pops must alternate so that every pop removes the entry we
	// pushed, never one owned by a host application.
	assertInOrder(t, got, pushMain, pop, pushMain, pop, pushMain, pop)
	// Replaying the stream must leave both screens' stacks empty, with no
	// pop ever hitting an empty stack.
	if mainDepth, altDepth := simulateKittyStacks(t, got); mainDepth != 0 || altDepth != 0 {
		t.Fatalf("expected balanced kitty keyboard stacks, got main=%d alt=%d in %q", mainDepth, altDepth, got)
	}
	if strings.Contains(got, ansi.KittyKeyboard(0, 1)) {
		t.Fatalf("expected kitty keyboard protocol not to be reset in-place with %q, got %q", ansi.KittyKeyboard(0, 1), got)
	}
}

func TestCursedRenderer_resumePushesKittyKeyboardAgain(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.start()

	view := NewView("hello")
	view.KeyboardEnhancements.ReportEventTypes = true

	render := func(v View) {
		t.Helper()
		r.render(v)
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}

	render(view)
	if err := r.close(); err != nil {
		t.Fatal(err)
	}

	// Resume with an unchanged view: close popped our entry, so the first
	// flush after the restart must push a fresh one for the final close to
	// pop.
	r.start()
	render(view)
	if err := r.close(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	push := ansi.PushKittyKeyboard(keyboardEnhancementsFlags(view.KeyboardEnhancements))
	pop := ansi.PopKittyKeyboard(1)
	assertInOrder(t, got, push, pop, push, pop)
	if n := countKittyPushes(got); n != 2 {
		t.Fatalf("expected 2 kitty keyboard pushes, got %d in %q", n, got)
	}
	if mainDepth, altDepth := simulateKittyStacks(t, got); mainDepth != 0 || altDepth != 0 {
		t.Fatalf("expected balanced kitty keyboard stacks, got main=%d alt=%d in %q", mainDepth, altDepth, got)
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
	if n := countKittyPushes(got); n != 1 {
		t.Fatalf("expected kitty keyboard protocol to be pushed once, got %d pushes in %q", n, got)
	}
}
