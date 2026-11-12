package tea

import (
	"bytes"
	"io"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/charmbracelet/x/vt"
)

// TestCursedRenderer_changedLinesMatchFullRedraw: drawing only the changed
// lines leaves the terminal showing exactly what a full redraw leaves, frame
// after frame (typing, scrolling output, repeated and blank lines, wide
// characters, resizes), and writes no more in total. Both outputs are played
// into a terminal emulator and the screens (and shown cursors) compared.
// Bytes may differ: a full redraw can re-send lines that didn't change.
func TestCursedRenderer_changedLinesMatchFullRedraw(t *testing.T) {
	t.Parallel()
	vocab := []string{
		"", "", "hello", "\x1b[1mbold\x1b[m text", "\x1b[38;2;10;200;30mgreen\x1b[m",
		"世界 wide", "emoji 🎉 here", "│ box │", "────────", "a", "b", "c",
	}
	const w, h = 24, 12
	emulator := func() *vt.Emulator {
		e := vt.NewEmulator(w, h)
		done := make(chan struct{})
		go func() { _, _ = io.Copy(io.Discard, e); close(done) }() // replies to queries
		t.Cleanup(func() {
			if c, ok := e.InputPipe().(io.Closer); ok {
				_ = c.Close()
			}
			<-done
		})
		return e
	}
	for seed := range uint64(50) {
		rnd := rand.New(rand.NewPCG(seed, 1))
		var outA, outB bytes.Buffer
		a := newCursedRenderer(&outA, []string{"TERM=xterm-256color"}, w, h)
		b := newCursedRenderer(&outB, []string{"TERM=xterm-256color"}, w, h)
		b.fullRedraw = true
		ea, eb := emulator(), emulator()
		a.start()
		b.start()
		var bytesA, bytesB int
		lines := make([]string, h)
		for i := range lines {
			lines[i] = vocab[rnd.IntN(len(vocab))]
		}
		cw, ch := w, h
		for step := range 300 {
			if rnd.IntN(30) == 0 { // a resize; a shorter screen than the frame clips it
				cw, ch = w-rnd.IntN(8), h-rnd.IntN(6)
				a.resize(cw, ch)
				b.resize(cw, ch)
				ea.Resize(cw, ch)
				eb.Resize(cw, ch)
			}
			switch rnd.IntN(5) {
			case 0: // type into one line
				i := rnd.IntN(h)
				lines[i] += string(rune('a' + rnd.IntN(26)))
				if len(lines[i]) > w-2 {
					lines[i] = ""
				}
			case 1: // scroll the middle up by one or more lines
				n := 1 + rnd.IntN(3)
				mid := lines[1 : h-2]
				copy(mid, mid[min(n, len(mid)):])
				for i := max(0, len(mid)-n); i < len(mid); i++ {
					mid[i] = vocab[rnd.IntN(len(vocab))]
				}
			case 2: // replace a line
				lines[rnd.IntN(h)] = vocab[rnd.IntN(len(vocab))]
			case 3: // replace many lines
				for range 1 + rnd.IntN(h) {
					lines[rnd.IntN(h)] = vocab[rnd.IntN(len(vocab))]
				}
			case 4: // nothing changes
			}
			v := NewView(strings.Join(lines, "\n"))
			v.AltScreen = seed%2 == 0 // odd seeds draw inline, below the prompt
			if rnd.IntN(2) == 0 {     // a shown cursor, like a text input's
				v.Cursor = NewCursor(rnd.IntN(cw), rnd.IntN(ch))
			}
			for _, r := range []*cursedRenderer{a, b} {
				r.render(v)
				if err := r.flush(false); err != nil {
					t.Fatal(err)
				}
			}
			bytesA += outA.Len()
			bytesB += outB.Len()
			_, _ = ea.Write(outA.Bytes())
			_, _ = eb.Write(outB.Bytes())
			// A hidden cursor is wherever the last write left it: only a shown
			// one must match.
			if sa, sb := ea.Render(), eb.Render(); sa != sb || (v.Cursor != nil && ea.CursorPosition() != eb.CursorPosition()) {
				t.Fatalf("seed %d step %d: the screen after drawing changed lines differs from a full redraw\nchanged (cursor %v):\n%s\nfull (cursor %v):\n%s\nlast output, changed: %q\nlast output, full:    %q",
					seed, step, ea.CursorPosition(), sa, eb.CursorPosition(), sb, outA.String(), outB.String())
			}
			outA.Reset()
			outB.Reset()
		}
		if bytesA > bytesB {
			t.Errorf("seed %d: drawing changed lines wrote %d bytes, a full redraw %d", seed, bytesA, bytesB)
		}
	}
}

// BenchmarkCursedRenderer_keystroke: one typed character on a full 120×40
// screen, from render to the bytes written. The full redraw re-diffs every
// line; drawing changed lines diffs one.
func BenchmarkCursedRenderer_keystroke(b *testing.B) {
	const w, h = 120, 40
	lines := make([]string, h)
	for i := range lines {
		lines[i] = strings.Repeat("\x1b[38;2;120;160;200mlorem\x1b[m ipsum ", 7)
	}
	for _, full := range []bool{true, false} {
		name := "changed-lines"
		if full {
			name = "full-redraw"
		}
		b.Run(name, func(b *testing.B) {
			var out bytes.Buffer
			r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, w, h)
			r.fullRedraw = full
			r.start()
			typed := ""
			b.ReportAllocs()
			for i := 0; b.Loop(); i++ {
				if len(typed) > 60 {
					typed = ""
				}
				typed += string(rune('a' + i%26))
				lines[h-1] = "> " + typed
				v := NewView(strings.Join(lines, "\n"))
				v.AltScreen = true
				r.render(v)
				if err := r.flush(false); err != nil {
					b.Fatal(err)
				}
				out.Reset()
			}
		})
	}
}
