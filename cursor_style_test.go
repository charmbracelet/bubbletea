package tea

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestParseCursorStyleReport(t *testing.T) {
	for style := 0; style <= 6; style++ {
		for _, framing := range [][2]string{{"\x1bP", "\x1b\\"}, {"\x90", "\x9c"}} {
			report := framing[0] + "1$r" + string(rune('0'+style)) + " q" + framing[1]
			if got, ok := parseCursorStyleReport(report); !ok || got != style {
				t.Fatalf("report %q = %d, %v", report, got, ok)
			}
		}
	}
	for _, report := range []string{"", "\x1bP0$r q\x1b\\", "\x1bP1$r7 q\x1b\\", "\x1bP1$r6m\x1b\\"} {
		if _, ok := parseCursorStyleReport(report); ok {
			t.Fatalf("accepted unrelated/invalid report %q", report)
		}
	}
}

func TestRestoreCursorStyle(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		var output bytes.Buffer
		r := newCursedRenderer(&output, []string{"TERM=xterm"}, 80, 24)
		r.saveCursorStyle(6)
		r.saveCursorStyle(2) // A later query must not overwrite the host style.
		r.start()
		view := NewView("query")
		view.Cursor = NewCursor(0, 0)
		r.render(view)
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
		if hidden {
			view.Cursor = nil
			r.render(view)
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}
		}
		output.Reset()
		if err := r.close(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), ansi.SetCursorStyle(6)) {
			t.Fatalf("hidden=%v: missing host style in %q", hidden, output.String())
		}
		// Restart must restore the application's blinking block, too.
		if !hidden {
			output.Reset()
			r.start()
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), ansi.SetCursorStyle(1)) {
				t.Fatalf("restart did not restore app style: %q", output.String())
			}
		}
	}
}

func TestPreserveCursorStyle(t *testing.T) {
	var output bytes.Buffer
	r := newCursedRenderer(&output, []string{"TERM=xterm"}, 80, 24)
	r.start()
	v := NewView("native")
	v.Cursor = NewCursor(2, 0)
	v.Cursor.PreserveStyle = true
	r.render(v)
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	r.start()
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	v.Cursor = nil
	r.render(v)
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if err := r.close(); err != nil {
		t.Fatal(err)
	}
	for style := 0; style <= 6; style++ {
		if strings.Contains(output.String(), ansi.SetCursorStyle(style)) {
			t.Fatalf("preserved cursor emitted style %d: %q", style, output.String())
		}
	}
}
