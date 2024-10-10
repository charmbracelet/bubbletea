package tea

import (
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

// cellRenderer is a cell-based terminal renderer.
type cellRenderer struct {
	mtx sync.Mutex
	out io.Writer

	inl *cellbuf.Screen
	alt *cellbuf.Screen

	scr *cellbuf.Screen // The current focused screen

	width      int
	lastHeight int // The height of the last render
	method     cellbuf.WidthMethod

	// modes
	altScreenMode bool
	hideCursor    bool
}

func newCellRenderer() *cellRenderer {
	r := &cellRenderer{}
	r.reset()
	return r
}

var _ renderer = &cellRenderer{}

// close implements renderer.
func (c *cellRenderer) close() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	var seq string
	_, y := c.scr.Pos()

	if !c.altScreenMode && y < c.lastHeight {
		diff := c.lastHeight - y - 1
		// Ensure the cursor is at the bottom of the screen
		seq += strings.Repeat("\n", diff)
	}

	_, err := io.WriteString(c.out, seq+ansi.EraseEntireLine+"\r")
	return err
}

// flush implements renderer.
func (c *cellRenderer) flush() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	seqs := c.scr.Commit()
	if seqs == "" {
		return nil
	}

	_, err := c.out.Write([]byte(seqs))
	return err
}

// render implements renderer.
func (c *cellRenderer) render(s string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.lastHeight = strings.Count(s, "\n") + 1
	c.scr.SetContent(s)
}

// reset implements renderer.
func (c *cellRenderer) reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.inl = cellbuf.NewScreen(c.width, c.method)
	c.alt = cellbuf.NewAltScreen(c.width, c.method)
	if c.altScreenMode {
		c.scr = c.alt
	} else {
		c.scr = c.inl
	}
}

// update implements renderer.
func (c *cellRenderer) update(msg Msg) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	switch msg := msg.(type) {
	case rendererWriter:
		c.out = msg.Writer

	case WindowSizeMsg:
		c.inl.SetWidth(msg.Width)
		c.alt.SetWidth(msg.Width)
		c.inl.Repaint()
		c.alt.Repaint()

	case clearScreenMsg:
		io.WriteString(c.out, ansi.EraseEntireDisplay+ansi.MoveCursorOrigin) //nolint:errcheck

	case repaintMsg:
		c.scr.Repaint()

	case printLineMessage:
		c.scr.InsertAbove(msg.messageBody)

	case enableModeMsg:
		switch string(msg) {
		case ansi.AltScreenBufferMode:
			if c.altScreenMode {
				return
			}

			c.scr = c.alt
			c.altScreenMode = true

			// NOTE: Using `CSI ? 1049` clears the screen so we need to repaint
			// the alt screen buffer.
			c.alt.Repaint()

		case ansi.CursorVisibilityMode:
			if !c.hideCursor {
				return
			}

			c.hideCursor = false
		}

	case disableModeMsg:
		switch string(msg) {
		case ansi.AltScreenBufferMode:
			if !c.altScreenMode {
				return
			}

			c.scr = c.inl
			c.altScreenMode = false

		case ansi.CursorVisibilityMode:
			if c.hideCursor {
				return
			}

			c.hideCursor = true
		}
	}
}
