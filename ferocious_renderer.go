package tea

import (
	"bytes"
	"image"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

var undefPoint = image.Pt(-1, -1)

// cursor represents a terminal cursor.
type cursor struct {
	image.Point
	visible bool
}

// screen represents a terminal screen.
type screen struct {
	dirty          map[int]int // keeps track of dirty cells
	linew          []int       // keeps track of the width of each line
	cellbuf.Buffer             // the cell buffer
	cur            cursor      // cursor state
}

// isDirty returns true if the cell at the given position is dirty.
func (s *screen) isDirty(x, y int) bool {
	idx := y*s.Width() + x
	v, ok := s.dirty[idx]
	return ok && v == 1
}

// reset resets the screen to its initial state.
func (s *screen) reset() {
	s.Buffer = cellbuf.Buffer{}
	s.dirty = make(map[int]int)
	s.cur = cursor{}
	s.linew = make([]int, 0)
}

// Set implements [cellbuf.Grid] and marks changed cells as dirty.
func (s *screen) SetCell(x, y int, cell cellbuf.Cell) (v bool) {
	c, ok := s.Cell(x, y)
	if !ok {
		return
	}

	if c.Equal(cell) {
		// Cells are the same, no need to update.
		return
	}

	v = s.Buffer.SetCell(x, y, cell)
	if v {
		// Mark the cell as dirty. You nasty one ;)
		idx := y*s.Width() + x
		s.dirty[idx] = 1
	}

	return
}

// ferociousRenderer is a cell-based terminal renderer. It's ferocious!
type ferociousRenderer struct {
	mtx sync.Mutex
	out io.Writer    // we only write to the output during flush and close
	buf bytes.Buffer // the internal buffer for rendering

	scrs [2]screen // Both inline and alt-screen
	scr  *screen   // Points to the current used screen

	method cellbuf.Method

	finalCur image.Point // The final cursor position

	pen  cellbuf.Style
	link cellbuf.Link

	queueAbove  []string
	lastRenders [2]string // The last render for both inline and alt-screen buffers
	lastRender  *string   // Points to the last render string
	frame       string    // The current frame to render
	lastHeight  int       // The height of the last render

	// modes
	altScreen    bool
	cursorHidden bool

	profile colorprofile.Profile
}

func newFerociousRenderer(p colorprofile.Profile) *ferociousRenderer {
	r := &ferociousRenderer{
		// TODO: Update this if Grapheme Clustering is supported.
		method:   cellbuf.WcWidth,
		finalCur: undefPoint,
		profile:  p,
	}
	r.reset()
	return r
}

var _ renderer = &ferociousRenderer{}

// close implements renderer.
func (c *ferociousRenderer) close() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	seq := c.buf.String()
	c.buf.Reset()

	y := c.scr.cur.Y
	if !c.altScreen && y < c.lastHeight {
		diff := c.lastHeight - y - 1
		// Ensure the cursor is at the bottom of the screen
		seq += strings.Repeat("\n", diff)
		y += diff
		c.scr.cur.Y = y
	}

	if c.scr.cur.X != 0 {
		seq += "\r"
		c.scr.cur.X = 0
	}
	if _, line := cellbuf.RenderLineWithProfile(c.scr, y, c.profile); line != "" {
		// OPTIM: We only clear the line if there's content on it.
		seq += ansi.EraseEntireLine
	}

	if seq == "" {
		// Nothing to clear.
		return nil
	}

	_, err := io.WriteString(c.out, seq)
	return err
}

// clearScreen returns a string to clear the screen and moves the cursor to the
// origin location i.e. top-left.
func (c *ferociousRenderer) clearScreen() {
	c.moveCursor(0, 0)
	if c.altScreen {
		c.buf.WriteString(ansi.EraseEntireScreen) //nolint:errcheck
		return
	}

	c.buf.WriteString(ansi.EraseScreenBelow) //nolint:errcheck
}

// flush implements renderer.
func (c *ferociousRenderer) flush() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.finalCur == c.scr.cur.Point && len(c.queueAbove) == 0 &&
		c.frame == *c.lastRender && c.lastHeight == c.scr.Height() {
		return nil
	}

	if !c.altScreen && len(c.queueAbove) > 0 {
		c.moveCursor(0, 0)
		for _, line := range c.queueAbove {
			c.buf.WriteString(line + ansi.EraseLineRight + "\r\n")
		}
		c.queueAbove = c.queueAbove[:0]
		c.repaint()
	}

	if *c.lastRender == "" {
		// First render and repaints clear the screen.
		c.clearScreen()
	}

	if c.scr.cur.X > c.scr.Width()-1 {
		// When the cursor is at EOL, we need to put it back to the beginning
		// of line. Otherwise, the autowrap (DECAWM), which is enabled by
		// default, will move the cursor to the next line on the next cell
		// write.
		c.buf.WriteByte(ansi.CR)
		c.scr.cur.X = 0
	}

	c.changes()

	// XXX: We need to move the cursor to the final position before rendering
	// the frame to avoid flickering.
	shouldHideCursor := !c.cursorHidden
	if c.finalCur != image.Pt(-1, -1) {
		shouldMove := c.finalCur != c.scr.cur.Point
		shouldHideCursor = shouldHideCursor && shouldMove
		if shouldMove {
			c.moveCursor(c.finalCur.X, c.finalCur.Y)
		}
	}

	c.scr.dirty = make(map[int]int)
	c.lastHeight = cellbuf.Height(c.frame)
	*c.lastRender = c.frame
	render := c.buf.String()
	c.buf.Reset()
	if render == "" {
		return nil
	}

	if shouldHideCursor {
		// Hide the cursor while rendering to avoid flickering.
		render = ansi.HideCursor + render + ansi.ShowCursor
	}

	_, err := io.WriteString(c.out, render)
	return err
}

// render implements renderer.
func (c *ferociousRenderer) render(s string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.frame = s
	// Ensure the buffer is at least the height of the new frame.
	height := cellbuf.Height(s)
	c.scr.Resize(c.scr.Width(), height)
	linew := cellbuf.SetContent(c.scr, c.method, s)
	c.scr.linew = linew
}

// reset implements renderer.
func (c *ferociousRenderer) reset() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.lastRenders[0] = ""
	c.lastRenders[1] = ""
	c.scrs[0].reset()
	c.scrs[1].reset()
	// alt-screen buffer cursor always starts from where the main buffer cursor
	// is. We need to set it to (-1,-1) to force the cursor to be moved to the
	// origin on the first render.
	c.scrs[1].cur.Point = undefPoint
	if c.altScreen {
		c.scr = &c.scrs[1]
		c.lastRender = &c.lastRenders[1]
	} else {
		c.scr = &c.scrs[0]
		c.lastRender = &c.lastRenders[0]
	}
}

// repaint forces a repaint of the screen.
func (c *ferociousRenderer) repaint() {
	*c.lastRender = ""
}

// updateCursorVisibility ensures the cursor state is in sync with the
// renderer.
func (c *ferociousRenderer) updateCursorVisibility() {
	if !c.cursorHidden != c.scr.cur.visible {
		c.scr.cur.visible = !c.cursorHidden
		// cmd.exe and other terminals keep separate cursor states for the AltScreen
		// and the main buffer. We have to explicitly reset the cursor visibility
		// whenever we exit AltScreen.
		if c.cursorHidden {
			io.WriteString(&c.buf, ansi.HideCursor) //nolint:errcheck
		} else {
			io.WriteString(&c.buf, ansi.ShowCursor) //nolint:errcheck
		}
	}
}

// update implements renderer.
func (c *ferociousRenderer) update(msg Msg) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	switch msg := msg.(type) {
	case ColorProfileMsg:
		c.profile = msg.Profile

	case rendererWriter:
		c.out = msg.Writer

	case WindowSizeMsg:
		c.scrs[0].Resize(msg.Width, msg.Height)
		c.scrs[1].Resize(msg.Width, msg.Height)
		c.lastRenders[0] = ""
		c.lastRenders[1] = ""

	case clearScreenMsg:
		seq := ansi.EraseEntireScreen + ansi.HomeCursorPosition
		if !c.cursorHidden {
			seq = ansi.HideCursor + seq + ansi.ShowCursor
		}

		io.WriteString(c.out, seq) //nolint:errcheck
		c.repaint()

	case repaintMsg:
		c.repaint()

	case printLineMessage:
		if !c.altScreen {
			c.queueAbove = append(c.queueAbove, strings.Split(msg.messageBody, "\n")...)
		}

	case enableModeMsg:
		switch string(msg) {
		case ansi.AltScreenBufferMode.String():
			if c.altScreen {
				return
			}

			c.scr = &c.scrs[1]
			c.altScreen = true

			// NOTE: Using `CSI ? 1049` clears the screen so we need to repaint
			// the alt screen buffer.
			c.repaint()

			// Some terminals keep separate cursor states for the AltScreen and
			// the main buffer. We have to explicitly reset the cursor visibility
			// whenever we enter or leave AltScreen.
			c.updateCursorVisibility()

		case ansi.CursorEnableMode.String():
			if !c.cursorHidden {
				return
			}

			c.cursorHidden = false
		}

	case disableModeMsg:
		switch string(msg) {
		case ansi.AltScreenBufferMode.String():
			if !c.altScreen {
				return
			}

			c.scr = &c.scrs[0]
			c.altScreen = false

			// Some terminals keep separate cursor states for the AltScreen and
			// the main buffer. We have to explicitly reset the cursor visibility
			// whenever we enter or leave AltScreen.
			c.updateCursorVisibility()

		case ansi.CursorEnableMode.String():
			if c.cursorHidden {
				return
			}

			c.cursorHidden = true
		}

	case setCursorPosMsg:
		c.finalCur = image.Pt(clamp(msg.X, 0, c.scr.Width()-1), clamp(msg.Y, 0, c.scr.Height()-1))
	}
}

var spaceCell = cellbuf.Cell{Content: " ", Width: 1}

// changes commits the changes from the cell buffer using the dirty cells map
// and writes them to the internal buffer.
func (c *ferociousRenderer) changes() {
	width := c.scr.Width()
	if width <= 0 {
		return
	}

	height := c.scr.Height()
	if *c.lastRender == "" {
		// We render the changes line by line to be able to get the cursor
		// position using the width of each line.
		var x int
		for y := 0; y < height; y++ {
			var line string
			x, line = cellbuf.RenderLineWithProfile(c.scr, y, c.profile)
			c.buf.WriteString(line)
			if y < height-1 {
				x = 0
				c.buf.WriteString("\r\n")
			}
		}

		c.scr.cur.X, c.scr.cur.Y = x, height-1
		return
	}

	// TODO: iterate over the dirty cells instead of the whole buffer.
	// TODO: optimize continuous space-only segments i.e. concatenate them to
	// erase the line instead of using spaces to erase the line.
	for y := 0; y < height; y++ {
		var seg *cellbuf.Segment
		var segX int    // The start position of the current segment.
		var eraser bool // Whether we're erasing using spaces and no styles or links.
		for x := 0; x < width; x++ {
			cell, ok := c.scr.Cell(x, y)
			if !ok || cell.Width == 0 {
				continue
			}

			// Convert the cell to respect the current color profile.
			cell.Style = cell.Style.Convert(c.profile)
			cell.Link = cell.Link.Convert(c.profile)

			if !c.scr.isDirty(x, y) {
				if seg != nil {
					erased := c.flushSegment(seg, image.Pt(segX, y), eraser)
					seg = nil
					if erased {
						// If the segment erased the rest of the line, we don't need to
						// render the rest of the line.
						break
					}
				}
				continue
			}

			if seg == nil {
				segX = x
				eraser = cell.Equal(spaceCell)
				seg = &cellbuf.Segment{
					Style:   cell.Style,
					Link:    cell.Link,
					Content: cell.Content,
					Width:   cell.Width,
				}
				continue
			}

			if !seg.Style.Equal(cell.Style) || seg.Link != cell.Link {
				erased := c.flushSegment(seg, image.Pt(segX, y), eraser)
				if erased {
					seg = nil
					// If the segment erased the rest of the line, we don't need to
					// render the rest of the line.
					break
				}
				segX = x
				eraser = cell.Equal(spaceCell)
				seg = &cellbuf.Segment{
					Style:   cell.Style,
					Link:    cell.Link,
					Content: cell.Content,
					Width:   cell.Width,
				}
				continue
			}

			eraser = eraser && cell.Equal(spaceCell)
			seg.Content += cell.Content
			seg.Width += cell.Width
		}

		if seg != nil {
			c.flushSegment(seg, image.Pt(segX, y), eraser)
			seg = nil
		}
	}

	// Reset the style and hyperlink if necessary.
	if c.link.URL != "" {
		c.buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
		c.link.Reset()
	}
	if !c.pen.Empty() {
		c.buf.WriteString(ansi.ResetStyle) //nolint:errcheck
		c.pen.Reset()
	}

	// Delete extra lines from previous render.
	if c.lastHeight > height {
		// Move the cursor to the last line of this render and erase the rest
		// of the screen.
		c.moveCursor(c.scr.cur.X, height-1)
		c.buf.WriteString(ansi.EraseScreenBelow)
	}
}

// flushSegment flushes the segment to the buffer. It returns true if the
// segment the rest of the line was erased.
func (c *ferociousRenderer) flushSegment(seg *cellbuf.Segment, to image.Point, eraser bool) (erased bool) {
	if c.scr.cur.Point != to {
		c.renderReset(seg)
		c.moveCursor(to.X, to.Y)
	}

	// We use [ansi.EraseLineRight] to erase the rest of the line if the segment
	// is an "eraser" i.e. it's just a bunch of spaces with no styles or links. We erase the
	// rest of the line when:
	// 1. The segment is an eraser.
	// 2. The segment reaches the end of the line to erase i.e. the new line is shorter.
	// 3. The segment takes more bytes than [ansi.EraseLineRight] to erase which is 4 bytes.
	if eraser && to.Y < len(c.scr.linew) && seg.Width > 4 && (c.scr.linew)[to.Y] < seg.Width+to.X {
		c.renderReset(seg)
		c.buf.WriteString(ansi.EraseLineRight) //nolint:errcheck
		erased = true
	} else {
		c.renderSegment(seg)
	}
	return
}

func (c *ferociousRenderer) renderReset(seg *cellbuf.Segment) {
	if seg.Link != c.link && c.link.URL != "" {
		c.buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
		c.link.Reset()
	}
	if seg.Style.Empty() && !c.pen.Empty() {
		c.buf.WriteString(ansi.ResetStyle) //nolint:errcheck
		c.pen.Reset()
	}
}

func (c *ferociousRenderer) renderSegment(seg *cellbuf.Segment) {
	isSpaces := strings.Trim(seg.Content, " ") == "" && c.pen.Empty() && seg.Style.Empty()
	if !isSpaces && !seg.Style.Equal(c.pen) {
		// We don't apply the style if the content is spaces. It's more efficient
		// to just write the spaces.
		c.buf.WriteString(seg.Style.DiffSequence(c.pen)) // nolint:errcheck
		c.pen = seg.Style
	}
	if seg.Link != c.link {
		c.buf.WriteString(ansi.SetHyperlink(seg.Link.URL, seg.Link.URLID)) // nolint:errcheck
		c.link = seg.Link
	}

	c.buf.WriteString(seg.Content)
	c.scr.cur.X += seg.Width

	if c.scr.cur.X >= c.scr.Width() {
		// NOTE: We need to reset the cursor when at phantom cell i.e. outside
		// the screen, otherwise, the cursor position will be out of sync.
		c.scr.cur.X = 0
		c.buf.WriteByte(ansi.CR)
	}
}

// moveCursor moves the cursor to the given position.
func (c *ferociousRenderer) moveCursor(x, y int) {
	if c.scr.cur.X == x && c.scr.cur.Y == y {
		return
	}

	if c.altScreen {
		// TODO: Optimize for small movements i.e. movements that cost less
		// than 8 bytes in total. [ansi.MoveCursor] is at least 6 bytes long.
		c.buf.WriteString(ansi.SetCursorPosition(x+1, y+1))
	} else {
		if c.scr.cur.X < x {
			dx := x - c.scr.cur.X
			switch dx {
			case 1:
				// OPTIM: We write the cell content under the cursor if it's the same
				// style and link. This is more efficient than moving the cursor which
				// costs at least 3 bytes [ansi.CursorRight].
				cell, ok := c.scr.Cell(c.scr.cur.X, c.scr.cur.Y)
				if ok &&
					(cell.Style.Equal(c.pen) && cell.Link == c.link) {
					c.buf.WriteString(cell.Content)
					break
				}
				fallthrough
			default:
				c.buf.WriteString(ansi.CursorRight(dx))
			}
		} else if c.scr.cur.X > x {
			if x == 0 {
				// We use [ansi.CR] instead of [ansi.CursorLeft] to avoid
				// writing multiple bytes.
				c.buf.WriteByte(ansi.CR)
			} else {
				dx := c.scr.cur.X - x
				if dx >= 3 {
					// [ansi.CursorLeft] is at least 3 bytes long, so we use [ansi.BS]
					// when we can to avoid writing more bytes than necessary.
					c.buf.WriteString(ansi.CursorLeft(dx))
				} else {
					c.buf.WriteString(strings.Repeat("\b", dx))
				}
			}
		}
		if c.scr.cur.Y < y {
			dy := y - c.scr.cur.Y
			if dy >= 3 {
				// [ansi.CursorDown] is at least 3 bytes long, so we use "\n" when
				// we can to avoid writing more bytes than necessary.
				c.buf.WriteString(ansi.CursorDown(dy))
			} else {
				c.buf.WriteString(strings.Repeat("\n", dy))
			}
		} else if c.scr.cur.Y > y {
			dy := c.scr.cur.Y - y
			c.buf.WriteString(ansi.CursorUp(dy))
		}
	}

	c.scr.cur.X, c.scr.cur.Y = x, y
}
