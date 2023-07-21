package tea

import (
	"strings"

	"github.com/muesli/termenv"
	"github.com/samber/lo"
)

func ctrlSeq(code string) string {
	return termenv.CSI + code + "m"
}

// FrameBuffer is a view of the terminal to render to
type FrameBuffer struct {
	B []rune
	// OPTIMIZE: store ranges of colors instead of color for every pixel
	backgrounds []string
	foregrounds []string
	Height      int
	Width       int
}

// NewFramebuffer creates a new Framebuffer
func NewFramebuffer(height, width int) FrameBuffer {
	buf := make([]rune, height*width)
	for i := range buf {
		buf[i] = ' '
	}
	return FrameBuffer{
		B:           buf,
		Height:      height,
		Width:       width,
		backgrounds: make([]string, height*width),
		foregrounds: make([]string, height*width),
	}
}

// Set writes a rune to the framebuffer to the given position
func (fb FrameBuffer) Set(y, x int, c rune) {
	// TODO: bounds check?
	// if y*fb.Width+x >= len(fb.B) {
	// 	return
	// }
	fb.B[y*fb.Width+x] = c
}

// Background colors y'th row bacground to given color from x1 to x2
func (fb FrameBuffer) Background(y, x1, x2 int, background termenv.Color) {
	for x := x1; x < x2; x++ {
		fb.backgrounds[y*fb.Width+x] = background.Sequence(true)
	}
}

// Background colors y'th row foreground to given color from x1 to x2
func (fb FrameBuffer) Foreground(y, x1, x2 int, foreground termenv.Color) {
	for x := x1; x < x2; x++ {
		fb.foregrounds[y*fb.Width+x] = foreground.Sequence(false)
	}
}

// Render framebuffer to string
func (fb FrameBuffer) Render() string {
	// OPTIMIZE: strings.Builder
	rows := make([]string, fb.Height)
	bg := ""
	fg := ""
	for y := 0; y < fb.Height; y++ {
		fullRow := fb.B[y*fb.Width : (y+1)*fb.Width]
		newRow := ""
		for x := 0; x < fb.Width; x++ {
			coloring := ""
			if fb.backgrounds[y*fb.Width+x] != bg || fb.foregrounds[y*fb.Width+x] != fg {
				bg = fb.backgrounds[y*fb.Width+x]
				fg = fb.foregrounds[y*fb.Width+x]
				coloring = ctrlSeq(termenv.ResetSeq) + lo.
					Switch[bool, string](true).
					Case(bg == "" && fg == "", "").
					Case(bg == "", ctrlSeq(fg)).
					Case(fg == "", ctrlSeq(bg)).
					Default(ctrlSeq(bg+";"+fg))
			}
			newRow += coloring + string([]rune{fullRow[x]})
		}

		rows[y] = newRow
	}
	return strings.Join(rows, "\n")
}

// TODO: is it needed? may be it is enough to use FrameBuffer as viewbox?
// Viewbox is rectangular view to terminal frame part
type Viewbox struct {
	FB     FrameBuffer
	Y      int
	X      int
	Height int
	Width  int
}

// Row returns view to current viewbox's row
func (vb Viewbox) Row(y int) Viewbox {
	return Viewbox{
		Y:      y + vb.Y,
		X:      vb.X,
		Height: 1,
		Width:  vb.Width,
		FB:     vb.FB,
	}
}

// PaddingOptions is padding options
type PaddingOptions struct {
	Top, Bottom, Left, Right int
}

// Padding returns view to current viewbox inner with given paddings and size
func (vb Viewbox) Padding(opt PaddingOptions) Viewbox {
	// TODO: check bounds
	return Viewbox{
		Y:      vb.Y + opt.Top,
		X:      vb.X + opt.Left,
		Height: vb.Height - opt.Top - opt.Bottom,
		Width:  vb.Width - opt.Left - opt.Right,
		FB:     vb.FB,
	}
}

// Set writes a rune to the framebuffer in position relative to viewbox
func (vb Viewbox) Set(y, x int, c rune) {
	// TODO: bounds check?
	vb.FB.Set(vb.Y+y, vb.X+x, c)
}

// Background colors y'th row bacground to given color from x1 to x2 with
// coordinates relative to viewbox
func (vb Viewbox) Background(y, x1, x2 int, background termenv.Color) {
	vb.FB.Background(y+vb.Y, x1+vb.X, x2+vb.X, background)
}

// Background colors y'th row foreground to given color from x1 to x2 with
// coordinates relative to viewbox
func (vb Viewbox) Foreground(y, x1, x2 int, foreground termenv.Color) {
	vb.FB.Foreground(y+vb.Y, x1+vb.X, x2+vb.X, foreground)
}
