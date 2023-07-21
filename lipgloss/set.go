package lipgloss

// This could (should) probably just be moved into NewStyle(). We've broken it
// out, so we can call it in a lazy way.
func (s *Style) init() {
	if s.rules == nil {
		s.rules = make(rules)
	}
}

// Set a value on the underlying rules map.
func (s *Style) set(key propKey, value interface{}) {
	s.init()

	switch v := value.(type) {
	case int:
		// We don't allow negative integers on any of our values, so just keep
		// them at zero or above. We could use uints instead, but the
		// conversions are a little tedious, so we're sticking with ints for
		// sake of usability.
		s.rules[key] = max(0, v)
	default:
		s.rules[key] = v
	}
}

// Bold sets a bold formatting rule.
func (s Style) Bold(v bool) Style {
	s.set(boldKey, v)
	return s
}

// Italic sets an italic formatting rule. In some terminal emulators this will
// render with "reverse" coloring if not italic font variant is available.
func (s Style) Italic(v bool) Style {
	s.set(italicKey, v)
	return s
}

// Underline sets an underline rule. By default, underlines will not be drawn on
// whitespace like margins and padding. To change this behavior set
// UnderlineSpaces.
func (s Style) Underline(v bool) Style {
	s.set(underlineKey, v)
	return s
}

// Strikethrough sets a strikethrough rule. By default, strikes will not be
// drawn on whitespace like margins and padding. To change this behavior set
// StrikethroughSpaces.
func (s Style) Strikethrough(v bool) Style {
	s.set(strikethroughKey, v)
	return s
}

// Reverse sets a rule for inverting foreground and background colors.
func (s Style) Reverse(v bool) Style {
	s.set(reverseKey, v)
	return s
}

// Blink sets a rule for blinking foreground text.
func (s Style) Blink(v bool) Style {
	s.set(blinkKey, v)
	return s
}

// Faint sets a rule for rendering the foreground color in a dimmer shade.
func (s Style) Faint(v bool) Style {
	s.set(faintKey, v)
	return s
}

// Foreground sets a foreground color.
//
//	// Sets the foreground to blue
//	s := lipgloss.NewStyle().Foreground(lipgloss.Color("#0000ff"))
//
//	// Removes the foreground color
//	s.Foreground(lipgloss.NoColor)
func (s Style) Foreground(c TerminalColor) Style {
	s.set(foregroundKey, c)
	return s
}

// Background sets a background color.
func (s Style) Background(c TerminalColor) Style {
	s.set(backgroundKey, c)
	return s
}

// Width sets the width of the block before applying margins. The width, if
// set, also determines where text will wrap.
func (s Style) Width(i int) Style {
	s.set(widthKey, i)
	return s
}

// Height sets the height of the block before applying margins. If the height of
// the text block is less than this value after applying padding (or not), the
// block will be set to this height.
func (s Style) Height(i int) Style {
	s.set(heightKey, i)
	return s
}

// Align is a shorthand method for setting horizontal and vertical alignment.
//
// With one argument, the position value is applied to the horizontal alignment.
//
// With two arguments, the value is applied to the vertical and horizontal
// alignments, in that order.
func (s Style) Align(p ...Position) Style {
	if len(p) > 0 {
		s.set(alignHorizontalKey, p[0])
	}
	if len(p) > 1 {
		s.set(alignVerticalKey, p[1])
	}
	return s
}

// AlignHorizontal sets a horizontal text alignment rule.
func (s Style) AlignHorizontal(p Position) Style {
	s.set(alignHorizontalKey, p)
	return s
}

// AlignVertical sets a vertical text alignment rule.
func (s Style) AlignVertical(p Position) Style {
	s.set(alignVerticalKey, p)
	return s
}

// Padding is a shorthand method for setting padding on all sides at once.
//
// With one argument, the value is applied to all sides.
//
// With two arguments, the value is applied to the vertical and horizontal
// sides, in that order.
//
// With three arguments, the value is applied to the top side, the horizontal
// sides, and the bottom side, in that order.
//
// With four arguments, the value is applied clockwise starting from the top
// side, followed by the right side, then the bottom, and finally the left.
//
// With more than four arguments no padding will be added.
func (s Style) Padding(i ...int) Style {
	top, right, bottom, left, ok := whichSidesInt(i...)
	if !ok {
		return s
	}

	s.set(paddingTopKey, top)
	s.set(paddingRightKey, right)
	s.set(paddingBottomKey, bottom)
	s.set(paddingLeftKey, left)
	return s
}

// PaddingLeft adds padding on the left.
func (s Style) PaddingLeft(i int) Style {
	s.set(paddingLeftKey, i)
	return s
}

// PaddingRight adds padding on the right.
func (s Style) PaddingRight(i int) Style {
	s.set(paddingRightKey, i)
	return s
}

// PaddingTop adds padding to the top of the block.
func (s Style) PaddingTop(i int) Style {
	s.set(paddingTopKey, i)
	return s
}

// PaddingBottom adds padding to the bottom of the block.
func (s Style) PaddingBottom(i int) Style {
	s.set(paddingBottomKey, i)
	return s
}

// ColorWhitespace determines whether or not the background color should be
// applied to the padding. This is true by default as it's more than likely the
// desired and expected behavior, but it can be disabled for certain graphic
// effects.
func (s Style) ColorWhitespace(v bool) Style {
	s.set(colorWhitespaceKey, v)
	return s
}

// Margin is a shorthand method for setting margins on all sides at once.
//
// With one argument, the value is applied to all sides.
//
// With two arguments, the value is applied to the vertical and horizontal
// sides, in that order.
//
// With three arguments, the value is applied to the top side, the horizontal
// sides, and the bottom side, in that order.
//
// With four arguments, the value is applied clockwise starting from the top
// side, followed by the right side, then the bottom, and finally the left.
//
// With more than four arguments no margin will be added.
func (s Style) Margin(i ...int) Style {
	top, right, bottom, left, ok := whichSidesInt(i...)
	if !ok {
		return s
	}

	s.set(marginTopKey, top)
	s.set(marginRightKey, right)
	s.set(marginBottomKey, bottom)
	s.set(marginLeftKey, left)
	return s
}

// MarginLeft sets the value of the left margin.
func (s Style) MarginLeft(i int) Style {
	s.set(marginLeftKey, i)
	return s
}

// MarginRight sets the value of the right margin.
func (s Style) MarginRight(i int) Style {
	s.set(marginRightKey, i)
	return s
}

// MarginTop sets the value of the top margin.
func (s Style) MarginTop(i int) Style {
	s.set(marginTopKey, i)
	return s
}

// MarginBottom sets the value of the bottom margin.
func (s Style) MarginBottom(i int) Style {
	s.set(marginBottomKey, i)
	return s
}

// MarginBackground sets the background color of the margin. Note that this is
// also set when inheriting from a style with a background color. In that case
// the background color on that style will set the margin color on this style.
func (s Style) MarginBackground(c TerminalColor) Style {
	s.set(marginBackgroundKey, c)
	return s
}

// Border is shorthand for setting the border style and which sides should
// have a border at once. The variadic argument sides works as follows:
//
// With one value, the value is applied to all sides.
//
// With two values, the values are applied to the vertical and horizontal
// sides, in that order.
//
// With three values, the values are applied to the top side, the horizontal
// sides, and the bottom side, in that order.
//
// With four values, the values are applied clockwise starting from the top
// side, followed by the right side, then the bottom, and finally the left.
//
// With more than four arguments the border will be applied to all sides.
//
// Examples:
//
//	// Applies borders to the top and bottom only
//	lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false)
//
//	// Applies rounded borders to the right and bottom only
//	lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, true, true, false)
func (s Style) Border(b Border, sides ...bool) Style {
	s.set(borderStyleKey, b)

	top, right, bottom, left, ok := whichSidesBool(sides...)
	if !ok {
		top = true
		right = true
		bottom = true
		left = true
	}

	s.set(borderTopKey, top)
	s.set(borderRightKey, right)
	s.set(borderBottomKey, bottom)
	s.set(borderLeftKey, left)

	return s
}

// BorderStyle defines the Border on a style. A Border contains a series of
// definitions for the sides and corners of a border.
//
// Note that if border visibility has not been set for any sides when setting
// the border style, the border will be enabled for all sides during rendering.
//
// You can define border characters as you'd like, though several default
// styles are included: NormalBorder(), RoundedBorder(), BlockBorder(),
// OuterHalfBlockBorder(), InnerHalfBlockBorder(), ThickBorder(),
// and DoubleBorder().
//
// Example:
//
//	lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder())
func (s Style) BorderStyle(b Border) Style {
	s.set(borderStyleKey, b)
	return s
}

// BorderTop determines whether or not to draw a top border.
func (s Style) BorderTop(v bool) Style {
	s.set(borderTopKey, v)
	return s
}

// BorderRight determines whether or not to draw a right border.
func (s Style) BorderRight(v bool) Style {
	s.set(borderRightKey, v)
	return s
}

// BorderBottom determines whether or not to draw a bottom border.
func (s Style) BorderBottom(v bool) Style {
	s.set(borderBottomKey, v)
	return s
}

// BorderLeft determines whether or not to draw a left border.
func (s Style) BorderLeft(v bool) Style {
	s.set(borderLeftKey, v)
	return s
}

// BorderForeground is a shorthand function for setting all of the
// foreground colors of the borders at once. The arguments work as follows:
//
// With one argument, the argument is applied to all sides.
//
// With two arguments, the arguments are applied to the vertical and horizontal
// sides, in that order.
//
// With three arguments, the arguments are applied to the top side, the
// horizontal sides, and the bottom side, in that order.
//
// With four arguments, the arguments are applied clockwise starting from the
// top side, followed by the right side, then the bottom, and finally the left.
//
// With more than four arguments nothing will be set.
func (s Style) BorderForeground(c ...TerminalColor) Style {
	if len(c) == 0 {
		return s
	}

	top, right, bottom, left, ok := whichSidesColor(c...)
	if !ok {
		return s
	}

	s.set(borderTopForegroundKey, top)
	s.set(borderRightForegroundKey, right)
	s.set(borderBottomForegroundKey, bottom)
	s.set(borderLeftForegroundKey, left)

	return s
}

// BorderTopForeground set the foreground color for the top of the border.
func (s Style) BorderTopForeground(c TerminalColor) Style {
	s.set(borderTopForegroundKey, c)
	return s
}

// BorderRightForeground sets the foreground color for the right side of the
// border.
func (s Style) BorderRightForeground(c TerminalColor) Style {
	s.set(borderRightForegroundKey, c)
	return s
}

// BorderBottomForeground sets the foreground color for the bottom of the
// border.
func (s Style) BorderBottomForeground(c TerminalColor) Style {
	s.set(borderBottomForegroundKey, c)
	return s
}

// BorderLeftForeground sets the foreground color for the left side of the
// border.
func (s Style) BorderLeftForeground(c TerminalColor) Style {
	s.set(borderLeftForegroundKey, c)
	return s
}

// BorderBackground is a shorthand function for setting all of the
// background colors of the borders at once. The arguments work as follows:
//
// With one argument, the argument is applied to all sides.
//
// With two arguments, the arguments are applied to the vertical and horizontal
// sides, in that order.
//
// With three arguments, the arguments are applied to the top side, the
// horizontal sides, and the bottom side, in that order.
//
// With four arguments, the arguments are applied clockwise starting from the
// top side, followed by the right side, then the bottom, and finally the left.
//
// With more than four arguments nothing will be set.
func (s Style) BorderBackground(c ...TerminalColor) Style {
	if len(c) == 0 {
		return s
	}

	top, right, bottom, left, ok := whichSidesColor(c...)
	if !ok {
		return s
	}

	s.set(borderTopBackgroundKey, top)
	s.set(borderRightBackgroundKey, right)
	s.set(borderBottomBackgroundKey, bottom)
	s.set(borderLeftBackgroundKey, left)

	return s
}

// BorderTopBackground sets the background color of the top of the border.
func (s Style) BorderTopBackground(c TerminalColor) Style {
	s.set(borderTopBackgroundKey, c)
	return s
}

// BorderRightBackground sets the background color of right side the border.
func (s Style) BorderRightBackground(c TerminalColor) Style {
	s.set(borderRightBackgroundKey, c)
	return s
}

// BorderBottomBackground sets the background color of the bottom of the
// border.
func (s Style) BorderBottomBackground(c TerminalColor) Style {
	s.set(borderBottomBackgroundKey, c)
	return s
}

// BorderLeftBackground set the background color of the left side of the
// border.
func (s Style) BorderLeftBackground(c TerminalColor) Style {
	s.set(borderLeftBackgroundKey, c)
	return s
}

// Inline makes rendering output one line and disables the rendering of
// margins, padding and borders. This is useful when you need a style to apply
// only to font rendering and don't want it to change any physical dimensions.
// It works well with Style.MaxWidth.
//
// Because this in intended to be used at the time of render, this method will
// not mutate the style and instead return a copy.
//
// Example:
//
//	var userInput string = "..."
//	var userStyle = text.Style{ /* ... */ }
//	fmt.Println(userStyle.Inline(true).Render(userInput))
func (s Style) Inline(v bool) Style {
	o := s.Copy()
	o.set(inlineKey, v)
	return o
}

// MaxWidth applies a max width to a given style. This is useful in enforcing
// a certain width at render time, particularly with arbitrary strings and
// styles.
//
// Because this in intended to be used at the time of render, this method will
// not mutate the style and instead return a copy.
//
// Example:
//
//	var userInput string = "..."
//	var userStyle = text.Style{ /* ... */ }
//	fmt.Println(userStyle.MaxWidth(16).Render(userInput))
func (s Style) MaxWidth(n int) Style {
	o := s.Copy()
	o.set(maxWidthKey, n)
	return o
}

// MaxHeight applies a max height to a given style. This is useful in enforcing
// a certain height at render time, particularly with arbitrary strings and
// styles.
//
// Because this in intended to be used at the time of render, this method will
// not mutate the style and instead return a copy.
func (s Style) MaxHeight(n int) Style {
	o := s.Copy()
	o.set(maxHeightKey, n)
	return o
}

// UnderlineSpaces determines whether to underline spaces between words. By
// default, this is true. Spaces can also be underlined without underlining the
// text itself.
func (s Style) UnderlineSpaces(v bool) Style {
	s.set(underlineSpacesKey, v)
	return s
}

// StrikethroughSpaces determines whether to apply strikethroughs to spaces
// between words. By default, this is true. Spaces can also be struck without
// underlining the text itself.
func (s Style) StrikethroughSpaces(v bool) Style {
	s.set(strikethroughSpacesKey, v)
	return s
}

// Renderer sets the renderer for the style. This is useful for changing the
// renderer for a style that is being used in a different context.
func (s Style) Renderer(r *Renderer) Style {
	s.r = r
	return s
}

// whichSidesInt is a helper method for setting values on sides of a block based
// on the number of arguments. It follows the CSS shorthand rules for blocks
// like margin, padding. and borders. Here are how the rules work:
//
// 0 args:  do nothing
// 1 arg:   all sides
// 2 args:  top -> bottom
// 3 args:  top -> horizontal -> bottom
// 4 args:  top -> right -> bottom -> left
// 5+ args: do nothing.
func whichSidesInt(i ...int) (top, right, bottom, left int, ok bool) {
	switch len(i) {
	case 1:
		top = i[0]
		bottom = i[0]
		left = i[0]
		right = i[0]
		ok = true
	case 2:
		top = i[0]
		bottom = i[0]
		left = i[1]
		right = i[1]
		ok = true
	case 3:
		top = i[0]
		left = i[1]
		right = i[1]
		bottom = i[2]
		ok = true
	case 4:
		top = i[0]
		right = i[1]
		bottom = i[2]
		left = i[3]
		ok = true
	}
	return top, right, bottom, left, ok
}

// whichSidesBool is like whichSidesInt, except it operates on a series of
// boolean values. See the comment on whichSidesInt for details on how this
// works.
func whichSidesBool(i ...bool) (top, right, bottom, left bool, ok bool) {
	switch len(i) {
	case 1:
		top = i[0]
		bottom = i[0]
		left = i[0]
		right = i[0]
		ok = true
	case 2:
		top = i[0]
		bottom = i[0]
		left = i[1]
		right = i[1]
		ok = true
	case 3:
		top = i[0]
		left = i[1]
		right = i[1]
		bottom = i[2]
		ok = true
	case 4:
		top = i[0]
		right = i[1]
		bottom = i[2]
		left = i[3]
		ok = true
	}
	return top, right, bottom, left, ok
}

// whichSidesColor is like whichSides, except it operates on a series of
// boolean values. See the comment on whichSidesInt for details on how this
// works.
func whichSidesColor(i ...TerminalColor) (top, right, bottom, left TerminalColor, ok bool) {
	switch len(i) {
	case 1:
		top = i[0]
		bottom = i[0]
		left = i[0]
		right = i[0]
		ok = true
	case 2:
		top = i[0]
		bottom = i[0]
		left = i[1]
		right = i[1]
		ok = true
	case 3:
		top = i[0]
		left = i[1]
		right = i[1]
		bottom = i[2]
		ok = true
	case 4:
		top = i[0]
		right = i[1]
		bottom = i[2]
		left = i[3]
		ok = true
	}
	return top, right, bottom, left, ok
}
