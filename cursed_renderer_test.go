package tea

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// TestResizeEmitsEagerEraseInline verifies that in inline mode,
// cursedRenderer.resize eagerly writes a physical erase sequence to the writer
// instead of deferring it to the next render. Otherwise, after
// [cursedRenderer.insertAbove] resets the renderer's tracked cursor to (0, 0),
// the next render's clearUpdate path emits a move(0, 0) that degenerates to a
// no-op, and the trailing ED-0 fires at the wrong physical row.
func TestResizeEmitsEagerEraseInline(t *testing.T) {
	cases := []struct {
		name string
		// cursorY is the renderer's tracked cursor row at the moment resize
		// fires. 0 mirrors the post-insertAbove state; >0 mirrors steady-state
		// mid-render.
		cursorY int
		// want is a substring the resize output must contain.
		want []byte
	}{
		{
			name:    "after_insert_above",
			cursorY: 0,
			want:    []byte("\r" + ansi.EraseScreenBelow),
		},
		{
			name:    "mid_view",
			cursorY: 3,
			want:    []byte("\r" + ansi.CursorUp(3) + ansi.EraseScreenBelow),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := newCursedRenderer(&buf, []string{"TERM=xterm-256color"}, 80, 24)

			r.render(NewView("a0\na1\na2\na3\na4\na5\na6"))
			if err := r.flush(false); err != nil {
				t.Fatalf("initial flush: %v", err)
			}
			// Pin the renderer's tracked cursor to a known position so the
			// assertions are independent of transformLine's end-of-line
			// behavior. (0, 0) mirrors the post-insertAbove state.
			r.scr.SetPosition(0, tc.cursorY)
			buf.Reset()

			r.resize(60, 24)

			got := buf.Bytes()
			if !bytes.Contains(got, tc.want) {
				t.Fatalf("resize did not emit the expected eager erase\n want substring %q\n got            %q",
					tc.want, got)
			}
		})
	}
}

// TestResizeAltScreenSkipsEagerErase verifies that the inline-mode eager-erase
// path does not fire when the active view is alt-screen. Alt-screen uses
// absolute positioning and its own clear contract; the inline workaround would
// be both incorrect and unnecessary there.
func TestResizeAltScreenSkipsEagerErase(t *testing.T) {
	var buf bytes.Buffer
	r := newCursedRenderer(&buf, []string{"TERM=xterm-256color"}, 80, 24)

	v := NewView("a0\na1")
	v.AltScreen = true
	r.render(v)
	if err := r.flush(false); err != nil {
		t.Fatalf("initial flush: %v", err)
	}
	buf.Reset()

	r.resize(60, 24)

	if buf.Len() != 0 {
		t.Fatalf("resize wrote %d bytes in alt-screen mode (should write zero): %q",
			buf.Len(), buf.Bytes())
	}
}

// TestResizeBeforeFirstViewSkipsEagerErase verifies that resize before any
// view has been set (i.e. the initial resize during Program startup) does not
// write an erase sequence — there is nothing on the terminal yet to erase.
func TestResizeBeforeFirstViewSkipsErase(t *testing.T) {
	var buf bytes.Buffer
	r := newCursedRenderer(&buf, []string{"TERM=xterm-256color"}, 80, 24)

	r.resize(60, 24)

	if buf.Len() != 0 {
		t.Fatalf("resize wrote %d bytes before any view was set: %q",
			buf.Len(), buf.Bytes())
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
