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

// rowsMovedUp sums the cursor-up movement the renderer emits before it starts
// erasing, which is how far it reckons the top of the previous frame is.
func rowsMovedUp(t *testing.T, output string) int {
	t.Helper()

	prefix, _, found := strings.Cut(output, ansi.EraseScreenBelow)
	if !found {
		t.Fatalf("output never erased the old frame: %q", output)
	}
	// Summed rather than matched once: the move up is emitted in two pieces,
	// the rows skipped past the rewrap and the renderer's own move to the top
	// of the frame, and only their total says where the erase begins.
	var rows int
	for _, m := range cursorUpSeq.FindAllStringSubmatch(prefix, -1) {
		if m[1] == "" {
			rows++ // a bare CUU moves one row
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			t.Fatalf("unreadable cursor-up count in %q: %v", m[0], err)
		}
		rows += n
	}
	return rows
}

var cursorUpSeq = regexp.MustCompile(`\x1b\[(\d*)A`)

// A narrowed terminal rewraps the frame it is holding, so the row the renderer
// counts as one is now two and the erase has further to travel to reach the top
// of the frame.
func TestCursedRenderer_resizeSkipsReflowedRows(t *testing.T) {
	for _, tc := range []struct {
		name string
		rule string
		to   int
		env  []string
		want int
	}{
		// Two frame rows sit above the cursor, and the rule among them spans 99
		// of the 100 columns, so at 60 columns it covers two terminal rows.
		{name: "rewrapping terminal", rule: strings.Repeat("-", 99), to: 60, want: 3},
		// 21 double-width cells will not divide into rows of 21 columns: each
		// row takes 10 of them and leaves its last column empty, so the row
		// covers three terminal rows where dividing its width would say two.
		{name: "double-width cells", rule: strings.Repeat("你", 21), to: 21, want: 4},
		// tmux has rewrapped pane text on resize since 1.8, so it gets the
		// same count as any other rewrapping terminal.
		{name: "tmux", rule: strings.Repeat("-", 99), to: 60, want: 3,
			env: []string{"TMUX=/tmp/tmux-501/default,1,0"}},
		// GNU screen truncates instead, and moving up the extra rows would
		// erase scrollback the renderer never drew.
		{name: "screen", rule: strings.Repeat("-", 99), to: 60, want: 2,
			env: []string{"STY=12345.pts-0.host"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			env := append([]string{"TERM=xterm-256color", "TERM_PROGRAM=Apple_Terminal"}, tc.env...)
			r := newCursedRenderer(&output, env, 100, 24)

			frame := NewView("\n" + tc.rule + "\n> \n" + tc.rule + "\nstatus")
			frame.Cursor = NewCursor(2, 2)
			r.render(frame)
			if err := r.flush(false); err != nil {
				t.Fatalf("flush initial frame: %v", err)
			}

			output.Reset()
			r.resize(tc.to, 24)
			r.render(frame)
			if err := r.flush(false); err != nil {
				t.Fatalf("flush after resize: %v", err)
			}

			if got := rowsMovedUp(t, output.String()); got != tc.want {
				t.Errorf("moved up %d rows to reach the top of the frame, want %d", got, tc.want)
			}
		})
	}
}
