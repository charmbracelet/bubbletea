package tea

import (
	"fmt"
	"io"
	"math/rand/v2"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// rowsFrame is a full-screen frame of styled rows, each closing its styles
// and hyperlinks, with wide runes in some; the rows in dirty end open.
func rowsFrame(rng *rand.Rand, w, h int, dirty map[int]bool) string {
	words := []string{"plain", "\x1b[1mbold\x1b[m", "\x1b[38;5;196mred\x1b[m", "世界", "\x1b]8;;https://x.y\x07link\x1b]8;;\x07", "😀", "\x1b[48;2;10;20;30m bg \x1b[m"}
	rows := make([]string, h)
	for y := range rows {
		var b strings.Builder
		for n := 0; n < w; n += 6 {
			b.WriteString(words[rng.IntN(len(words))])
			b.WriteByte(' ')
		}
		if dirty[y] {
			b.WriteString("\x1b[7mopen")
		} else {
			b.WriteString("\x1b[m")
		}
		rows[y] = b.String()
	}
	return strings.Join(rows, "\n")
}

// TestCursedRenderer_changedRows: a frame that redraws only its changed rows
// leaves the buffer as a frame drawn whole does, whether its rows can be
// drawn apart or not, and when they scrolled.
func TestCursedRenderer_changedRows(t *testing.T) {
	t.Parallel()

	const w, h = 60, 12
	rng := rand.New(rand.NewPCG(1, 2))
	r := newCursedRenderer(io.Discard, []string{"TERM=xterm-256color"}, w, h)
	r.start()

	rows := strings.Split(rowsFrame(rng, w, h, nil), "\n")
	for i := range 200 {
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
		uv.NewStyledString(v.Content).Draw(want, want.Bounds())
		for y := range h {
			if got, exp := r.cellbuf.Line(y).Render(), want.Line(y).Render(); got != exp {
				t.Fatalf("frame %d, row %d:\n got %q\nwant %q", i, y, got, exp)
			}
		}
	}
}

// BenchmarkCursedRenderer_changedRows flushes a 220x55 frame of styled rows,
// two of which change each time.
func BenchmarkCursedRenderer_changedRows(b *testing.B) {
	const w, h = 220, 55
	rng := rand.New(rand.NewPCG(1, 2))
	r := newCursedRenderer(io.Discard, []string{"TERM=xterm-256color"}, w, h)
	r.start()

	rows := strings.Split(rowsFrame(rng, w, h, nil), "\n")
	frames := make([]string, 16)
	for i := range frames {
		for _, y := range []int{rng.IntN(h), rng.IntN(h)} {
			rows[y] = fmt.Sprintf("\x1b[1m%d\x1b[m %s", i, rows[(y+1)%h])
		}
		frames[i] = strings.Join(rows, "\n")
	}

	b.ReportAllocs()
	i := 0
	for b.Loop() {
		v := NewView(frames[i%len(frames)])
		v.AltScreen = true
		r.render(v)
		if err := r.flush(false); err != nil {
			b.Fatal(err)
		}
		i++
	}
}
