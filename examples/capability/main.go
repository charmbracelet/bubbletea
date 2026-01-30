package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	input textinput.Model
	width int
}

var _ tea.Model = model{}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return m.input.Focus()
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
func (m model) View() tea.View {
	w := min(m.width, 60)

	instructions := lipgloss.NewStyle().
		Width(w).
		Render("Query for terminal capabilities. You can enter things like 'TN', 'RGB', 'cols', and so on. This will not work in all terminals and multiplexers.")

	return tea.NewView("\n" + instructions + "\n\n" +
		m.input.View() +
		"\n\nPress enter to request capability, or ctrl+c to quit.")
}

func main() {
	m := model{}
	m.input = textinput.New()
	m.input.Placeholder = "Enter capability name to request"
	m.input.Focus()

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Uh oh:", err)
		os.Exit(1)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
