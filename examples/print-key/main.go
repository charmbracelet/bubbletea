package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
)

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyboardEnhancementsMsg:
		return m, tea.Printf("Keyboard enhancements: EventTypes: %v\n",
			msg.SupportsEventTypes())
	case tea.KeyMsg:
		key := msg.Key()
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			}
		}
		format := "(%T) You pressed: %s"
		args := []any{msg, msg.String()}
		if len(key.Text) > 0 {
			format += " (text: %q)"
			args = append(args, key.Text)
		}
		return m, tea.Printf(format, args...)
	}
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView("Press any key to see its details printed to the terminal. Press 'ctrl+c' to quit.")
	v.KeyboardEnhancements.ReportEventTypes = true
	return v
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
	}
}
