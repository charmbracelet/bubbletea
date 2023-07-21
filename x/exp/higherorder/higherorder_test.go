package higherorder

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func Test_Foldl(t *testing.T) {
	x := Foldl(func(a, b int) int {
		return a + b
	}, 0, []int{1, 2, 3})

	assert.Equal(t, 6, x)
}

func Test_Foldr(t *testing.T) {
	x := Foldl(func(a, b int) int {
		return a - b
	}, 6, []int{1, 2, 3})

	assert.Equal(t, 0, x)
}

func Test_Map(t *testing.T) {
	{
		// Map over ints, returning the square of each int.
		// (Take ints, return ints.)
		x := Map(func(a int) int {
			return a * a
		}, []int{2, 3, 4})

		assert.Equal(t, []int{4, 9, 16}, x)
	}
	{
		// Map over strings, returning the length of each string.
		// (Take ints, return strings.)
		x := Map(utf8.RuneCountInString, []string{"one", "two", "three"})

		assert.Equal(t, []int{3, 3, 5}, x)
	}
}
