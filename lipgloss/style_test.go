package lipgloss

import (
	"fmt"
	"io"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func TestStyleRender(t *testing.T) {
	renderer.SetColorProfile(termenv.TrueColor)
	renderer.SetHasDarkBackground(true)
	t.Parallel()

	for i, tc := range []struct {
		style    Style
		expected string
	}{
		{
			NewStyle().Foreground(Color("#5A56E0")),
			"\x1b[38;2;89;86;224mhello\x1b[0m",
		},
		{
			NewStyle().Foreground(AdaptiveColor{Light: "#fffe12", Dark: "#5A56E0"}),
			"\x1b[38;2;89;86;224mhello\x1b[0m",
		},
		{
			NewStyle().Bold(true),
			"\x1b[1mhello\x1b[0m",
		},
		{
			NewStyle().Italic(true),
			"\x1b[3mhello\x1b[0m",
		},
		{
			NewStyle().Underline(true),
			"\x1b[4;4mh\x1b[0m\x1b[4;4me\x1b[0m\x1b[4;4ml\x1b[0m\x1b[4;4ml\x1b[0m\x1b[4;4mo\x1b[0m",
		},
		{
			NewStyle().Blink(true),
			"\x1b[5mhello\x1b[0m",
		},
		{
			NewStyle().Faint(true),
			"\x1b[2mhello\x1b[0m",
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			s := tc.style.Copy().SetString("hello")
			res := s.Render()
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestStyleCustomRender(t *testing.T) {
	r := NewRenderer(io.Discard)
	r.SetHasDarkBackground(false)
	r.SetColorProfile(termenv.TrueColor)
	for i, tc := range []struct {
		style    Style
		expected string
	}{
		{
			r.NewStyle().Foreground(Color("#5A56E0")),
			"\x1b[38;2;89;86;224mhello\x1b[0m",
		},
		{
			r.NewStyle().Foreground(AdaptiveColor{Light: "#fffe12", Dark: "#5A56E0"}),
			"\x1b[38;2;255;254;18mhello\x1b[0m",
		},
		{
			r.NewStyle().Bold(true),
			"\x1b[1mhello\x1b[0m",
		},
		{
			r.NewStyle().Italic(true),
			"\x1b[3mhello\x1b[0m",
		},
		{
			r.NewStyle().Underline(true),
			"\x1b[4;4mh\x1b[0m\x1b[4;4me\x1b[0m\x1b[4;4ml\x1b[0m\x1b[4;4ml\x1b[0m\x1b[4;4mo\x1b[0m",
		},
		{
			r.NewStyle().Blink(true),
			"\x1b[5mhello\x1b[0m",
		},
		{
			r.NewStyle().Faint(true),
			"\x1b[2mhello\x1b[0m",
		},
		{
			NewStyle().Faint(true).Renderer(r),
			"\x1b[2mhello\x1b[0m",
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			s := tc.style.Copy().SetString("hello")
			res := s.Render()
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestStyleRenderer(t *testing.T) {
	r := NewRenderer(io.Discard)
	s1 := NewStyle().Bold(true)
	s2 := s1.Renderer(r)
	assert.NotEqual(t, s1.r, s2.r)
}

func TestValueCopy(t *testing.T) {
	t.Parallel()

	s := NewStyle().
		Bold(true)

	i := s
	i.Bold(false)

	assert.Equal(t, s.GetBold(), i.GetBold())
}

func TestStyleInherit(t *testing.T) {
	t.Parallel()

	s := NewStyle().
		Bold(true).
		Italic(true).
		Underline(true).
		Strikethrough(true).
		Blink(true).
		Faint(true).
		Foreground(Color("#ffffff")).
		Background(Color("#111111")).
		Margin(1, 1, 1, 1).
		Padding(1, 1, 1, 1)

	i := NewStyle().Inherit(s)

	assert.Equal(t, s.GetBold(), i.GetBold())
	assert.Equal(t, s.GetItalic(), i.GetItalic())
	assert.Equal(t, s.GetUnderline(), i.GetUnderline())
	assert.Equal(t, s.GetStrikethrough(), i.GetStrikethrough())
	assert.Equal(t, s.GetBlink(), i.GetBlink())
	assert.Equal(t, s.GetFaint(), i.GetFaint())
	assert.Equal(t, s.GetForeground(), i.GetForeground())
	assert.Equal(t, s.GetBackground(), i.GetBackground())

	assert.NotEqual(t, s.GetMarginLeft(), i.GetMarginLeft())
	assert.NotEqual(t, s.GetMarginRight(), i.GetMarginRight())
	assert.NotEqual(t, s.GetMarginTop(), i.GetMarginTop())
	assert.NotEqual(t, s.GetMarginBottom(), i.GetMarginBottom())
	assert.NotEqual(t, s.GetPaddingLeft(), i.GetPaddingLeft())
	assert.NotEqual(t, s.GetPaddingRight(), i.GetPaddingRight())
	assert.NotEqual(t, s.GetPaddingTop(), i.GetPaddingTop())
	assert.NotEqual(t, s.GetPaddingBottom(), i.GetPaddingBottom())
}

func TestStyleCopy(t *testing.T) {
	t.Parallel()

	s := NewStyle().
		Bold(true).
		Italic(true).
		Underline(true).
		Strikethrough(true).
		Blink(true).
		Faint(true).
		Foreground(Color("#ffffff")).
		Background(Color("#111111")).
		Margin(1, 1, 1, 1).
		Padding(1, 1, 1, 1)

	i := s.Copy()

	assert.Equal(t, s.GetBold(), i.GetBold())
	assert.Equal(t, s.GetItalic(), i.GetItalic())
	assert.Equal(t, s.GetUnderline(), i.GetUnderline())
	assert.Equal(t, s.GetStrikethrough(), i.GetStrikethrough())
	assert.Equal(t, s.GetBlink(), i.GetBlink())
	assert.Equal(t, s.GetFaint(), i.GetFaint())
	assert.Equal(t, s.GetForeground(), i.GetForeground())
	assert.Equal(t, s.GetBackground(), i.GetBackground())

	assert.Equal(t, s.GetMarginLeft(), i.GetMarginLeft())
	assert.Equal(t, s.GetMarginRight(), i.GetMarginRight())
	assert.Equal(t, s.GetMarginTop(), i.GetMarginTop())
	assert.Equal(t, s.GetMarginBottom(), i.GetMarginBottom())
	assert.Equal(t, s.GetPaddingLeft(), i.GetPaddingLeft())
	assert.Equal(t, s.GetPaddingRight(), i.GetPaddingRight())
	assert.Equal(t, s.GetPaddingTop(), i.GetPaddingTop())
	assert.Equal(t, s.GetPaddingBottom(), i.GetPaddingBottom())
}

func TestStyleUnset(t *testing.T) {
	t.Parallel()

	s := NewStyle().Bold(true)
	assert.True(t, s.GetBold())
	s.UnsetBold()
	assert.False(t, s.GetBold())

	s = NewStyle().Italic(true)
	assert.True(t, s.GetItalic())
	s.UnsetItalic()
	assert.False(t, s.GetItalic())

	s = NewStyle().Underline(true)
	assert.True(t, s.GetUnderline())
	s.UnsetUnderline()
	assert.False(t, s.GetUnderline())

	s = NewStyle().Strikethrough(true)
	assert.True(t, s.GetStrikethrough())
	s.UnsetStrikethrough()
	assert.False(t, s.GetStrikethrough())

	s = NewStyle().Reverse(true)
	assert.True(t, s.GetReverse())
	s.UnsetReverse()
	assert.False(t, s.GetReverse())

	s = NewStyle().Blink(true)
	assert.True(t, s.GetBlink())
	s.UnsetBlink()
	assert.False(t, s.GetBlink())

	s = NewStyle().Faint(true)
	assert.True(t, s.GetFaint())
	s.UnsetFaint()
	assert.False(t, s.GetFaint())

	s = NewStyle().Inline(true)
	assert.True(t, s.GetInline())
	s.UnsetInline()
	assert.False(t, s.GetInline())

	// colors
	col := Color("#ffffff")
	s = NewStyle().Foreground(col)
	assert.Equal(t, col, s.GetForeground())
	s.UnsetForeground()
	assert.NotEqual(t, col, s.GetForeground())

	s = NewStyle().Background(col)
	assert.Equal(t, col, s.GetBackground())
	s.UnsetBackground()
	assert.NotEqual(t, col, s.GetBackground())

	// margins
	s = NewStyle().Margin(1, 2, 3, 4)
	assert.Equal(t, 1, s.GetMarginTop())
	s.UnsetMarginTop()
	assert.Equal(t, 0, s.GetMarginTop())

	assert.Equal(t, 2, s.GetMarginRight())
	s.UnsetMarginRight()
	assert.Equal(t, 0, s.GetMarginRight())

	assert.Equal(t, 3, s.GetMarginBottom())
	s.UnsetMarginBottom()
	assert.Equal(t, 0, s.GetMarginBottom())

	assert.Equal(t, 4, s.GetMarginLeft())
	s.UnsetMarginLeft()
	assert.Equal(t, 0, s.GetMarginLeft())

	// padding
	s = NewStyle().Padding(1, 2, 3, 4)
	assert.Equal(t, 1, s.GetPaddingTop())
	s.UnsetPaddingTop()
	assert.Equal(t, 0, s.GetPaddingTop())

	assert.Equal(t, 2, s.GetPaddingRight())
	s.UnsetPaddingRight()
	assert.Equal(t, 0, s.GetPaddingRight())

	assert.Equal(t, 3, s.GetPaddingBottom())
	s.UnsetPaddingBottom()
	assert.Equal(t, 0, s.GetPaddingBottom())

	assert.Equal(t, 4, s.GetPaddingLeft())
	s.UnsetPaddingLeft()
	assert.Equal(t, 0, s.GetPaddingLeft())

	// border
	s = NewStyle().Border(normalBorder, true, true, true, true)
	assert.True(t, s.GetBorderTop())
	s.UnsetBorderTop()
	assert.False(t, s.GetBorderTop())

	assert.True(t, s.GetBorderRight())
	s.UnsetBorderRight()
	assert.False(t, s.GetBorderRight())

	assert.True(t, s.GetBorderBottom())
	s.UnsetBorderBottom()
	assert.False(t, s.GetBorderBottom())

	assert.True(t, s.GetBorderLeft())
	s.UnsetBorderLeft()
	assert.False(t, s.GetBorderLeft())
}

func TestStyleValue(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		name     string
		style    Style
		expected string
	}{
		"empty": {
			style:    NewStyle(),
			expected: "foo",
		},
		"set string": {
			style:    NewStyle().SetString("bar"),
			expected: "bar foo",
		},
		"set string with bold": {
			style:    NewStyle().SetString("bar").Bold(true),
			expected: "\x1b[1mbar foo\x1b[0m",
		},
		"new style with string": {
			style:    NewStyle().SetString("bar", "foobar"),
			expected: "bar foobar foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			res := test.style.Render("foo")
			assert.Equal(t, test.expected, res)
		})
	}
}

func BenchmarkStyleRender(b *testing.B) {
	s := NewStyle().
		Bold(true).
		Foreground(Color("#ffffff"))

	for i := 0; i < b.N; i++ {
		s.Render("Hello world")
	}
}
