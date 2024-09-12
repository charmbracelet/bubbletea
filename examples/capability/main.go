package main

import (
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	input textinput.Model
}

var _ tea.Model = model{}

// Init implements tea.Model.
func (m model) Init() (tea.Model, tea.Cmd) {
	m.input = textinput.New()
	m.input.Placeholder = "Enter capability name to request"
	return m, m.input.Focus()
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			input := m.input.Value()
			m.input.Reset()
			return m, tea.RequestCapability(input)
		}
	case tea.CapabilityMsg:
		return m, tea.Printf("Got capability: %s", msg)
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m model) View() string {
	return m.input.View() + "\n\nPress enter to request capability, or ctrl+c to quit."
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		log.Fatal(err)
	}
}
