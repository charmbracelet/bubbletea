package lipgloss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinVertical(t *testing.T) {
	for name, test := range map[string]struct {
		result   string
		expected string
	}{
		"pos0":    {JoinVertical(0, "A", "BBBB"), "A   \nBBBB"},
		"pos1":    {JoinVertical(1, "A", "BBBB"), "   A\nBBBB"},
		"pos0.25": {JoinVertical(0.25, "A", "BBBB"), " A  \nBBBB"},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.result)
		})
	}
}
