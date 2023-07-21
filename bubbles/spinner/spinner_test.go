package spinner_test

import (
	"testing"

	"github.com/rprtr258/bubbletea/bubbles/spinner"
	"github.com/stretchr/testify/assert"
)

func TestSpinnerNew(t *testing.T) {
	assertEqualSpinner := func(t *testing.T, exp, got spinner.Spinner) {
		t.Helper()

		assert.Equal(t, exp.FPS, got.FPS)
		assert.Equal(t, exp.Frames, got.Frames)
	}
	t.Run("default", func(t *testing.T) {
		s := spinner.New()
		assertEqualSpinner(t, spinner.Line, s.Spinner)
	})

	t.Run("WithSpinner", func(t *testing.T) {
		customSpinner := spinner.Spinner{
			Frames: []string{"a", "b", "c", "d"},
			FPS:    16,
		}

		s := spinner.New(spinner.WithSpinner(customSpinner))
		assertEqualSpinner(t, customSpinner, s.Spinner)
	})

	for name, s := range map[string]spinner.Spinner{
		"Line":    spinner.Line,
		"Dot":     spinner.Dot,
		"MiniDot": spinner.MiniDot,
		"Jump":    spinner.Jump,
		"Pulse":   spinner.Pulse,
		"Points":  spinner.Points,
		"Globe":   spinner.Globe,
		"Moon":    spinner.Moon,
		"Monkey":  spinner.Monkey,
	} {
		t.Run(name, func(t *testing.T) {
			assertEqualSpinner(t, spinner.New(spinner.WithSpinner(s)).Spinner, s)
		})
	}
}
