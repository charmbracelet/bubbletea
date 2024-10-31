package main

import (
	"log"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type model struct {
	input textinput.Model
	width int
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyPressMsg:
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
	w := min(m.width, 60)

	instructions := lipgloss.NewStyle().
		Width(w).
		Render("Query for terminal capabilities. You can enter things like 'TN', 'RGB', 'cols', and so on. This will not work in all terminals and multiplexers.")

	return "\n" + instructions + "\n\n" +
		m.input.View() +
		"\n\nPress enter to request capability, or ctrl+c to quit."
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		log.Fatal(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
