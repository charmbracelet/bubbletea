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

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
			return m, tea.Quit
		}

	case tea.MouseClickMsg, tea.MouseReleaseMsg, tea.MouseWheelMsg, tea.MouseMotionMsg:
		var mouse tea.Mouse
		switch msg := msg.(type) {
		case tea.MouseClickMsg:
			mouse = tea.Mouse(msg)
		case tea.MouseReleaseMsg:
			mouse = tea.Mouse(msg)
		case tea.MouseWheelMsg:
			mouse = tea.Mouse(msg)
		case tea.MouseMotionMsg:
			mouse = tea.Mouse(msg)
		}
		x, y := mouse.X, mouse.Y
		return m, tea.Printf("(X: %d, Y: %d) %s", x, y, msg)
	}

	return m, nil
}

func (m model) View() string {
	s := "Do mouse stuff. When you're done press q to quit.\n"

	return s
}
