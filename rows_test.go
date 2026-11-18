package tea

import (
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"reflect"
	"slices"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// rowsFrame is a full-screen frame of styled rows, each closing its styles
// and hyperlinks, with wide runes in some; the rows in dirty carry one of
// the things a row drawn on its own would leave out: a style or a link left
// open, a sequence waiting for a cell, a mark for the row above.
func rowsFrame(rng *rand.Rand, w, h int, dirty map[int]bool) string {
	words := []string{"plain", "\x1b[1mbold\x1b[m", "\x1b[38;5;196mred\x1b[m", "世界", "\x1b]8;;https://x.y\x07link\x1b]8;;\x07", "😀", "\x1b[48;2;10;20;30m bg \x1b[m"}
	open := []string{"\x1b[7mopen", "\x1b]8;;https://x.y\x07open", "\x9d8;;https://x.y\x9copen", "\x1b]8;id=1;\x07", "\x1b_Gx\x1b\\", "\x1b[1", "\x1b]0;title", "\xe4\xb8\x96\xe7"}
	rows := make([]string, h)
	for y := range rows {
		k := -1 // the row's flaw: one of open, or the mark when len(open)
		if dirty[y] {
			k = rng.IntN(len(open) + 1)
		}
		if k == len(open) && rng.IntN(3) == 0 { // a row of nothing but a sequence waiting for a cell
			rows[y] = "\x1b_Gx\x1b\\"
			continue
		}
		var b strings.Builder
		if k == len(open) {
			b.WriteString("\u0301")
		}
		for n := 0; n < w; n += 6 {
			b.WriteString(words[rng.IntN(len(words))])
			b.WriteByte(' ')
		}
		if k >= 0 && k < len(open) {
			b.WriteString(open[k])
		} else {
			b.WriteString("\x1b[m")
		}
		rows[y] = b.String()
	}
	return strings.Join(rows, "\n")
}

// TestCursedRenderer_changedRows: a frame that redraws only its changed rows
// leaves the buffer as a frame drawn whole does, cell for cell, whether its
// rows can be drawn apart or not, and when they scrolled, under either width
// method.
func TestCursedRenderer_changedRows(t *testing.T) {
	t.Parallel()

	for _, m := range []ansi.Method{ansi.WcWidth, ansi.GraphemeWidth} {
		const w, h = 60, 12
		rng := rand.New(rand.NewPCG(1, 2))
		r := newCursedRenderer(io.Discard, []string{"TERM=xterm-256color"}, w, h)
		r.start()
		r.setWidthMethod(m)

		rows := strings.Split(rowsFrame(rng, w, h, nil), "\n")
		for i := range 1500 {
			for range rng.IntN(3) + 1 { // a few rows change, the rest stay
				y := rng.IntN(h)
				rows[y] = strings.Split(rowsFrame(rng, w, h, map[int]bool{y: i%7 == 0}), "\n")[y]
			}
			if i%5 == 0 { // or scroll, by one row or a few, either way
				n := rng.IntN(h-1) + 1
				rows = append(rows[n:], rows[:n]...)
			}

			v := NewView(strings.Join(rows, "\n"))
			v.AltScreen = true
			r.render(v)
			if err := r.flush(false); err != nil {
				t.Fatal(err)
			}

			want := uv.NewScreenBuffer(w, h)
			want.Method = m
			uv.NewStyledString(v.Content).Draw(want, want.Bounds())
			for y := range h {
				if got, exp := r.cellbuf.Line(y), want.Line(y); !reflect.DeepEqual(got, exp) {
					t.Fatalf("method %d, frame %d, row %d:\n got %q\nwant %q", m, i, y, got.Render(), exp.Render())
				}
			}
		}
	}
}

// TestCursedRenderer_changedRowsWritten: the rows a frame changed reach the
// terminal, whether the frame is drawn whole or row by row. They are drawn
// into the cells without the buffer tracking them, so the renderer diffs
// them only because their lines are touched first.
func TestCursedRenderer_changedRowsWritten(t *testing.T) {
	t.Parallel()

	const h = 10
	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 20, h)
	r.start()

	rows := make([]string, h)
	for i, c := range []struct {
		word    string
		changed []int
	}{
		{"zero", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}, // the first frame, drawn whole
		{"one", []int{0}}, // drawn whole again, its rows checked
		{"two", []int{3}}, // drawn row by row
		{"six", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}, // drawn whole
	} {
		for _, y := range c.changed {
			rows[y] = fmt.Sprintf("%s %d", c.word, y)
		}
		out.Reset()
		v := NewView(strings.Join(rows, "\n"))
		v.AltScreen = true
		r.render(v)
		if err := r.flush(false); err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(out.String(), c.word); n != len(c.changed) {
			t.Errorf("frame %d: %q written %d times, want %d: %q", i, c.word, n, len(c.changed), out.String())
		}
	}
}

// TestIndependentRow checks rows the quick read takes and rows it leaves to
// the full one, under either width method.
func TestIndependentRow(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		row  string
		want bool
	}{
		{"plain \x1b[1mbold\x1b[m tail", true},
		{"\x1b[38;2;1;2;3m│\x1b[m 世界", true},
		{"x\x1b[?25h", true},
		{"", true},
		{"\x1b[1mbold\x1b[22m", true}, // the last style is no reset: read in full
		{"x\x1b]8;;https://x.y\x07link\x1b]8;;\x07", true},
		{"x\x1b]8;;https://x.y\x1b\\link\x1b]8;;\x1b\\y", true},
		{"x\x1b]8;;\x07y", true},
		{"x\x1b]0;title\x07", true},                             // an OSC other than a link: read in full
		{"\x1b[1mbold", false},                                  // a style left open
		{"x\x1b]8;;https://x.y\x07link", false},                 // a link left open
		{"x\x1b]8;id=1;\x07y", false},                           // a link closed with params, which stay
		{"x\x1b]8;;https://x.y\x1b[1m\x07\x1b]8;;\x07y", false}, // a link cut short by a sequence
		{"x\x1b]8;;", false},                                    // a link left unterminated
		{"x\x1b_Gx\x1b\\", false},                               // a sequence waiting for a cell
		{"x\x1b[1", false},                                      // a sequence left open
		{"x\x1b[0\x1b[7m", false},                               // a sequence cut short by the next
		{"\u0301x\x1b[m", false},                                // a mark for the row above
		{"\x1b[m\u0301x", false},                                // the same, after a sequence
		{"\t\u0301x", false},                                    // the same, after a control
	} {
		for _, m := range []ansi.Method{ansi.WcWidth, ansi.GraphemeWidth} {
			if got := independentRow(c.row, m, false); got != c.want {
				t.Errorf("method %d, %q: %v, want %v", m, c.row, got, c.want)
			}
		}
	}
}

// BenchmarkCursedRenderer_changedRows flushes a 220x55 frame of styled rows:
// two rows changed each time, in the alt screen, without and with a row that
// is never independent, and every row moved by a scroll, in the alt screen
// and inline.
func BenchmarkCursedRenderer_changedRows(b *testing.B) {
	const w, h = 220, 55
	rng := rand.New(rand.NewPCG(1, 2))
	first := strings.Split(rowsFrame(rng, w, h, nil), "\n")

	two, scrolled := make([]string, 16), make([]string, 16)
	rows := slices.Clone(first)
	for i := range two {
		for _, y := range []int{rng.IntN(h), rng.IntN(h)} {
			rows[y] = fmt.Sprintf("\x1b[1m%d\x1b[m %s", i, rows[(y+1)%h])
		}
		two[i] = strings.Join(rows, "\n")
	}
	stuck := make([]string, 16) // two rows change, and a row with an image stays
	rows = slices.Clone(first)
	rows[h-3] = "\x1b_Ga=p,i=1\x1b\\" + rows[h-3]
	for i := range stuck {
		for _, y := range []int{rng.IntN(h - 4), rng.IntN(h - 4)} {
			rows[y] = fmt.Sprintf("\x1b[1m%d\x1b[m %s", i, rows[(y+1)%h])
		}
		stuck[i] = strings.Join(rows, "\n")
	}
	rows = slices.Clone(first)
	for i := range scrolled {
		rows = append(rows[1:], fmt.Sprintf("\x1b[1m%d\x1b[m %s", i, rows[0]))
		scrolled[i] = strings.Join(rows, "\n")
	}

	for _, c := range []struct {
		name   string
		frames []string
		alt    bool
	}{{"two rows", two, true}, {"two rows, an image kept", stuck, true}, {"scroll", scrolled, true}, {"scroll inline", scrolled, false}} {
		b.Run(c.name, func(b *testing.B) {
			r := newCursedRenderer(io.Discard, []string{"TERM=xterm-256color"}, w, h)
			r.start()
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				v := NewView(c.frames[i%len(c.frames)])
				v.AltScreen = c.alt
				r.render(v)
				if err := r.flush(false); err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	}
}
