package main

// A simple example illustrating how to run a series of commands in order.

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd {
	return tea.Sequence(
		tea.Println("A"),
		tea.Println("B"),
		tea.Println("C"),
		tea.Quit,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	return ""
}

func main() {
	if err := tea.NewProgram(model{}).Start(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
