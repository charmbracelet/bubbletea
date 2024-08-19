package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
		return m, tea.Printf("You pressed: %s\n", msg.String())
	}
	return m, nil
}

func (m model) View() string {
	return "Press any key to see it printed to the terminal. Press 'ctrl+c' to quit."
}

func main() {
	p := tea.NewProgram(model{}, tea.WithEnhancedKeyboard())
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
	}
}
