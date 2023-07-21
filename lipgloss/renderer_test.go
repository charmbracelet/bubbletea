package lipgloss

import (
	"os"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func TestRendererHasDarkBackground(t *testing.T) {
	r1 := NewRenderer(os.Stdout)
	r1.SetHasDarkBackground(false)
	assert.False(t, r1.HasDarkBackground(), "Expected renderer to have light background")

	r2 := NewRenderer(os.Stdout)
	r2.SetHasDarkBackground(true)
	assert.True(t, r2.HasDarkBackground(), "Expected renderer to have dark background")
}

func TestRendererWithOutput(t *testing.T) {
	f, err := os.Create(t.Name())
	assert.NoError(t, err)
	defer f.Close()
	defer os.Remove(f.Name())

	r := NewRenderer(f)
	r.SetColorProfile(termenv.TrueColor)
	assert.Equal(t, termenv.TrueColor, r.output.Profile)
}
