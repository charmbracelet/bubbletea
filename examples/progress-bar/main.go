package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var body = lipgloss.NewStyle().Padding(1, 2)

type model struct {
	value int
	width int
	state tea.ProgressBarState
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.value < 100 {
				m.value += 10
			}
		case "down", "j":
			if m.value > 0 {
				m.value -= 10
			}
		case "left", "h":
			if m.state > 0 {
				m.state--
			}
		case "right", "l":
			if m.state < 4 {
				m.state++
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	s := body.Width(m.width - body.GetHorizontalPadding()).Render(
		"This demo requires a terminal emulator that supports an indeterminate progress bar, such a Windows Terminal or Ghostty. In other terminals (including tmux in a supporting terminal) nothing will happen.\n\nPress up/down to change value, left/right to change state, q to quit.",
	)
	v := tea.NewView(s)
	v.ProgressBar = tea.NewProgressBar(m.state, m.value)
	return v
}

func main() {
	p := tea.NewProgram(model{value: 50, state: tea.ProgressBarIndeterminate})
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
