package main

// A simple program that opens the alternate screen buffer and displays mouse
// coordinates and events.

import (
	"log"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	p := tea.NewProgram(model{}, tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct{}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
			return m, tea.Quit
		}

	case tea.MouseMsg:
		mouse := msg.Mouse()
		return m, tea.Printf("(X: %d, Y: %d) %s", mouse.X, mouse.Y, mouse)
	}

	return m, nil
}

func (m model) View() string {
	s := "Do mouse stuff. When you're done press q to quit.\n"

	return s
}
