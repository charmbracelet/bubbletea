package main

// A simple program that queries and displays the window-size.

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model{})
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
	case tea.KeyMsg:
		if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
			return m, tea.Quit
		}

		return m, tea.WindowSize()

	case tea.WindowSizeMsg:
		return m, tea.Printf("%dx%d", msg.Width, msg.Height)
	}

	return m, nil
}

func (m model) View() string {
	s := "When you're done press q to quit. Press any other key to query the window-size.\n"

	return s
}
