package tea

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
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

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1822
func TestCursedRenderer_insertAboveMultiScreen(t *testing.T) {
	t.Parallel()

	const width, height = 40, 10
	e := vt.NewEmulator(width, height)
	defer e.Close()
	visible := true
	e.SetCallbacks(vt.Callbacks{CursorVisibility: func(v bool) { visible = v }})
	if _, err := e.Write([]byte("SHELL-HISTORY\r\n")); err != nil {
		t.Fatal(err)
	}
	r := newCursedRenderer(e, []string{"TERM=xterm-256color"}, width, height)
	r.setNoInput(true)
	view := NewView("CONTROLS\n> input\nSTATUS")
	view.Cursor = NewCursor(2, 1)
	r.render(view)
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(e.String(), view.Content) {
		t.Fatalf("initial View was not painted: %q", e.String())
	}

	lines := make([]string, 80)
	for i := range lines {
		lines[i] = fmt.Sprintf("ROW-%03d", i)
	}
	// No pending View, resize, subsequent flush, or cleanup may repair the
	// screen before these assertions. This isolates the multi-screen case
	// from inserting before the first flush or after a changed View.
	if err := r.insertAbove(strings.Join(lines, "\n")); err != nil {
		t.Fatal(err)
	}
	assertTerminalLines(t, e, append([]string{"SHELL-HISTORY"}, lines...))
	if e.ScrollbackLen() == 0 || e.IsAltScreen() {
		t.Error("insertion did not preserve native scrollback")
	}
	screen := e.String()
	if !strings.Contains(screen, view.Content) {
		t.Errorf("managed View missing immediately after insertion: %q", screen)
	}
	pos := e.CursorPosition()
	if !visible || pos.X != 2 || strings.Split(screen, "\n")[pos.Y] != "> input" {
		t.Errorf("cursor not restored: visible=%v position=%+v screen=%q", visible, pos, screen)
	}
}

func TestCursedRenderer_insertAbovePendingView(t *testing.T) {
	t.Parallel()

	for _, size := range [][2]int{{40, 10}, {80, 24}, {120, 40}, {12, 3}} {
		for _, full := range []bool{false, true} {
			for _, hidden := range []bool{false, true} {
				t.Run(fmt.Sprintf("%dx%d/full=%v/hidden=%v", size[0], size[1], full, hidden), func(t *testing.T) {
					e := vt.NewEmulator(size[0], size[1])
					defer e.Close()
					visible := true
					e.SetCallbacks(vt.Callbacks{CursorVisibility: func(v bool) { visible = v }})
					if _, err := e.Write([]byte("SHELL\r\n")); err != nil {
						t.Fatal(err)
					}
					var output bytes.Buffer
					r := newCursedRenderer(io.MultiWriter(&output, e), []string{"TERM=xterm-256color"}, 80, 24)
					r.setNoInput(true)
					r.resize(size[0], size[1])
					frame := "CONTROLS\n> input\nSTATUS"
					if full {
						frame = strings.Repeat("managed\n", size[1]-3) + frame
					}
					view := NewView(frame)
					if !hidden {
						view.Cursor = NewCursor(2, strings.Count(frame, "\n")-1)
					}
					r.render(view) // The first View and resize have not been flushed.
					lines := make([]string, 80)
					for i := range lines {
						lines[i] = fmt.Sprintf("ROW-%03d", i)
					}
					if err := r.insertAbove(strings.Join(lines, "\n")); err != nil {
						t.Fatal(err)
					}
					assertTerminalLines(t, e, append([]string{"SHELL"}, lines...))
					screen := e.String()
					if !strings.Contains(screen, frame) {
						t.Fatalf("missing View immediately after insertion: %q", screen)
					}
					if visible == hidden {
						t.Errorf("cursor visibility = %v, want %v", visible, !hidden)
					}
					if !hidden {
						pos := e.CursorPosition()
						if pos.X != 2 || strings.Split(screen, "\n")[pos.Y] != "> input" {
							t.Errorf("cursor not restored: %+v screen=%q", pos, screen)
						}
					}
					if e.IsAltScreen() || strings.Contains(output.String(), "\x1b[3J") {
						t.Error("insertion changed native history ownership")
					}
					before := terminalHistory(e)
					if err := r.flush(false); err != nil {
						t.Fatal(err)
					}
					if before != terminalHistory(e) {
						t.Error("identical subsequent View changed history")
					}
				})
			}
		}
	}
}

func TestCursedRenderer_insertAbovePendingShrink(t *testing.T) {
	t.Parallel()

	e := vt.NewEmulator(40, 10)
	defer e.Close()
	if _, err := e.Write([]byte("SHELL\r\n")); err != nil {
		t.Fatal(err)
	}
	r := newCursedRenderer(e, []string{"TERM=xterm-256color"}, 40, 10)
	r.setNoInput(true)
	r.setColorProfile(colorprofile.TrueColor)
	r.render(NewView(strings.Repeat("old frame\n", 7) + "old footer"))
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	view := NewView("CONTROLS\n> input\nSTATUS")
	view.Cursor = NewCursor(2, 1)
	r.render(view)
	exact := strings.Repeat("界", 20)
	payload := "\x1b[31m" + exact + exact + "\x1b[0m\nemoji 👋\nlast"
	if err := r.insertAbove(payload); err != nil {
		t.Fatal(err)
	}
	history := terminalHistory(e)
	if strings.Count(history, exact) != 2 {
		t.Fatalf("lost styled/wide/exact-margin content: %q", history)
	}
	assertTerminalLines(t, e, []string{"SHELL", "emoji 👋", "last"})
	if strings.Contains(history, "old frame") || strings.Contains(history, "old footer") {
		t.Errorf("pending shrink left old content in history: %q", history)
	}
	screen := e.String()
	pos := e.CursorPosition()
	if !strings.Contains(screen, view.Content) || pos.X != 2 || strings.Split(screen, "\n")[pos.Y] != "> input" {
		t.Errorf("View or cursor not restored: %+v screen=%q", pos, screen)
	}
}

func TestCursedRenderer_insertAboveFollowingFrame(t *testing.T) {
	t.Parallel()

	for _, size := range [][2]int{{1, 1}, {2, 2}, {12, 3}, {40, 10}} {
		for _, frameHeight := range []int{0, 1, size[1], size[1] + 3} {
			t.Run(fmt.Sprintf("%dx%d/frame=%d", size[0], size[1], frameHeight), func(t *testing.T) {
				e := vt.NewEmulator(size[0], size[1])
				defer e.Close()
				r := newCursedRenderer(e, []string{"TERM=xterm-256color"}, size[0], size[1])
				r.setNoInput(true)
				frame := strings.TrimSuffix(strings.Repeat("M\n", frameHeight), "\n")
				r.render(NewView(frame))
				lines := []string{"A", "B", "C", "D", "E", "F"}
				if err := r.insertAbove(strings.Join(lines, "\n")); err != nil {
					t.Fatal(err)
				}
				assertTerminalLines(t, e, lines)
				r.render(NewView(strings.ReplaceAll(frame, "M", "N")))
				if err := r.flush(false); err != nil {
					t.Fatal(err)
				}
				if err := r.insertAbove("G"); err != nil {
					t.Fatal(err)
				}
				assertTerminalLines(t, e, append(lines, "G"))
				screen := e.String()
				if strings.Contains(screen, "M") || strings.Count(screen, "N") != min(frameHeight, size[1]) {
					t.Errorf("subsequent View not restored: %q", screen)
				}
			})
		}
	}
}

func TestCursedRenderer_insertAboveAltScreen(t *testing.T) {
	t.Parallel()

	e := vt.NewEmulator(40, 10)
	defer e.Close()
	r := newCursedRenderer(e, []string{"TERM=xterm-256color"}, 40, 10)
	r.setNoInput(true)
	view := NewView("alternate")
	view.AltScreen = true
	r.render(view)
	if err := r.insertAbove("must not print"); err != nil {
		t.Fatal(err)
	}
	if !e.IsAltScreen() || !strings.Contains(e.String(), "alternate") || strings.Contains(e.String(), "must not print") {
		t.Fatalf("pending alternate screen not preserved: %q", e.String())
	}
	before := e.String()
	if err := r.insertAbove("must not print either"); err != nil {
		t.Fatal(err)
	}
	if before != e.String() {
		t.Error("insertion changed alternate screen")
	}
	r.render(NewView("inline"))
	if err := r.insertAbove("printed"); err != nil {
		t.Fatal(err)
	}
	assertTerminalLines(t, e, []string{"printed", "inline"})
	if e.IsAltScreen() {
		t.Error("did not leave alternate screen")
	}
}

func TestCursedRenderer_insertAboveWriterErrors(t *testing.T) {
	t.Parallel()

	failure := errors.New("terminal disconnected")
	for _, tt := range []struct {
		name    string
		pending bool
		full    bool
		failAt  int
	}{
		{name: "initial flush", pending: true, failAt: 1},
		{name: "bounded insertion", failAt: 1},
		{name: "cursor restoration", failAt: 2},
		{name: "release full frame", full: true, failAt: 1},
		{name: "append after release", full: true, failAt: 2},
		{name: "repaint full frame", full: true, failAt: 4},
	} {
		for _, short := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/short=%v", tt.name, short), func(t *testing.T) {
				r := newCursedRenderer(io.Discard, []string{"TERM=xterm-256color"}, 40, 10)
				r.setNoInput(true)
				frame := "CONTROLS\n> input\nSTATUS"
				if tt.full {
					frame = strings.Repeat("managed\n", 7) + frame
				}
				view := NewView(frame)
				view.Cursor = NewCursor(2, strings.Count(frame, "\n")-1)
				r.render(view)
				if !tt.pending {
					if err := r.flush(false); err != nil {
						t.Fatal(err)
					}
				}
				r.w = &insertErrorWriter{failAt: tt.failAt, short: short, err: failure}
				want := failure
				if short {
					want = io.ErrShortWrite
				}
				if err := r.insertAbove("first\nsecond"); !errors.Is(err, want) {
					t.Errorf("insertion error = %v, want %v", err, want)
				}
			})
		}
	}
}

type insertErrorWriter struct {
	failAt int
	writes int
	short  bool
	err    error
}

func (w *insertErrorWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		if w.short {
			return len(p) / 2, nil
		}
		return 0, w.err
	}
	return len(p), nil
}

func terminalHistory(e *vt.Emulator) string {
	var b strings.Builder
	for _, line := range e.Scrollback().Lines() {
		for _, cell := range line {
			b.WriteString(cell.Content)
		}
		b.WriteByte('\n')
	}
	b.WriteString(e.String())
	return b.String()
}

func assertTerminalLines(t *testing.T, e *vt.Emulator, wants []string) {
	t.Helper()
	history := terminalHistory(e)
	lines := strings.Split(history, "\n")
	previous := -1
	for _, want := range wants {
		count, at := 0, -1
		for i, line := range lines {
			if strings.TrimRight(line, " ") == want {
				count++
				at = i
			}
		}
		if count != 1 || at <= previous {
			t.Fatalf("missing, duplicate, or out-of-order %q: %q", want, history)
		}
		previous = at
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
