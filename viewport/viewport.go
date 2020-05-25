package viewport

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MODEL

type Model struct {
	Err    error
	Width  int
	Height int
	Y      int

	lines []string
}

// Scrollpercent returns the amount scrolled as a float between 0 and 1.
func (m Model) ScrollPercent() float64 {
	if m.Height >= len(m.lines) {
		return 1.0
	}
	y := float64(m.Y)
	h := float64(m.Height)
	t := float64(len(m.lines))
	return y / (t - h)
}

// SetContent set the pager's text content.
func (m *Model) SetContent(s string) {
	s = strings.Replace(s, "\r\n", "\n", -1) // normalize line endings
	m.lines = strings.Split(s, "\n")
}

// NewModel creates a new pager model. Pass the dimensions of the pager.
func NewModel(width, height int) Model {
	return Model{
		Width:  width,
		Height: height,
	}
}

// ViewDown moves the view down by the number of lines in the viewport.
// Basically, "page down".
func (m *Model) ViewDown() {
	m.Y = min(len(m.lines)-m.Height, m.Y+m.Height)
}

// ViewUp moves the view up by one height of the viewport. Basically, "page up".
func (m *Model) ViewUp() {
	m.Y = max(0, m.Y-m.Height)
}

// HalfViewUp moves the view up by half the height of the viewport.
func (m *Model) HalfViewUp() {
	m.Y = max(0, m.Y-m.Height/2)
}

// HalfViewDown moves the view down by half the height of the viewport.
func (m *Model) HalfViewDown() {
	m.Y = min(len(m.lines)-m.Height, m.Y+m.Height/2)
}

// LineDown moves the view up by the given number of lines.
func (m *Model) LineDown(n int) {
	m.Y = min(len(m.lines)-m.Height, m.Y+n)
}

// LineDown moves the view down by the given number of lines.
func (m *Model) LineUp(n int) {
	m.Y = max(0, m.Y-n)
}

// UPDATE

// Update runs the update loop with default keybindings. To define your own
// keybindings use the methods on Model.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		// Down one page
		case "pgdown":
			fallthrough
		case " ": // spacebar
			fallthrough
		case "f":
			m.ViewDown()
			return m, nil

		// Up one page
		case "pgup":
			fallthrough
		case "b":
			m.ViewUp()
			return m, nil

		// Down half page
		case "d":
			m.HalfViewDown()
			return m, nil

		// Up half page
		case "u":
			m.HalfViewUp()
			return m, nil

		// Down one line
		case "down":
			fallthrough
		case "j":
			m.LineDown(1)
			return m, nil

		// Up one line
		case "up":
			fallthrough
		case "k":
			m.LineUp(1)
			return m, nil
		}
	}

	return m, nil
}

// VIEW

// View renders the viewport into a string.
func View(m Model) string {
	if m.Err != nil {
		return m.Err.Error()
	}

	var lines []string

	if len(m.lines) > 0 {
		top := max(0, m.Y)
		bottom := min(len(m.lines), m.Y+m.Height)
		lines = m.lines[top:bottom]
	}

	// Fill empty space with newlines
	extraLines := ""
	if len(lines) < m.Height {
		extraLines = strings.Repeat("\n", m.Height-len(lines))
	}

	return strings.Join(lines, "\n") + extraLines
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
