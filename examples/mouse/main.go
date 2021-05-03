package main

// A simple program that opens the alternate screen buffer and displays mouse
// coordinates and events.

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := tea.NewProgram(model{}).Start(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	init       bool
	mouseEvent tea.MouseEvent
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tea.EnableMouseAllMotion)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
			return m, tea.Quit
		}

	case tea.MouseMsg:
		m.init = true
		m.mouseEvent = tea.MouseEvent(msg)
	}

	return m, nil
}

func (m model) View() string {
	s := "Do mouse stuff. When you're done press q to quit.\n\n"

	if m.init {
		e := m.mouseEvent
		s += fmt.Sprintf("(X: %d, Y: %d) %s", e.X, e.Y, e)
	}

	return s
}
