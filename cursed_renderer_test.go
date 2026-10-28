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

func TestCursedRenderer_insertAboveLineBoundaries(t *testing.T) {
	t.Parallel()

	const width = 40
	tests := []struct {
		name string
		line string
		rows int
	}{
		{name: "short", line: strings.Repeat("-", width-2) + "#", rows: 1},
		{name: "exact width", line: strings.Repeat("-", width-1) + "#", rows: 1},
		{name: "one cell over", line: strings.Repeat("-", width) + "#", rows: 2},
		{name: "two exact rows", line: strings.Repeat("-", 2*width-1) + "#", rows: 2},
		{name: "three exact rows", line: strings.Repeat("-", 3*width-1) + "#", rows: 3},
		{name: "wide characters", line: strings.Repeat("界", width-1) + "-#", rows: 2},
		{name: "styled", line: "\x1b[31m" + strings.Repeat("-", 2*width-1) + "#" + ansi.ResetStyle, rows: 2},
		{name: "multiple lines", line: strings.Repeat("-", width-1) + "#\n" + strings.Repeat("-", 2*width-1) + "#", rows: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, width, 2)
			r.scr.SetPosition(0, 0)
			if err := r.insertAbove(tt.line); err != nil {
				t.Fatal(err)
			}

			got := out.String()
			prefix := "\r" + ansi.CursorDown(1) + strings.Repeat("\n", tt.rows) + ansi.CursorUp(tt.rows+1) + ansi.InsertLine(tt.rows)
			if !strings.HasPrefix(got, prefix) {
				t.Fatalf("expected room for %d terminal rows with prefix %q, got %q", tt.rows, prefix, got)
			}
			for _, line := range strings.Split(tt.line, "\n") {
				if ansi.StringWidth(line)%width == 0 && strings.Contains(got, line+ansi.EraseLineRight) {
					t.Fatalf("erasing at the right margin would delete the inserted line's last cell, got %q", got)
				}
			}
		})
	}
}
