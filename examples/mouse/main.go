package main

// A simple program that opens the alternate screen buffer and displays mouse
// coordinates and events.

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model{}, tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	mouseEvent tea.MouseEvent
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
			return m, tea.Quit
		}

	case tea.MouseDownMsg, tea.MouseUpMsg, tea.MouseWheelMsg, tea.MouseMotionMsg:
		var x, y int
		switch msg := msg.(type) {
		case tea.MouseDownMsg:
			x, y = msg.X, msg.Y
		case tea.MouseUpMsg:
			x, y = msg.X, msg.Y
		case tea.MouseWheelMsg:
			x, y = msg.X, msg.Y
		case tea.MouseMotionMsg:
			x, y = msg.X, msg.Y
		}
		return m, tea.Printf("(X: %d, Y: %d) %s", x, y, msg)
	}

	return m, nil
}

func (m model) View() string {
	s := "Do mouse stuff. When you're done press q to quit.\n"

	return s
}
