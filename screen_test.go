package tea

import (
	"bytes"
	"io"
	"testing"
)

func TestScreen(t *testing.T) {
	exercise := func(t *testing.T, fn func(io.Writer), expect []byte) {
		var w bytes.Buffer
		fn(&w)
		if !bytes.Equal(w.Bytes(), expect) {
			t.Errorf("expected %q, got %q", expect, w.Bytes())
		}
	}

	t.Run("change scrolling region", func(t *testing.T) {
		exercise(t, func(w io.Writer) {
			changeScrollingRegion(w, 16, 22)
		}, []byte("\x1b[16;22r"))
	})

	t.Run("line", func(t *testing.T) {
		t.Run("clear", func(t *testing.T) {
			exercise(t, clearLine, []byte("\x1b[2K"))
		})

		t.Run("insert", func(t *testing.T) {
			exercise(t, func(w io.Writer) {
				insertLine(w, 12)
			}, []byte("\x1b[12L"))
		})
	})

	t.Run("cursor", func(t *testing.T) {
		t.Run("hide", func(t *testing.T) {
			exercise(t, hideCursor, []byte("\x1b[?25l"))
		})

		t.Run("show", func(t *testing.T) {
			exercise(t, showCursor, []byte("\x1b[?25h"))
		})

		t.Run("up", func(t *testing.T) {
			exercise(t, cursorUp, []byte("\x1b[1A"))
		})

		t.Run("down", func(t *testing.T) {
			exercise(t, func(w io.Writer) {
				cursorDownBy(w, 3)
			}, []byte("\x1b[3B"))
		})

		t.Run("upBy", func(t *testing.T) {
			exercise(t, func(w io.Writer) {
				cursorUpBy(w, 3)
			}, []byte("\x1b[3A"))
		})

		t.Run("downBy", func(t *testing.T) {
			exercise(t, cursorDown, []byte("\x1b[1B"))
		})

		t.Run("move", func(t *testing.T) {
			exercise(t, func(w io.Writer) {
				moveCursor(w, 10, 20)
			}, []byte("\x1b[10;20H"))
		})

		t.Run("back", func(t *testing.T) {
			exercise(t, func(w io.Writer) {
				cursorBack(w, 15)
			}, []byte("\x1b[15D"))
		})
	})

}
