package tea

import (
	"fmt"
	"io"

	te "github.com/muesli/termenv"
)

func clearLine(w io.Writer) {
	fmt.Fprintf(w, te.CSI+te.EraseLineSeq, 2)
}

func cursorUp(w io.Writer) {
	fmt.Fprintf(w, te.CSI+te.CursorUpSeq, 1)
}

func cursorDown(w io.Writer) {
	fmt.Fprintf(w, te.CSI+te.CursorDownSeq, 1)
}

func insertLine(w io.Writer, numLines int) {
	fmt.Fprintf(w, te.CSI+"%dL", numLines)
}

func moveCursor(w io.Writer, row, col int) {
	fmt.Fprintf(w, te.CSI+te.CursorPositionSeq, row, col)
}

func saveCursorPosition(w io.Writer) {
	fmt.Fprint(w, te.CSI+"s")
}

func restoreCursorPosition(w io.Writer) {
	fmt.Fprint(w, te.CSI+"u")
}

func changeScrollingRegion(w io.Writer, top, bottom int) {
	fmt.Fprintf(w, te.CSI+"%d;%dr", top, bottom)
}
