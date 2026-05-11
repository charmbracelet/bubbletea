package tea

import (
	"bytes"
	"testing"

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
