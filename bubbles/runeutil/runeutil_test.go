package runeutil

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	td := []struct {
		input, output string
	}{
		{"", ""},
		{"x", "x"},
		{"\n", "XX"},
		{"\na\n", "XXaXX"},
		{"\n\n", "XXXX"},
		{"\t", ""},
		{"hello", "hello"},
		{"hel\nlo", "helXXlo"},
		{"hel\rlo", "helXXlo"},
		{"hel\tlo", "hello"},
		{"he\n\nl\tlo", "heXXXXllo"},
		{"he\tl\n\nlo", "helXXXXlo"},
		{"hel\x1blo", "hello"},
		{"hello\xc2", "hello"}, // invalid utf8
	}

	for _, tc := range td {
		runes := make([]rune, 0, len(tc.input))
		b := []byte(tc.input)
		for i, w := 0, 0; i < len(b); i += w {
			var r rune
			r, w = utf8.DecodeRune(b[i:])
			runes = append(runes, r)
		}
		t.Logf("input runes: %+v", runes)
		s := NewSanitizer(ReplaceNewlines("XX"), ReplaceTabs(""))
		result := s.Sanitize(runes)
		rs := string(result)
		assert.Equal(t, tc.output, rs, "input: %q, result: %v", tc.input, result)
	}
}
