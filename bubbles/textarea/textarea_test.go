package textarea

import (
	"strings"
	"testing"

	tea "github.com/rprtr258/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	textarea := newTextArea()
	view := textarea.View()

	assert.Contains(t, view, ">", "Text area did not render the prompt")
	assert.Contains(t, view, "World!", "Text area did not render the placeholder")
}

func TestInput(t *testing.T) {
	textarea := newTextArea()

	input := "foo"

	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	assert.Contains(t, view, input, "Text area did not render the input")
	assert.Len(t, input, textarea.col, "Text area did not move the cursor to the correct position")
}

func TestSoftWrap(t *testing.T) {
	textarea := newTextArea()
	textarea.Prompt = ""
	textarea.ShowLineNumbers = false
	textarea.SetWidth(5)
	textarea.SetHeight(5)
	textarea.CharLimit = 60

	textarea, _ = textarea.Update(nil)

	input := "foo bar baz"

	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	for _, word := range strings.Split(input, " ") {
		assert.Contains(t, view, word, "Text area did not render the input")
	}

	// Due to the word wrapping, each word will be on a new line and the
	// text area will look like this:
	//
	// > foo
	// > bar
	// > baz█
	//
	// However, due to soft-wrapping the column will still be at the end of the line.
	assert.Equal(t, 0, textarea.row)
	assert.Len(t, input, textarea.col, "Text area did not move the cursor to the correct position")
}

func TestCharLimit(t *testing.T) {
	textarea := newTextArea()

	// First input (foo bar) should be accepted as it will fall within the
	// CharLimit. Second input (baz) should not appear in the input.
	input := []string{"foo bar", "baz"}
	textarea.CharLimit = len(input[0])

	for _, k := range strings.Join(input, " ") {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()
	assert.NotContains(t, view, input[1], "Text area should not include input past the character limit")
}

func TestVerticalScrolling(t *testing.T) {
	textarea := newTextArea()
	textarea.Prompt = ""
	textarea.ShowLineNumbers = false
	textarea.SetHeight(1)
	textarea.SetWidth(20)
	textarea.CharLimit = 100

	textarea, _ = textarea.Update(nil)

	for _, k := range "This is a really long line that should wrap around the text area." {
		textarea, _ = textarea.Update(keyPress(k))
	}

	view := textarea.View()

	// The view should contain the first "line" of the input.
	assert.Contains(t, view, "This is a really", "Text area did not render the input")

	// But we should be able to scroll to see the next line.
	// Let's scroll down for each line to view the full input.
	for _, line := range []string{
		"long line that",
		"should wrap around",
		"the text area.",
	} {
		textarea.viewport.LineDown(1)
		view = textarea.View()
		assert.Contains(t, view, line, "Text area did not render the correct scrolled input")
	}
}

func TestWordWrapOverflowing(t *testing.T) {
	// An interesting edge case is when the user enters many words that fill up
	// the text area and then goes back up and inserts a few words which causes
	// a cascading wrap and causes an overflow of the last line.
	//
	// In this case, we should not let the user insert more words if, after the
	// entire wrap is complete, the last line is overflowing.
	textarea := newTextArea()

	textarea.SetHeight(3)
	textarea.SetWidth(20)
	textarea.CharLimit = 500

	textarea, _ = textarea.Update(nil)

	input := "Testing Testing Testing Testing Testing Testing Testing Testing"

	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	// We have essentially filled the text area with input.
	// Let's see if we can cause wrapping to overflow the last line.
	textarea.row = 0
	textarea.col = 0

	input = "Testing"

	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	lastLineWidth := textarea.LineInfo().Width
	assert.LessOrEqual(t, lastLineWidth, 20)
}

func TestValueSoftWrap(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(16)
	textarea.SetHeight(10)
	textarea.CharLimit = 500

	textarea, _ = textarea.Update(nil)

	input := "Testing Testing Testing Testing Testing Testing Testing Testing"

	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
		textarea.View()
	}

	value := textarea.Value()
	assert.Equal(t, input, value)
}

func TestSetValue(t *testing.T) {
	textarea := newTextArea()
	textarea.SetValue(strings.Join([]string{"Foo", "Bar", "Baz"}, "\n"))

	assert.Equal(t, 2, textarea.row, "Cursor Should be on row 2 column 3 after inserting 2 new lines")
	assert.Equal(t, 3, textarea.col, "Cursor Should be on row 2 column 3 after inserting 2 new lines")

	value := textarea.Value()
	assert.Equal(t, "Foo\nBar\nBaz", value)

	// SetValue should reset text area
	textarea.SetValue("Test")
	value = textarea.Value()
	assert.Equal(t, "Test", value, "Text area was not reset when SetValue() was called")
}

func TestInsertString(t *testing.T) {
	textarea := newTextArea()

	// Insert some text
	for _, k := range "foo baz" {
		textarea, _ = textarea.Update(keyPress(k))
	}

	// Put cursor in the middle of the text
	textarea.col = 4

	textarea.InsertString("bar ")

	value := textarea.Value()
	assert.Equal(t, "foo bar baz", value, "Expected insert string to insert bar between foo and baz")
}

func TestCanHandleEmoji(t *testing.T) {
	textarea := newTextArea()
	input := "🧋"

	for _, k := range input {
		textarea, _ = textarea.Update(keyPress(k))
	}

	value := textarea.Value()
	assert.Equal(t, input, value, "Expected emoji to be inserted")

	input = "🧋🧋🧋"

	textarea.SetValue(input)

	value = textarea.Value()
	assert.Equal(t, input, value, "Expected emoji to be inserted")

	assert.Equal(t, 3, textarea.col, "Expected cursor to be on the third character")

	charOffset := textarea.LineInfo().CharOffset
	assert.Equal(t, 6, charOffset, "Expected cursor to be on the sixth character")
}

func TestVerticalNavigationKeepsCursorHorizontalPosition(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(20)

	textarea.SetValue(strings.Join([]string{"你好你好", "Hello"}, "\n"))

	textarea.row = 0
	textarea.col = 2

	// 你好|你好
	// Hell|o
	// 1234|

	// Let's imagine our cursor is on the first line where the pipe is.
	// We press the down arrow to get to the next line.
	// The issue is that if we keep the cursor on the same column, the cursor will jump to after the `e`.
	//
	// 你好|你好
	// He|llo
	//
	// But this is wrong because visually we were at the 4th character due to
	// the first line containing double-width runes.
	// We want to keep the cursor on the same visual column.
	//
	// 你好|你好
	// Hell|o
	//
	// This test ensures that the cursor is kept on the same visual column by
	// ensuring that the column offset goes from 2 -> 4.

	lineInfo := textarea.LineInfo()
	assert.Equal(
		t, 4, lineInfo.CharOffset,
		"Expected cursor to be on the fourth character because there are two double width runes on the first line.",
	)
	assert.Equal(t, 2, lineInfo.ColumnOffset)

	downMsg := tea.MsgKey{Type: tea.KeyDown, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(downMsg)

	lineInfo = textarea.LineInfo()
	assert.Equal(
		t, 4, lineInfo.CharOffset,
		"Expected cursor to be on the fourth character because we came down from the first line.",
	)
	assert.Equal(t, 4, lineInfo.ColumnOffset)
}

func TestVerticalNavigationShouldRememberPositionWhileTraversing(t *testing.T) {
	textarea := newTextArea()
	textarea.SetWidth(40)

	// Let's imagine we have a text area with the following content:
	//
	// Hello
	// World
	// This is a long line.
	//
	// If we are at the end of the last line and go up, we should be at the end
	// of the second line.
	// And, if we go up again we should be at the end of the first line.
	// But, if we go back down twice, we should be at the end of the last line
	// again and not the fifth (length of second line) character of the last line.
	//
	// In other words, we should remember the last horizontal position while
	// traversing vertically.

	textarea.SetValue(strings.Join([]string{"Hello", "World", "This is a long line."}, "\n"))

	// We are at the end of the last line.
	assert.Equal(t, 20, textarea.col, "Expected cursor to be on the 20th character of the last line")
	assert.Equal(t, 2, textarea.row)

	// Let's go up.
	upMsg := tea.MsgKey{Type: tea.KeyUp, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(upMsg)

	// We should be at the end of the second line.
	assert.Equal(t, 5, textarea.col, "Expected cursor to be on the 5th character of the second line")
	assert.Equal(t, 1, textarea.row)

	// And, again.
	textarea, _ = textarea.Update(upMsg)

	// We should be at the end of the first line.
	assert.Equal(t, 5, textarea.col, "Expected cursor to be on the 5th character of the first line")
	assert.Equal(t, 0, textarea.row)

	// Let's go down, twice.
	downMsg := tea.MsgKey{Type: tea.KeyDown, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(downMsg)
	textarea, _ = textarea.Update(downMsg)

	// We should be at the end of the last line.
	assert.Equal(t, 20, textarea.col, "Expected cursor to be on the 20th character of the last line")
	assert.Equal(t, 2, textarea.row)

	// Now, for correct behavior, if we move right or left, we should forget
	// (reset) the saved horizontal position. Since we assume the user wants to
	// keep the cursor where it is horizontally. This is how most text areas
	// work.

	textarea, _ = textarea.Update(upMsg)
	leftMsg := tea.MsgKey{Type: tea.KeyLeft, Alt: false, Runes: []rune{}}
	textarea, _ = textarea.Update(leftMsg)

	assert.Equal(t, 4, textarea.col, "Expected cursor to be on the 5th character of the second line")
	assert.Equal(t, 1, textarea.row)

	// Going down now should keep us at the 4th column since we moved left and
	// reset the horizontal position saved state.
	textarea, _ = textarea.Update(downMsg)
	assert.Equal(t, 4, textarea.col, "Expected cursor to be on the 4th character of the last line")
	assert.Equal(t, 2, textarea.row)
}

func TestRendersEndOfLineBuffer(t *testing.T) {
	textarea := newTextArea()
	textarea.ShowLineNumbers = true
	textarea.SetWidth(20)

	view := textarea.View()
	assert.Contains(t, view, "~", "Expected to see a tilde at the end of the line")
}

func newTextArea() Model {
	textarea := New()

	textarea.Prompt = "> "
	textarea.Placeholder = "Hello, World!"

	textarea.Focus()

	textarea, _ = textarea.Update(nil)

	return textarea
}

func keyPress(key rune) tea.Msg {
	return tea.MsgKey{
		Type:  tea.KeyRunes,
		Runes: []rune{key},
		Alt:   false,
	}
}
