package slice

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Take(t *testing.T) {
	for _, test := range []struct {
		input    []int
		take     int
		expected []int
	}{
		{
			input:    []int{1, 2, 3, 4, 5},
			take:     3,
			expected: []int{1, 2, 3},
		},
		{
			input:    []int{1, 2, 3},
			take:     5,
			expected: []int{1, 2, 3},
		},
		{
			input:    []int{},
			take:     2,
			expected: []int{},
		},
		{
			input:    []int{1, 2, 3},
			take:     0,
			expected: []int{},
		},
		{
			input:    nil,
			take:     2,
			expected: nil,
		},
	} {
		t.Run(fmt.Sprint(test.input, test.take), func(t *testing.T) {
			assert.Equal(t, test.expected, Take(test.input, test.take))
		})
	}
}
