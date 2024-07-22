package main

import (
	"image/color"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	lightGreen = lipgloss.Color("#8cc14c")
	darkGreen  = lipgloss.Color("#3b6c2b")
)

type model struct {
	green color.Color
}

func (m model) Init() (tea.Model, tea.Cmd) {
	m.green = lightGreen
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		col, ok := colorful.MakeColor(msg)
		if ok {
			if lipgloss.IsDarkColor(col) {
				m.green = lightGreen
			} else {
				m.green = darkGreen
			}
			return m, tea.Printf("Background color: %s", col.Hex())
		}

	case tea.WindowSizeMsg:
		return m, tea.Printf("Window size: %dx%d", msg.Width, msg.Height)

	case tea.KeyPressMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	return lipgloss.NewStyle().Foreground(m.green).
		Render("Press any key to quit\n")
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
