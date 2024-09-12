package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

var _ tea.Model = model{}

// Init implements tea.Model.
func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tea.TerminalVersion
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.BackgroundColorMsg, tea.ForegroundColorMsg, tea.CursorColorMsg, tea.TerminalVersionMsg, tea.ColorProfileMsg:
		return m, tea.Printf("Received a terminal startup message: %T: %s", msg, msg)
	}
	return m, nil
}

// View implements tea.Model.
func (m model) View() string {
	return "Press any key to exit."
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
