package lipgloss

import (
	"strings"

	"github.com/muesli/reflow/ansi"
)

// GetBold returns the style's bold value. If no value is set false is returned.
func (s Style) GetBold() bool {
	return s.getAsBool(boldKey, false)
}

// GetItalic returns the style's italic value. If no value is set false is
// returned.
func (s Style) GetItalic() bool {
	return s.getAsBool(italicKey, false)
}

// GetUnderline returns the style's underline value. If no value is set false is
// returned.
func (s Style) GetUnderline() bool {
	return s.getAsBool(underlineKey, false)
}

// GetStrikethrough returns the style's strikethrough value. If no value is set false
// is returned.
func (s Style) GetStrikethrough() bool {
	return s.getAsBool(strikethroughKey, false)
}

// GetReverse returns the style's reverse value. If no value is set false is
// returned.
func (s Style) GetReverse() bool {
	return s.getAsBool(reverseKey, false)
}

// GetBlink returns the style's blink value. If no value is set false is
// returned.
func (s Style) GetBlink() bool {
	return s.getAsBool(blinkKey, false)
}

// GetFaint returns the style's faint value. If no value is set false is
// returned.
func (s Style) GetFaint() bool {
	return s.getAsBool(faintKey, false)
}

// GetForeground returns the style's foreground color. If no value is set
// NoColor{} is returned.
func (s Style) GetForeground() TerminalColor {
	return s.getAsColor(foregroundKey)
}

// GetBackground returns the style's background color. If no value is set
// NoColor{} is returned.
func (s Style) GetBackground() TerminalColor {
	return s.getAsColor(backgroundKey)
}

// GetWidth returns the style's width setting. If no width is set 0 is
// returned.
func (s Style) GetWidth() int {
	return s.getAsInt(widthKey)
}

// GetHeight returns the style's height setting. If no height is set 0 is
// returned.
func (s Style) GetHeight() int {
	return s.getAsInt(heightKey)
}

// GetAlign returns the style's implicit horizontal alignment setting.
// If no alignment is set Position.Left is returned.
func (s Style) GetAlign() Position {
	v := s.getAsPosition(alignHorizontalKey)
	if v == Position(0) {
		return Left
	}
	return v
}

// GetAlignHorizontal returns the style's implicit horizontal alignment setting.
// If no alignment is set Position.Left is returned.
func (s Style) GetAlignHorizontal() Position {
	v := s.getAsPosition(alignHorizontalKey)
	if v == Position(0) {
		return Left
	}
	return v
}

// GetAlignVertical returns the style's implicit vertical alignment setting.
// If no alignment is set Position.Top is returned.
func (s Style) GetAlignVertical() Position {
	v := s.getAsPosition(alignVerticalKey)
	if v == Position(0) {
		return Top
	}
	return v
}

// GetPadding returns the style's top, right, bottom, and left padding values,
// in that order. 0 is returned for unset values.
func (s Style) GetPadding() (top, right, bottom, left int) {
	return s.getAsInt(paddingTopKey),
		s.getAsInt(paddingRightKey),
		s.getAsInt(paddingBottomKey),
		s.getAsInt(paddingLeftKey)
}

// GetPaddingTop returns the style's top padding. If no value is set 0 is
// returned.
func (s Style) GetPaddingTop() int {
	return s.getAsInt(paddingTopKey)
}

// GetPaddingRight returns the style's right padding. If no value is set 0 is
// returned.
func (s Style) GetPaddingRight() int {
	return s.getAsInt(paddingRightKey)
}

// GetPaddingBottom returns the style's bottom padding. If no value is set 0 is
// returned.
func (s Style) GetPaddingBottom() int {
	return s.getAsInt(paddingBottomKey)
}

// GetPaddingLeft returns the style's left padding. If no value is set 0 is
// returned.
func (s Style) GetPaddingLeft() int {
	return s.getAsInt(paddingLeftKey)
}

// GetHorizontalPadding returns the style's left and right padding. Unset
// values are measured as 0.
func (s Style) GetHorizontalPadding() int {
	return s.getAsInt(paddingLeftKey) + s.getAsInt(paddingRightKey)
}

// GetVerticalPadding returns the style's top and bottom padding. Unset values
// are measured as 0.
func (s Style) GetVerticalPadding() int {
	return s.getAsInt(paddingTopKey) + s.getAsInt(paddingBottomKey)
}

// GetColorWhitespace returns the style's whitespace coloring setting. If no
// value is set false is returned.
func (s Style) GetColorWhitespace() bool {
	return s.getAsBool(colorWhitespaceKey, false)
}

// GetMargin returns the style's top, right, bottom, and left margins, in that
// order. 0 is returned for unset values.
func (s Style) GetMargin() (top, right, bottom, left int) {
	return s.getAsInt(marginTopKey),
		s.getAsInt(marginRightKey),
		s.getAsInt(marginBottomKey),
		s.getAsInt(marginLeftKey)
}

// GetMarginTop returns the style's top margin. If no value is set 0 is
// returned.
func (s Style) GetMarginTop() int {
	return s.getAsInt(marginTopKey)
}

// GetMarginRight returns the style's right margin. If no value is set 0 is
// returned.
func (s Style) GetMarginRight() int {
	return s.getAsInt(marginRightKey)
}

// GetMarginBottom returns the style's bottom margin. If no value is set 0 is
// returned.
func (s Style) GetMarginBottom() int {
	return s.getAsInt(marginBottomKey)
}

// GetMarginLeft returns the style's left margin. If no value is set 0 is
// returned.
func (s Style) GetMarginLeft() int {
	return s.getAsInt(marginLeftKey)
}

// GetHorizontalMargins returns the style's left and right margins. Unset
// values are measured as 0.
func (s Style) GetHorizontalMargins() int {
	return s.getAsInt(marginLeftKey) + s.getAsInt(marginRightKey)
}

// GetVerticalMargins returns the style's top and bottom margins. Unset values
// are measured as 0.
func (s Style) GetVerticalMargins() int {
	return s.getAsInt(marginTopKey) + s.getAsInt(marginBottomKey)
}

// GetBorder returns the style's border style (type Border) and value for the
// top, right, bottom, and left in that order. If no value is set for the
// border style, Border{} is returned. For all other unset values false is
// returned.
func (s Style) GetBorder() (b Border, top, right, bottom, left bool) {
	return s.getBorderStyle(),
		s.getAsBool(borderTopKey, false),
		s.getAsBool(borderRightKey, false),
		s.getAsBool(borderBottomKey, false),
		s.getAsBool(borderLeftKey, false)
}

// GetBorderStyle returns the style's border style (type Border). If no value
// is set Border{} is returned.
func (s Style) GetBorderStyle() Border {
	return s.getBorderStyle()
}

// GetBorderTop returns the style's top border setting. If no value is set
// false is returned.
func (s Style) GetBorderTop() bool {
	return s.getAsBool(borderTopKey, false)
}

// GetBorderRight returns the style's right border setting. If no value is set
// false is returned.
func (s Style) GetBorderRight() bool {
	return s.getAsBool(borderRightKey, false)
}

// GetBorderBottom returns the style's bottom border setting. If no value is
// set false is returned.
func (s Style) GetBorderBottom() bool {
	return s.getAsBool(borderBottomKey, false)
}

// GetBorderLeft returns the style's left border setting. If no value is
// set false is returned.
func (s Style) GetBorderLeft() bool {
	return s.getAsBool(borderLeftKey, false)
}

// GetBorderTopForeground returns the style's border top foreground color. If
// no value is set NoColor{} is returned.
func (s Style) GetBorderTopForeground() TerminalColor {
	return s.getAsColor(borderTopForegroundKey)
}

// GetBorderRightForeground returns the style's border right foreground color.
// If no value is set NoColor{} is returned.
func (s Style) GetBorderRightForeground() TerminalColor {
	return s.getAsColor(borderRightForegroundKey)
}

// GetBorderBottomForeground returns the style's border bottom foreground
// color.  If no value is set NoColor{} is returned.
func (s Style) GetBorderBottomForeground() TerminalColor {
	return s.getAsColor(borderBottomForegroundKey)
}

// GetBorderLeftForeground returns the style's border left foreground
// color.  If no value is set NoColor{} is returned.
func (s Style) GetBorderLeftForeground() TerminalColor {
	return s.getAsColor(borderLeftForegroundKey)
}

// GetBorderTopBackground returns the style's border top background color. If
// no value is set NoColor{} is returned.
func (s Style) GetBorderTopBackground() TerminalColor {
	return s.getAsColor(borderTopBackgroundKey)
}

// GetBorderRightBackground returns the style's border right background color.
// If no value is set NoColor{} is returned.
func (s Style) GetBorderRightBackground() TerminalColor {
	return s.getAsColor(borderRightBackgroundKey)
}

// GetBorderBottomBackground returns the style's border bottom background
// color.  If no value is set NoColor{} is returned.
func (s Style) GetBorderBottomBackground() TerminalColor {
	return s.getAsColor(borderBottomBackgroundKey)
}

// GetBorderLeftBackground returns the style's border left background
// color.  If no value is set NoColor{} is returned.
func (s Style) GetBorderLeftBackground() TerminalColor {
	return s.getAsColor(borderLeftBackgroundKey)
}

// GetBorderTopWidth returns the width of the top border. If borders contain
// runes of varying widths, the widest rune is returned. If no border exists on
// the top edge, 0 is returned.
//
// Deprecated: This function simply calls Style.GetBorderTopSize.
func (s Style) GetBorderTopWidth() int {
	return s.GetBorderTopSize()
}

// GetBorderTopSize returns the width of the top border. If borders contain
// runes of varying widths, the widest rune is returned. If no border exists on
// the top edge, 0 is returned.
func (s Style) GetBorderTopSize() int {
	if !s.getAsBool(borderTopKey, false) {
		return 0
	}
	return s.getBorderStyle().GetTopSize()
}

// GetBorderLeftSize returns the width of the left border. If borders contain
// runes of varying widths, the widest rune is returned. If no border exists on
// the left edge, 0 is returned.
func (s Style) GetBorderLeftSize() int {
	if !s.getAsBool(borderLeftKey, false) {
		return 0
	}
	return s.getBorderStyle().GetLeftSize()
}

// GetBorderBottomSize returns the width of the bottom border. If borders
// contain runes of varying widths, the widest rune is returned. If no border
// exists on the left edge, 0 is returned.
func (s Style) GetBorderBottomSize() int {
	if !s.getAsBool(borderBottomKey, false) {
		return 0
	}
	return s.getBorderStyle().GetBottomSize()
}

// GetBorderRightSize returns the width of the right border. If borders
// contain runes of varying widths, the widest rune is returned. If no border
// exists on the right edge, 0 is returned.
func (s Style) GetBorderRightSize() int {
	if !s.getAsBool(borderRightKey, false) {
		return 0
	}
	return s.getBorderStyle().GetBottomSize()
}

// GetHorizontalBorderSize returns the width of the horizontal borders. If
// borders contain runes of varying widths, the widest rune is returned. If no
// border exists on the horizontal edges, 0 is returned.
func (s Style) GetHorizontalBorderSize() int {
	b := s.getBorderStyle()
	return b.GetLeftSize() + b.GetRightSize()
}

// GetVerticalBorderSize returns the width of the vertical borders. If
// borders contain runes of varying widths, the widest rune is returned. If no
// border exists on the vertical edges, 0 is returned.
func (s Style) GetVerticalBorderSize() int {
	b := s.getBorderStyle()
	return b.GetTopSize() + b.GetBottomSize()
}

// GetInline returns the style's inline setting. If no value is set false is
// returned.
func (s Style) GetInline() bool {
	return s.getAsBool(inlineKey, false)
}

// GetMaxWidth returns the style's max width setting. If no value is set 0 is
// returned.
func (s Style) GetMaxWidth() int {
	return s.getAsInt(maxWidthKey)
}

// GetMaxHeight returns the style's max height setting. If no value is set 0 is
// returned.
func (s Style) GetMaxHeight() int {
	return s.getAsInt(maxHeightKey)
}

// GetUnderlineSpaces returns whether or not the style is set to underline
// spaces. If not value is set false is returned.
func (s Style) GetUnderlineSpaces() bool {
	return s.getAsBool(underlineSpacesKey, false)
}

// GetStrikethroughSpaces returns whether or not the style is set to strikethrough
// spaces. If not value is set false is returned.
func (s Style) GetStrikethroughSpaces() bool {
	return s.getAsBool(strikethroughSpacesKey, false)
}

// GetHorizontalFrameSize returns the sum of the style's horizontal margins, padding
// and border widths.
//
// Provisional: this method may be renamed.
func (s Style) GetHorizontalFrameSize() int {
	return s.GetHorizontalMargins() + s.GetHorizontalPadding() + s.GetHorizontalBorderSize()
}

// GetVerticalFrameSize returns the sum of the style's vertical margins, padding
// and border widths.
//
// Provisional: this method may be renamed.
func (s Style) GetVerticalFrameSize() int {
	return s.GetVerticalMargins() + s.GetVerticalPadding() + s.GetVerticalBorderSize()
}

// GetFrameSize returns the sum of the margins, padding and border width for
// both the horizontal and vertical margins.
func (s Style) GetFrameSize() (x, y int) {
	return s.GetHorizontalFrameSize(), s.GetVerticalFrameSize()
}

// Returns whether or not the given property is set.
func (s Style) isSet(k propKey) bool {
	_, exists := s.rules[k]
	return exists
}

func (s Style) getAsBool(k propKey, defaultVal bool) bool {
	v, ok := s.rules[k]
	if !ok {
		return defaultVal
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return defaultVal
}

func (s Style) getAsColor(k propKey) TerminalColor {
	v, ok := s.rules[k]
	if !ok {
		return noColor
	}
	if c, ok := v.(TerminalColor); ok {
		return c
	}
	return noColor
}

func (s Style) getAsInt(k propKey) int {
	v, ok := s.rules[k]
	if !ok {
		return 0
	}
	if i, ok := v.(int); ok {
		return i
	}
	return 0
}

func (s Style) getAsPosition(k propKey) Position {
	v, ok := s.rules[k]
	if !ok {
		return Position(0)
	}
	if p, ok := v.(Position); ok {
		return p
	}
	return Position(0)
}

func (s Style) getBorderStyle() Border {
	v, ok := s.rules[borderStyleKey]
	if !ok {
		return noBorder
	}
	if b, ok := v.(Border); ok {
		return b
	}
	return noBorder
}

// Split a string into lines, additionally returning the size of the widest
// line.
func getLines(s string) (lines []string, widest int) {
	lines = strings.Split(s, "\n")

	for _, l := range lines {
		w := ansi.PrintableRuneWidth(l)
		if widest < w {
			widest = w
		}
	}

	return lines, widest
}
