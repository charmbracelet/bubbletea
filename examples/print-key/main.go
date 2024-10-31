package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type model struct{}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyboardEnhancementsMsg:
		return m, tea.Printf("Keyboard enhancements enabled! ReleaseKeys: %v\n", msg.SupportsKeyReleases())
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

func (m model) View() string {
	return "Press any key to see its details printed to the terminal. Press 'ctrl+c' to quit."
}

func main() {
	p := tea.NewProgram(model{}, tea.WithKeyboardEnhancements(tea.WithKeyReleases, tea.WithUniformKeyLayout))
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
	}
}
