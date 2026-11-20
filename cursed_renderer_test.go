package tea

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func TestCursedRendererEnablesUnicodeModeForGraphemeWidth(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 10, 3)
	r.start()
	r.setWidthMethod(ansi.GraphemeWidth)
	r.render(NewView("hello"))
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), ansi.SetModeUnicodeCore) {
		t.Fatalf("grapheme rendering did not enable Unicode mode; output %q", out.String())
	}
}

func TestCursedRendererWcWidthDoesNotEnableUnicodeMode(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 10, 3)
	r.start()
	r.setWidthMethod(ansi.WcWidth)
	r.render(NewView("hello"))
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	out.Reset()
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	if got := strings.Contains(out.String(), ansi.ResetModeUnicodeCore); got != false {
		t.Fatalf("close reset Unicode mode = %v, want false; output %q", got, out.String())
	}

	// Discard shutdown output so only the resume sequence can satisfy the assertion.
	out.Reset()
	r.start()
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if got := strings.Contains(out.String(), ansi.SetModeUnicodeCore); got != false {
		t.Fatalf("resume enabled Unicode mode = %v, want false; output %q", got, out.String())
	}
}

func TestCursedRendererResumeRestoresUnicodeMode(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 10, 3)
	r.start()
	r.setWidthMethod(ansi.GraphemeWidth)
	r.render(NewView("hello"))
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	out.Reset()
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	if got := strings.Contains(out.String(), ansi.ResetModeUnicodeCore); got != true {
		t.Fatalf("close reset Unicode mode = %v, want true; output %q", got, out.String())
	}

	// Discard shutdown output so only the resume sequence can satisfy the assertion.
	out.Reset()
	r.start()
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if got := strings.Contains(out.String(), ansi.SetModeUnicodeCore); got != true {
		t.Fatalf("resume enabled Unicode mode = %v, want true; output %q", got, out.String())
	}
}

func TestCursedRendererSkipsUnchangedView(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 4, 3)
	r.start()
	r.render(NewView("hello"))
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if drawn, err := r.flush(false); err != nil || drawn || out.Len() != 0 {
		t.Fatalf("unchanged view: drawn=%v, err=%v, output=%q", drawn, err, out.String())
	}
}

func TestCursedRendererDrawsChangedEmptyAndTallViews(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		before, after View
	}{
		{"empty inline becomes nonempty", View{}, NewView("new")},
		{"empty alt screen becomes nonempty", View{AltScreen: true}, View{Content: "new", AltScreen: true}},
		{"tall inline changes visible rows", NewView("a\nb\nc\nd\ne"), NewView("a\nb\nx\ny\nz")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 4, 3)
			r.start()
			r.render(tt.before)
			if _, err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			out.Reset()
			r.render(tt.after)
			if drawn, err := r.flush(false); err != nil || !drawn || out.Len() == 0 {
				t.Fatalf("changed view: drawn=%v, err=%v, output=%q", drawn, err, out.String())
			}
		})
	}
}

func TestCursedRendererSkipsUnchangedEmptyAndTallViews(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		view View
	}{
		{"empty inline", View{}},
		{"empty alt screen", View{AltScreen: true}},
		{"tall inline", NewView("a\nb\nc\nd\ne\nf\ng")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 4, 3)
			r.start()
			r.render(tt.view)
			if _, err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			// Measure only the second flush, after the initial frame is drawn.
			out.Reset()
			if drawn, err := r.flush(false); err != nil || drawn || out.Len() != 0 {
				t.Fatalf("unchanged view: drawn=%v, err=%v, output=%q", drawn, err, out.String())
			}
		})
	}
}

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

// uncomparableColor is a color.Color implementation whose dynamic type is
// uncomparable (it contains a slice), so comparing two values with == panics.
type uncomparableColor struct {
	rgba color.RGBA
	pad  []int
}

func (c uncomparableColor) RGBA() (r, g, b, a uint32) {
	return c.rgba.RGBA()
}

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1848
func TestCursedRenderer_uncomparableViewColors(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.start()

	red := color.RGBA{R: 0xff, A: 0xff}
	view := NewView("hello")
	view.AltScreen = true
	view.ForegroundColor = uncomparableColor{rgba: red, pad: []int{1}}
	view.BackgroundColor = uncomparableColor{rgba: red, pad: []int{2}}
	view.Cursor = &Cursor{Color: uncomparableColor{rgba: red, pad: []int{3}}}

	// Render the same view twice. The second flush takes the viewEquals
	// early-return path, which compares the color.Color interface values.
	for range 2 {
		r.render(view)
		if _, err := r.flush(false); err != nil {
			t.Fatal(err)
		}
	}
}

// Colors are compared by value: the same color, whatever its type, doesn't
// redraw the view, while a different color or a reset does.
func TestCursedRenderer_viewColorChanges(t *testing.T) {
	t.Parallel()

	red := color.RGBA{R: 0xff, A: 0xff}
	blue := color.RGBA{B: 0xff, A: 0xff}
	for _, tt := range []struct {
		name  string
		set   func(*View, color.Color)
		code  string
		reset string
	}{
		{"foreground", func(v *View, c color.Color) { v.ForegroundColor = c }, "10", "110"},
		{"background", func(v *View, c color.Color) { v.BackgroundColor = c }, "11", "111"},
		{"cursor", func(v *View, c color.Color) { v.Cursor = NewCursor(0, 0); v.Cursor.Color = c }, "12", "112"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
			render := func(c color.Color) bool {
				t.Helper()
				view := NewView("hello")
				tt.set(&view, c)
				r.render(view)
				drew, err := r.flush(false)
				if err != nil {
					t.Fatal(err)
				}
				return drew
			}

			render(uncomparableColor{rgba: red, pad: []int{1}})
			if want := "\x1b]" + tt.code + ";#ff0000\a"; !strings.Contains(out.String(), want) {
				t.Fatalf("missing red color sequence %q in %q", want, out.String())
			}

			out.Reset()
			if render(uncomparableColor{rgba: red, pad: []int{2}}) || out.Len() != 0 {
				t.Fatalf("unchanged color redrew the view: %q", out.String())
			}
			if render(red) || out.Len() != 0 {
				t.Fatalf("the same color of another type redrew the view: %q", out.String())
			}

			render(uncomparableColor{rgba: blue})
			if want := "\x1b]" + tt.code + ";#0000ff\a"; !strings.Contains(out.String(), want) {
				t.Fatalf("missing blue color sequence %q in %q", want, out.String())
			}

			out.Reset()
			render(nil)
			if want := "\x1b]" + tt.reset + "\a"; !strings.Contains(out.String(), want) {
				t.Fatalf("missing color reset %q in %q", want, out.String())
			}
		})
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
		if _, err := r.flush(false); err != nil {
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
		if _, err := r.flush(false); err != nil {
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
		via  int
		to   int
		env  []string
		want int
	}{
		// Two frame rows sit above the cursor, and the rule among them spans 99
		// of the 100 columns, so at 60 columns it covers two terminal rows.
		{name: "rewrapping terminal", rule: strings.Repeat("-", 99), to: 60, want: 3},
		// The 100-column frame clips the rule before the terminal can reflow it.
		{name: "clipped content", rule: strings.Repeat("-", 199), to: 60, want: 3},
		// Multiple resizes before the next draw must not stack cursor moves.
		{name: "coalesced resize", rule: strings.Repeat("-", 99), via: 80, to: 60, want: 3},
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
			if _, err := r.flush(false); err != nil {
				t.Fatalf("flush initial frame: %v", err)
			}

			output.Reset()
			if tc.via > 0 {
				r.resize(tc.via, 24)
				r.render(frame)
			}
			r.resize(tc.to, 24)
			r.render(frame)
			if _, err := r.flush(false); err != nil {
				t.Fatalf("flush after resize: %v", err)
			}

			if got := rowsMovedUp(t, output.String()); got != tc.want {
				t.Errorf("moved up %d rows to reach the top of the frame, want %d", got, tc.want)
			}
		})
	}
}

func TestCursedRenderer_insertAboveBeforeFirstFlush(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	renderer := newCursedRenderer(&output, []string{"TERM=xterm-256color"}, 80, 24)
	renderer.render(NewView("footer"))

	if err := renderer.insertAbove("startup"); err != nil {
		t.Fatal(err)
	}

	// The one-line view is drawn first, then the line is inserted above it,
	// making room for one line rather than for the terminal's height.
	got := output.String()
	frame := "\r" + ansi.EraseScreenBelow + "footer"
	insert := "\r\n" + ansi.CursorUp(1) +
		ansi.InsertLine(1) + ansi.EraseLineRight + "startup" + "\r\n"
	if !strings.HasSuffix(got, frame+insert) {
		t.Fatalf("unexpected insertAbove output:\nwant suffix: %q\n         got: %q", frame+insert, got)
	}
}

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1780
func TestCursedRenderer_resizeStaleFrame(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)

	// Render the initial frame at 80x24.
	initialView := NewView("initial 80x24 content")
	initialView.AltScreen = true
	r.render(initialView)
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	out.Reset()

	// Simulate window resize: resize() is called first, updating dimensions.
	r.resize(100, 30)

	// In the race window, the 60Hz ticker fires flush() BEFORE render() provides the new View.
	// flush() must NOT output the stale initialView at the new 100x30 size.
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	if out.Len() > 0 {
		t.Fatalf("expected no output during stale resize race window, got %q", out.String())
	}

	// Now render() provides the new View matching the new 100x30 size.
	newView := NewView("resized 100x30 content")
	newView.AltScreen = true
	r.render(newView)

	// The subsequent flush() must draw the new view.
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out.String(), "resized 100x30 content") {
		t.Fatalf("expected output to contain resized view, got %q", out.String())
	}
}

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1802
//
// insertAbove must measure wrapped lines with the active screen buffer's
// width method. When the terminal never answers the Unicode-core query the
// buffer stays on WcWidth, but insertAbove used grapheme semantics, so wide
// grapheme clusters inserted the wrong number of lines.
func TestCursedRenderer_insertAboveUsesActiveWidthMethod(t *testing.T) {
	t.Parallel()

	insertNewlines := func(method ansi.Method, s string) int {
		var out bytes.Buffer
		r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 8, 4)
		r.setWidthMethod(method)
		if got := r.cellbuf.WidthMethod(); got != method {
			t.Fatalf("expected width method %v, got %v", method, got)
		}
		if err := r.insertAbove(s); err != nil {
			t.Fatal(err)
		}
		// One "\n" per inserted line, plus the trailing newline after each
		// printed line.
		return bytes.Count(out.Bytes(), []byte("\n"))
	}

	// "👨‍👩‍👧‍👦" is a single grapheme cluster (width 2) made of four wide
	// codepoints (width 8), so the wrapped-line count differs per method.
	// With an 8-column buffer: grapheme width 8 fits in 1 row, wcwidth 14
	// wraps to 2 rows.
	const line = "👨‍👩‍👧‍👦aaaaaa"

	// WcWidth fallback: 1 line + (14-1)/8 extra -> 2 inserted lines, 3 newlines.
	if got := insertNewlines(ansi.WcWidth, line); got != 3 {
		t.Errorf("WcWidth: expected 3 newlines, got %d", got)
	}

	// GraphemeWidth (Unicode-core mode): 1 line + (8-1)/8 extra -> 1 inserted
	// line, 2 newlines.
	if got := insertNewlines(ansi.GraphemeWidth, line); got != 2 {
		t.Errorf("GraphemeWidth: expected 2 newlines, got %d", got)
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
			r.render(NewView("view\nview"))
			if _, err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			out.Reset()
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

func TestCursedRenderer_insertAboveFlushesQueuedView(t *testing.T) {
	var output bytes.Buffer
	r := newCursedRenderer(
		&output,
		[]string{"TERM=xterm-256color", "TERM_PROGRAM=Apple_Terminal"},
		80,
		24,
	)

	r.render(NewView("stale frame\nstale row"))
	if _, err := r.flush(false); err != nil {
		t.Fatalf("flush initial view: %v", err)
	}

	output.Reset()
	r.render(NewView("current frame"))
	if err := r.insertAbove("committed block"); err != nil {
		t.Fatalf("insert above: %v", err)
	}

	if r.lastView == nil || r.lastView.Content != "current frame" {
		t.Fatalf("insertAbove did not flush the queued view: %#v", r.lastView)
	}
	written := output.String()
	currentAt := strings.Index(written, "current frame")
	committedAt := strings.Index(written, "committed block")
	if currentAt < 0 || committedAt < 0 {
		t.Fatalf("output missing markers (current=%d, committed=%d): %q", currentAt, committedAt, written)
	}
	if currentAt > committedAt {
		t.Fatalf("inserted scrollback before flushing current frame: %q", written)
	}
}

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1881
func TestCursedRenderer_insertAboveAfterLeavingAltScreen(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	alt := NewView("alt screen view")
	alt.AltScreen = true
	r.render(alt)
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	// The update leaves the alt screen and prints a line before the next
	// frame is drawn. The line must be printed on the normal screen, after
	// leaving the alt screen, or it is lost with it.
	r.render(NewView("inline view"))
	if err := r.insertAbove("printed line"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	assertInOrder(t, got, ansi.SetModeAltScreenSaveCursor, ansi.ResetModeAltScreenSaveCursor, "inline view", "printed line")
	if n := strings.Count(got, "printed line"); n != 1 {
		t.Fatalf("expected the line to be printed once, got %d times in %q", n, got)
	}
}

func TestCursedRenderer_insertAboveBeforeEnteringAltScreen(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
	r.render(NewView("inline view"))
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	// The update enters the alt screen and prints a line before the next
	// frame is drawn. The line is printed on the normal screen, before
	// entering the alt screen, so it's still there after leaving it.
	alt := NewView("alt screen view")
	alt.AltScreen = true
	r.render(alt)
	if err := r.insertAbove("printed line"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.flush(false); err != nil {
		t.Fatal(err)
	}

	assertInOrder(t, out.String(), "inline view", "printed line", ansi.SetModeAltScreenSaveCursor, "alt screen view")
}
