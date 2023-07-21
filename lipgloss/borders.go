package lipgloss

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
)

// Border contains a series of values which comprise the various parts of a
// border.
type Border struct {
	Top         string
	Bottom      string
	Left        string
	Right       string
	TopLeft     string
	TopRight    string
	BottomRight string
	BottomLeft  string
}

// GetTopSize returns the width of the top border. If borders contain runes of
// varying widths, the widest rune is returned. If no border exists on the top
// edge, 0 is returned.
func (b Border) GetTopSize() int {
	return getBorderEdgeWidth(b.TopLeft, b.Top, b.TopRight)
}

// GetRightSize returns the width of the right border. If borders contain
// runes of varying widths, the widest rune is returned. If no border exists on
// the right edge, 0 is returned.
func (b Border) GetRightSize() int {
	return getBorderEdgeWidth(b.TopRight, b.Right, b.BottomRight)
}

// GetBottomSize returns the width of the bottom border. If borders contain
// runes of varying widths, the widest rune is returned. If no border exists on
// the bottom edge, 0 is returned.
func (b Border) GetBottomSize() int {
	return getBorderEdgeWidth(b.BottomLeft, b.Bottom, b.BottomRight)
}

// GetLeftSize returns the width of the left border. If borders contain runes
// of varying widths, the widest rune is returned. If no border exists on the
// left edge, 0 is returned.
func (b Border) GetLeftSize() int {
	return getBorderEdgeWidth(b.TopLeft, b.Left, b.BottomLeft)
}

func getBorderEdgeWidth(borderParts ...string) (maxWidth int) {
	for _, piece := range borderParts {
		w := maxRuneWidth(piece)
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

var (
	noBorder = Border{}

	normalBorder = Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}

	roundedBorder = Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	blockBorder = Border{
		Top:         "█",
		Bottom:      "█",
		Left:        "█",
		Right:       "█",
		TopLeft:     "█",
		TopRight:    "█",
		BottomLeft:  "█",
		BottomRight: "█",
	}

	outerHalfBlockBorder = Border{
		Top:         "▀",
		Bottom:      "▄",
		Left:        "▌",
		Right:       "▐",
		TopLeft:     "▛",
		TopRight:    "▜",
		BottomLeft:  "▙",
		BottomRight: "▟",
	}

	innerHalfBlockBorder = Border{
		Top:         "▄",
		Bottom:      "▀",
		Left:        "▐",
		Right:       "▌",
		TopLeft:     "▗",
		TopRight:    "▖",
		BottomLeft:  "▝",
		BottomRight: "▘",
	}

	thickBorder = Border{
		Top:         "━",
		Bottom:      "━",
		Left:        "┃",
		Right:       "┃",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
	}

	doubleBorder = Border{
		Top:         "═",
		Bottom:      "═",
		Left:        "║",
		Right:       "║",
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
	}

	hiddenBorder = Border{
		Top:         " ",
		Bottom:      " ",
		Left:        " ",
		Right:       " ",
		TopLeft:     " ",
		TopRight:    " ",
		BottomLeft:  " ",
		BottomRight: " ",
	}
)

// NormalBorder returns a standard-type border with a normal weight and 90
// degree corners.
func NormalBorder() Border {
	return normalBorder
}

// RoundedBorder returns a border with rounded corners.
func RoundedBorder() Border {
	return roundedBorder
}

// BlockBorder returns a border that takes the whole block.
func BlockBorder() Border {
	return blockBorder
}

// OuterHalfBlockBorder returns a half-block border that sits outside the frame.
func OuterHalfBlockBorder() Border {
	return outerHalfBlockBorder
}

// InnerHalfBlockBorder returns a half-block border that sits inside the frame.
func InnerHalfBlockBorder() Border {
	return innerHalfBlockBorder
}

// ThickBorder returns a border that's thicker than the one returned by
// NormalBorder.
func ThickBorder() Border {
	return thickBorder
}

// DoubleBorder returns a border comprised of two thin strokes.
func DoubleBorder() Border {
	return doubleBorder
}

// HiddenBorder returns a border that renders as a series of single-cell
// spaces. It's useful for cases when you want to remove a standard border but
// maintain layout positioning. This said, you can still apply a background
// color to a hidden border.
func HiddenBorder() Border {
	return hiddenBorder
}

func (s Style) applyBorder(str string) string {
	var (
		topSet    = s.isSet(borderTopKey)
		rightSet  = s.isSet(borderRightKey)
		bottomSet = s.isSet(borderBottomKey)
		leftSet   = s.isSet(borderLeftKey)

		border    = s.getBorderStyle()
		hasTop    = s.getAsBool(borderTopKey, false)
		hasRight  = s.getAsBool(borderRightKey, false)
		hasBottom = s.getAsBool(borderBottomKey, false)
		hasLeft   = s.getAsBool(borderLeftKey, false)

		topFG    = s.getAsColor(borderTopForegroundKey)
		rightFG  = s.getAsColor(borderRightForegroundKey)
		bottomFG = s.getAsColor(borderBottomForegroundKey)
		leftFG   = s.getAsColor(borderLeftForegroundKey)

		topBG    = s.getAsColor(borderTopBackgroundKey)
		rightBG  = s.getAsColor(borderRightBackgroundKey)
		bottomBG = s.getAsColor(borderBottomBackgroundKey)
		leftBG   = s.getAsColor(borderLeftBackgroundKey)
	)

	// If a border is set and no sides have been specifically turned on or off
	// render borders on all sides.
	if border != noBorder && !(topSet || rightSet || bottomSet || leftSet) {
		hasTop = true
		hasRight = true
		hasBottom = true
		hasLeft = true
	}

	// If no border is set or all borders are been disabled, abort.
	if border == noBorder || (!hasTop && !hasRight && !hasBottom && !hasLeft) {
		return str
	}

	lines, width := getLines(str)

	if hasLeft {
		if border.Left == "" {
			border.Left = " "
		}
		width += maxRuneWidth(border.Left)
	}

	if hasRight && border.Right == "" {
		border.Right = " "
	}

	// If corners should be rendered but are set with the empty string, fill them
	// with a single space.
	if hasTop && hasLeft && border.TopLeft == "" {
		border.TopLeft = " "
	}
	if hasTop && hasRight && border.TopRight == "" {
		border.TopRight = " "
	}
	if hasBottom && hasLeft && border.BottomLeft == "" {
		border.BottomLeft = " "
	}
	if hasBottom && hasRight && border.BottomRight == "" {
		border.BottomRight = " "
	}

	// Figure out which corners we should actually be using based on which
	// sides are set to show.
	if hasTop {
		switch {
		case !hasLeft && !hasRight:
			border.TopLeft = ""
			border.TopRight = ""
		case !hasLeft:
			border.TopLeft = ""
		case !hasRight:
			border.TopRight = ""
		}
	}
	if hasBottom {
		switch {
		case !hasLeft && !hasRight:
			border.BottomLeft = ""
			border.BottomRight = ""
		case !hasLeft:
			border.BottomLeft = ""
		case !hasRight:
			border.BottomRight = ""
		}
	}

	// For now, limit corners to one rune.
	border.TopLeft = getFirstRuneAsString(border.TopLeft)
	border.TopRight = getFirstRuneAsString(border.TopRight)
	border.BottomRight = getFirstRuneAsString(border.BottomRight)
	border.BottomLeft = getFirstRuneAsString(border.BottomLeft)

	var out strings.Builder

	// Render top
	if hasTop {
		top := renderHorizontalEdge(border.TopLeft, border.Top, border.TopRight, width)
		top = s.styleBorder(top, topFG, topBG)
		out.WriteString(top)
		out.WriteRune('\n')
	}

	leftRunes := []rune(border.Left)
	leftIndex := 0

	rightRunes := []rune(border.Right)
	rightIndex := 0

	// Render sides
	for i, l := range lines {
		if hasLeft {
			r := string(leftRunes[leftIndex])
			leftIndex++
			if leftIndex >= len(leftRunes) {
				leftIndex = 0
			}
			out.WriteString(s.styleBorder(r, leftFG, leftBG))
		}
		out.WriteString(l)
		if hasRight {
			r := string(rightRunes[rightIndex])
			rightIndex++
			if rightIndex >= len(rightRunes) {
				rightIndex = 0
			}
			out.WriteString(s.styleBorder(r, rightFG, rightBG))
		}
		if i < len(lines)-1 {
			out.WriteRune('\n')
		}
	}

	// Render bottom
	if hasBottom {
		bottom := renderHorizontalEdge(border.BottomLeft, border.Bottom, border.BottomRight, width)
		bottom = s.styleBorder(bottom, bottomFG, bottomBG)
		out.WriteRune('\n')
		out.WriteString(bottom)
	}

	return out.String()
}

// Render the horizontal (top or bottom) portion of a border.
func renderHorizontalEdge(left, middle, right string, width int) string {
	if width < 1 {
		return ""
	}

	if middle == "" {
		middle = " "
	}

	leftWidth := ansi.PrintableRuneWidth(left)
	rightWidth := ansi.PrintableRuneWidth(right)

	runes := []rune(middle)
	j := 0

	out := strings.Builder{}
	out.WriteString(left)
	for i := leftWidth + rightWidth; i < width+rightWidth; {
		out.WriteRune(runes[j])
		j++
		if j >= len(runes) {
			j = 0
		}
		i += ansi.PrintableRuneWidth(string(runes[j]))
	}
	out.WriteString(right)

	return out.String()
}

// Apply foreground and background styling to a border.
func (s Style) styleBorder(border string, fg, bg TerminalColor) string {
	if fg == noColor && bg == noColor {
		return border
	}

	style := termenv.Style{}

	if fg != noColor {
		style = style.Foreground(fg.color(s.r))
	}
	if bg != noColor {
		style = style.Background(bg.color(s.r))
	}

	return style.Styled(border)
}

func maxRuneWidth(str string) int {
	width := 0
	for _, r := range str {
		w := runewidth.RuneWidth(r)
		if w > width {
			width = w
		}
	}
	return width
}

func getFirstRuneAsString(str string) string {
	if str == "" {
		return str
	}
	r := []rune(str)
	return string(r[0])
}
