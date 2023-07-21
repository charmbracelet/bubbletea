package lipgloss

import (
	"strings"

	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
)

// Perform text alignment. If the string is multi-lined, we also make all lines
// the same width by padding them with spaces. If a termenv style is passed,
// use that to style the spaces added.
func alignTextHorizontal(str string, pos Position, width int, style *termenv.Style) string {
	lines, widestLine := getLines(str)
	var b strings.Builder

	for i, l := range lines {
		lineWidth := ansi.PrintableRuneWidth(l)

		shortAmount := widestLine - lineWidth                // difference from the widest line
		shortAmount += max(0, width-(shortAmount+lineWidth)) // difference from the total width, if set

		if shortAmount > 0 {
			switch pos {
			case Right:
				s := strings.Repeat(" ", shortAmount)
				if style != nil {
					s = style.Styled(s)
				}
				l = s + l
			case Center:
				left := shortAmount / 2
				right := left + shortAmount%2 // note that we put the remainder on the right

				leftSpaces := strings.Repeat(" ", left)
				rightSpaces := strings.Repeat(" ", right)

				if style != nil {
					leftSpaces = style.Styled(leftSpaces)
					rightSpaces = style.Styled(rightSpaces)
				}
				l = leftSpaces + l + rightSpaces
			default: // Left
				s := strings.Repeat(" ", shortAmount)
				if style != nil {
					s = style.Styled(s)
				}
				l += s
			}
		}

		b.WriteString(l)
		if i < len(lines)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

func alignTextVertical(str string, pos Position, height int, _ *termenv.Style) string {
	strHeight := strings.Count(str, "\n") + 1
	if height < strHeight {
		return str
	}

	switch pos {
	case Top:
		return str + strings.Repeat("\n", height-strHeight)
	case Center:
		topPadding, bottomPadding := (height-strHeight)/2, (height-strHeight)/2
		if strHeight+topPadding+bottomPadding > height {
			topPadding--
		} else if strHeight+topPadding+bottomPadding < height {
			bottomPadding++
		}
		return strings.Repeat("\n", topPadding) + str + strings.Repeat("\n", bottomPadding)
	case Bottom:
		return strings.Repeat("\n", height-strHeight) + str
	}
	return str
}
