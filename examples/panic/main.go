package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "x":
			return m, func() tea.Msg {
				panic("oh no!")
				return nil
			}
		case "y":
			panic("oh no!")
		}
	}
	return m, nil
}

func (m model) View() string {
	return "Hello, World!"
}

func main() {
	p := tea.NewProgram(model{})

	_, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
}
