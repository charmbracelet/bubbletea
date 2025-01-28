package main

// A simple example illustrating how to run a series of commands in order.

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type model struct{}

func (m model) Init() (model, tea.Cmd) {
	// A tea.Sequence is a command that runs a series of commands in
	// order. Contrast this with tea.Batch, which runs a series of commands
	// concurrently, with no order guarantees.
	return m, tea.Sequence(
		tea.Batch(
			// These will always resolve first, in any order.
			tea.Println("A"),
			tea.Println("B"),
			tea.Println("C"),
		),
		// This will always resolve last.
		tea.Println("Z"),
		tea.Quit,
	)
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyPressMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() fmt.Stringer {
	return tea.NewFrame("")
}

func main() {
	if err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
