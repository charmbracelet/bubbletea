package main

// A simple example illustrating how to run a series of commands in order.

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type model struct{}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tea.Sequence(
		tea.Batch(
			tea.Println("A"),
			tea.Println("B"),
			tea.Println("C"),
		),
		tea.Println("Z"),
		tea.Quit,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyPressMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	return ""
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
