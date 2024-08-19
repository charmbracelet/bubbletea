package tea

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// ForegroundColorMsg represents a foreground color message.
// This message is emitted when the program requests the terminal foreground
// color.
type ForegroundColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e ForegroundColorMsg) String() string {
	return colorToHex(e)
}

// BackgroundColorMsg represents a background color message.
// This message is emitted when the program requests the terminal background
// color.
type BackgroundColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e BackgroundColorMsg) String() string {
	return colorToHex(e)
}

// CursorColorMsg represents a cursor color change message.
// This message is emitted when the program requests the terminal cursor color.
type CursorColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e CursorColorMsg) String() string {
	return colorToHex(e)
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

		return color.RGBA{uint8(shift(r)), uint8(shift(g)), uint8(shift(b)), 255}
	case strings.HasPrefix(s, "rgba:"):
		parts := strings.Split(s[5:], "/")
		if len(parts) != 4 {
			return color.Black
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)
		a, _ := strconv.ParseUint(parts[3], 16, 32)

		return color.RGBA{uint8(shift(r)), uint8(shift(g)), uint8(shift(b)), uint8(shift(a))}
	}
	return color.Black
}
