package tea

import (
	"bytes"
	"image/color"
	"strings"
	"testing"
)

type sliceColor []uint32

func (c sliceColor) RGBA() (uint32, uint32, uint32, uint32) {
	return c[0], c[1], c[2], c[3]
}

func TestCursedRenderer_noncomparableColors(t *testing.T) {
	t.Parallel()
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
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("rendering a valid color.Color panicked: %v", r)
				}
			}()
			var out bytes.Buffer
			r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)
			view := NewView("hello")
			render := func(c color.Color) {
				t.Helper()
				tt.set(&view, c)
				r.render(view)
				if err := r.flush(false); err != nil {
					t.Fatal(err)
				}
			}
			render(sliceColor{65535, 0, 0, 65535})
			if want := "\x1b]" + tt.code + ";#ff0000\a"; !strings.Contains(out.String(), want) {
				t.Fatalf("missing red color sequence %q in %q", want, out.String())
			}
			out.Reset()
			render(sliceColor{65535, 0, 0, 65535})
			if out.Len() != 0 {
				t.Fatalf("unchanged color redrew view: %q", out.String())
			}
			render(sliceColor{0, 0, 65535, 65535})
			if want := "\x1b]" + tt.code + ";#0000ff\a"; !strings.Contains(out.String(), want) {
				t.Fatalf("missing blue color sequence %q in %q", want, out.String())
			}
			out.Reset()
			render(color.RGBA{B: 255, A: 255})
			if out.Len() != 0 {
				t.Fatalf("equivalent color redrew view: %q", out.String())
			}
			render(nil)
			if want := "\x1b]" + tt.reset + "\a"; !strings.Contains(out.String(), want) {
				t.Fatalf("missing color reset %q in %q", want, out.String())
			}
		})
	}
}
