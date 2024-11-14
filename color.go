package tea

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
)

// backgroundColorMsg is a message that requests the terminal background color.
type backgroundColorMsg struct{}

// RequestBackgroundColor is a command that requests the terminal background color.
func RequestBackgroundColor() Msg {
	return backgroundColorMsg{}
}

// foregroundColorMsg is a message that requests the terminal foreground color.
type foregroundColorMsg struct{}

// RequestForegroundColor is a command that requests the terminal foreground color.
func RequestForegroundColor() Msg {
	return foregroundColorMsg{}
}

// cursorColorMsg is a message that requests the terminal cursor color.
type cursorColorMsg struct{}

// RequestCursorColor is a command that requests the terminal cursor color.
func RequestCursorColor() Msg {
	return cursorColorMsg{}
}

// setBackgroundColorMsg is a message that sets the terminal background color.
type setBackgroundColorMsg struct{ color.Color }

// SetBackgroundColor is a command that sets the terminal background color.
func SetBackgroundColor(c color.Color) Cmd {
	return func() Msg {
		return setBackgroundColorMsg{c}
	}
}

// setForegroundColorMsg is a message that sets the terminal foreground color.
type setForegroundColorMsg struct{ color.Color }

// SetForegroundColor is a command that sets the terminal foreground color.
func SetForegroundColor(c color.Color) Cmd {
	return func() Msg {
		return setForegroundColorMsg{c}
	}
}

// setCursorColorMsg is a message that sets the terminal cursor color.
type setCursorColorMsg struct{ color.Color }

// SetCursorColor is a command that sets the terminal cursor color.
func SetCursorColor(c color.Color) Cmd {
	return func() Msg {
		return setCursorColorMsg{c}
	}
}

// ForegroundColorMsg represents a foreground color message. This message is
// emitted when the program requests the terminal foreground color with the
// [RequestForegroundColor] Cmd.
type ForegroundColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e ForegroundColorMsg) String() string {
	return colorToHex(e.Color)
}

// IsDark returns whether the color is dark.
func (e ForegroundColorMsg) IsDark() bool {
	return isDarkColor(e.Color)
}

// BackgroundColorMsg represents a background color message. This message is
// emitted when the program requests the terminal background color with the
// [RequestBackgroundColor] Cmd.
//
// This is commonly used in [Update.Init] to get the terminal background color
// for style definitions. For that you'll want to call
// [BackgroundColorMsg.IsDark] to determine if the color is dark or light. For
// example:
//
//	func (m Model) Init() (Model, Cmd) {
//	  return m, RequestBackgroundColor()
//	}
//
//	func (m Model) Update(msg Msg) (Model, Cmd) {
//	  switch msg := msg.(type) {
//	  case BackgroundColorMsg:
//	      m.styles = newStyles(msg.IsDark())
//	  }
//	}
type BackgroundColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e BackgroundColorMsg) String() string {
	return colorToHex(e)
}

// IsDark returns whether the color is dark.
func (e BackgroundColorMsg) IsDark() bool {
	return isDarkColor(e.Color)
}

// CursorColorMsg represents a cursor color change message. This message is
// emitted when the program requests the terminal cursor color.
type CursorColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e CursorColorMsg) String() string {
	return colorToHex(e)
}

// IsDark returns whether the color is dark.
func (e CursorColorMsg) IsDark() bool {
	return isDarkColor(e)
}

type shiftable interface {
	~uint | ~uint16 | ~uint32 | ~uint64
}

func shift[T shiftable](x T) T {
	if x > 0xff {
		x >>= 8
	}
	return x
}

func colorToHex(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", shift(r), shift(g), shift(b))
}

func xParseColor(s string) color.Color {
	switch {
	case strings.HasPrefix(s, "rgb:"):
		parts := strings.Split(s[4:], "/")
		if len(parts) != 3 {
			return color.Black
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)

		return color.RGBA{uint8(shift(r)), uint8(shift(g)), uint8(shift(b)), 255} //nolint:gosec
	case strings.HasPrefix(s, "rgba:"):
		parts := strings.Split(s[5:], "/")
		if len(parts) != 4 {
			return color.Black
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)
		a, _ := strconv.ParseUint(parts[3], 16, 32)

		return color.RGBA{uint8(shift(r)), uint8(shift(g)), uint8(shift(b)), uint8(shift(a))} //nolint:gosec
	}
	return nil
}

func getMaxMin(a, b, c float64) (max, min float64) { //nolint:predeclared
	// TODO: use go1.21 min/max functions
	if a > b {
		max = a
		min = b
	} else {
		max = b
		min = a
	}
	if c > max {
		max = c
	} else if c < min {
		min = c
	}
	return max, min
}

func round(x float64) float64 {
	return math.Round(x*1000) / 1000
}

// rgbToHSL converts an RGB triple to an HSL triple.
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	// convert uint32 pre-multiplied value to uint8
	// The r,g,b values are divided by 255 to change the range from 0..255 to 0..1:
	Rnot := float64(r) / 255
	Gnot := float64(g) / 255
	Bnot := float64(b) / 255
	Cmax, Cmin := getMaxMin(Rnot, Gnot, Bnot)
	Δ := Cmax - Cmin
	// Lightness calculation:
	l = (Cmax + Cmin) / 2
	// Hue and Saturation Calculation:
	if Δ == 0 {
		h = 0
		s = 0
	} else {
		switch Cmax {
		case Rnot:
			h = 60 * (math.Mod((Gnot-Bnot)/Δ, 6))
		case Gnot:
			h = 60 * (((Bnot - Rnot) / Δ) + 2)
		case Bnot:
			h = 60 * (((Rnot - Gnot) / Δ) + 4)
		}
		if h < 0 {
			h += 360
		}

		s = Δ / (1 - math.Abs((2*l)-1))
	}

	return h, round(s), round(l)
}

// isDarkColor returns whether the given color is dark.
func isDarkColor(c color.Color) bool {
	if c == nil {
		return true
	}

	r, g, b, _ := c.RGBA()
	_, _, l := rgbToHSL(uint8(r>>8), uint8(g>>8), uint8(b>>8)) //nolint:gosec
	return l < 0.5
}
