package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type model bool

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		m = true
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() tea.View {
	if m {
		return tea.NewView("")
	}
	return tea.NewView("Press any key to quit.\n(When this program quits, it will vanish without a trace.)")
}

func main() {
	p := tea.NewProgram(model(false))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Oh no:", err)
	}
}
