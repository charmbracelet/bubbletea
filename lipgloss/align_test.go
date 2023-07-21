package lipgloss

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlignTextVertical(t *testing.T) {
	for _, test := range []struct {
		str    string
		pos    Position
		height int
		want   string
	}{
		{str: "Foo", pos: Top, height: 2, want: "Foo\n"},
		{str: "Foo", pos: Center, height: 5, want: "\n\nFoo\n\n"},
		{str: "Foo", pos: Bottom, height: 5, want: "\n\n\n\nFoo"},

		{str: "Foo\nBar", pos: Bottom, height: 5, want: "\n\n\nFoo\nBar"},
		{str: "Foo\nBar", pos: Center, height: 5, want: "\nFoo\nBar\n\n"},
		{str: "Foo\nBar", pos: Top, height: 5, want: "Foo\nBar\n\n\n"},

		{str: "Foo\nBar\nBaz", pos: Bottom, height: 5, want: "\n\nFoo\nBar\nBaz"},
		{str: "Foo\nBar\nBaz", pos: Center, height: 5, want: "\nFoo\nBar\nBaz\n"},

		{str: "Foo\nBar\nBaz", pos: Bottom, height: 3, want: "Foo\nBar\nBaz"},
		{str: "Foo\nBar\nBaz", pos: Center, height: 3, want: "Foo\nBar\nBaz"},
		{str: "Foo\nBar\nBaz", pos: Top, height: 3, want: "Foo\nBar\nBaz"},

		{str: "Foo\n\n\n\nBar", pos: Bottom, height: 5, want: "Foo\n\n\n\nBar"},
		{str: "Foo\n\n\n\nBar", pos: Center, height: 5, want: "Foo\n\n\n\nBar"},
		{str: "Foo\n\n\n\nBar", pos: Top, height: 5, want: "Foo\n\n\n\nBar"},

		{str: "Foo\nBar\nBaz", pos: Center, height: 9, want: "\n\n\nFoo\nBar\nBaz\n\n\n"},
		{str: "Foo\nBar\nBaz", pos: Center, height: 10, want: "\n\n\nFoo\nBar\nBaz\n\n\n\n"},
	} {
		t.Run(fmt.Sprintf("str=%q pos=%v height=%d", test.str, test.pos, test.height), func(t *testing.T) {
			assert.Equal(t, test.want, alignTextVertical(test.str, test.pos, test.height, nil))
		})
	}
}
