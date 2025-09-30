package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type model struct {
	value int
	state tea.ProgressBarState
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
	v := tea.NewView("Press up/down to change value, left/right to change state, q to quit.\n")
	v.ProgressBar = tea.NewProgressBar(m.state, m.value)
	return v
}

func main() {
	p := tea.NewProgram(model{value: 50, state: tea.ProgressBarIndeterminate})
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
