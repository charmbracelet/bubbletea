package main

// A simple example illustrating how to set a window title.

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type model struct{}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tea.SetWindowTitle("Bubble Tea Example")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyPressMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	return "\nPress any key to quit."
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
