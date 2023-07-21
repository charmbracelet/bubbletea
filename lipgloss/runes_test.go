package lipgloss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStyleRunes(t *testing.T) {
	matchedStyle := NewStyle().Reverse(true)
	unmatchedStyle := NewStyle()

	for name, test := range map[string]struct {
		input    string
		indices  []int
		expected string
	}{
		"hello 0": {
			"hello",
			[]int{0},
			"\x1b[7mh\x1b[0mello",
		},
		"你好 1": {
			"你好",
			[]int{1},
			"你\x1b[7m好\x1b[0m",
		},
		"hello 你好 6,7": {
			"hello 你好",
			[]int{6, 7},
			"hello \x1b[7m你好\x1b[0m",
		},
		"hello 1,3": {
			"hello",
			[]int{1, 3},
			"h\x1b[7me\x1b[0ml\x1b[7ml\x1b[0mo",
		},
		"你好 0,1": {
			"你好",
			[]int{0, 1},
			"\x1b[7m你好\x1b[0m",
		},
	} {
		t.Run(name, func(t *testing.T) {
			actual := StyleRunes(test.input, test.indices, matchedStyle, unmatchedStyle)
			assert.Equal(t, test.expected, actual)
		})
	}
}
