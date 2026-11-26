package main

// A simple example illustrating how to change the mouse pointer shape. Move
// the mouse over a shape name to switch the pointer to that shape. The pointer
// is reset to its default shape on exit.

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// header is the number of lines printed above the list of shapes.
const header = 2

var shapes = []tea.PointerShape{
	tea.PointerDefault,
	tea.PointerText,
	tea.PointerPointer,
	tea.PointerMove,
	tea.PointerGrab,
	tea.PointerCrosshair,
	tea.PointerWait,
	tea.PointerHelp,
	tea.PointerNotAllowed,
	tea.PointerEWResize,
	tea.PointerNSResize,
	tea.PointerNWSEResize,
	tea.PointerNESWResize,
	tea.PointerZoomIn,
}

var (
	normalStyle = lipgloss.NewStyle().Padding(0, 1)
	hoverStyle  = normalStyle.Reverse(true)
)

type model struct {
	hovered tea.PointerShape // the shape under the mouse, or empty
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
			return m, tea.Quit
		}

	case tea.MouseMotionMsg:
		mouse := msg.Mouse()
		m.hovered = ""
		if i := mouse.Y - header; i >= 0 && i < len(shapes) && mouse.X < lipgloss.Width(normalStyle.Render(string(shapes[i]))) {
			m.hovered = shapes[i]
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString("Hover over a shape name to change the mouse pointer. Press q to quit.\n\n")
	for _, shape := range shapes {
		style := normalStyle
		if shape == m.hovered {
			style = hoverStyle
		}
		b.WriteString(style.Render(string(shape)) + "\n")
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	v.PointerShape = m.hovered
	return v
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
