package main

// A simple example illustrating how to set a window title.

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

const windowTitle = "Hello, Bubble Tea"

type model struct{}

func (m model) Init() (model, tea.Cmd) {
	return m, tea.SetWindowTitle(windowTitle)
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyPressMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() fmt.Stringer {
	wrap := lipgloss.NewStyle().Width(78).Render
	return tea.NewFrame(wrap("The window title has been set to '"+windowTitle+"'. It will be cleared on exit.") +
		"\n\nPress any key to quit.")
}

func main() {
	if err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
