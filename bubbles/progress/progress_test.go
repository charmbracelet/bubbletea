package progress

import (
	"strings"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

const (
	AnsiReset = "\x1b[0m"
)

func TestGradient(t *testing.T) {
	colA := "#FF0000"
	colB := "#00FF00"

	for name, opts := range map[string][]Option{
		"progress bar with gradient": {
			WithColorProfile(termenv.TrueColor),
			WithoutPercentage(),
			WithGradient(colA, colB),
		},
		"progress bar with scaled gradient": {
			WithColorProfile(termenv.TrueColor),
			WithoutPercentage(),
			WithScaledGradient(colA, colB),
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := New(opts...)

			// build the expected colors by colorizing an empty string and then cutting off the following reset sequence
			sb := strings.Builder{}
			sb.WriteString(termenv.String("").Foreground(p.color(colA)).String())
			expFirst := strings.Split(sb.String(), AnsiReset)[0]
			sb.Reset()
			sb.WriteString(termenv.String("").Foreground(p.color(colB)).String())
			expLast := strings.Split(sb.String(), AnsiReset)[0]

			for _, width := range []int{3, 5, 50} {
				p.Width = width
				res := p.ViewAs(1.0)

				// extract colors from the progrss bar by splitting at p.Full+AnsiReset, leaving us with just the color sequences
				colors := strings.Split(res, string(p.Full)+AnsiReset)

				// discard the last color, because it is empty (no new color comes after the last char of the bar)
				colors = colors[0 : len(colors)-1]

				assert.Equal(t, expFirst, colors[0])
				assert.Equal(t, expLast, colors[len(colors)-1])
			}
		})
	}
}
