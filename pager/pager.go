package pager

import (
	"errors"
	"strings"

	"github.com/charmbracelet/boba"
)

// MODEL

type Model struct {
	Err    error
	Width  int
	Height int
	Y      int

	lines []string
}

func (m Model) ScrollPercent() float64 {
	if m.Height >= len(m.lines) {
		return 1.0
	}
	y := float64(m.Y)
	h := float64(m.Height)
	t := float64(len(m.lines))
	return (y + h) / t
}

// Content set the pager's text content
func (m *Model) Content(s string) {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\r\n", "\n", -1) // normalize line endings
	m.lines = strings.Split(s, "\n")
}

func NewModel(width, height int) Model {
	return Model{
		Width:  width,
		Height: height,
	}
}

// UPDATE

func Update(msg boba.Msg, model boba.Model) (boba.Model, boba.Cmd) {
	m, ok := model.(Model)
	if !ok {
		return Model{
			Err: errors.New("could not perform assertion on model in update in pager; are you sure you passed the correct model?"),
		}, nil
	}

	switch msg := msg.(type) {

	case boba.KeyMsg:
		switch msg.String() {

		// Up one page
		case "pgup":
			fallthrough
		case "b":
			m.Y = max(0, m.Y-m.Height)
			return m, nil

		// Down one page
		case "pgdown":
			fallthrough
		case "space":
			fallthrough
		case "f":
			m.Y = min(len(m.lines)-m.Height, m.Y+m.Height)
			return m, nil

		// Up half page
		case "u":
			m.Y = max(0, m.Y-m.Height/2)
			return m, nil

		// Down half page
		case "d":
			m.Y = min(len(m.lines)-m.Height, m.Y+m.Height/2)
			return m, nil

		// Up one line
		case "up":
			fallthrough
		case "k":
			m.Y = max(0, m.Y-1)
			return m, nil

		// Down one line
		case "down":
			fallthrough
		case "j":
			m.Y = min(len(m.lines)-m.Height, m.Y+1)
			return m, nil

		// Re-render
		case "ctrl+l":
			return m, nil

		}
	}

	return model, nil
}

// VIEW

func View(model boba.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "could not perform assertion on model in view in pager; are you sure you passed the correct model?"
	}

	if m.Err != nil {
		return m.Err.Error()
	}

	if len(m.lines) == 0 {
		return ""
	}

	// Render viewport
	top := max(0, m.Y)
	bottom := min(len(m.lines), m.Y+m.Height)
	lines := m.lines[top:bottom]
	return "\n" + strings.Join(lines, "\n")
}

// ETC

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
