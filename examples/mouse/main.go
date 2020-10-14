package main

// A simple program that opens the alternate screen buffer and displays mouse
// coordinates and events.

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialize, update, view)

	p.EnterAltScreen()
	defer p.ExitAltScreen()
	p.EnableMouseAllMotion()
	defer p.DisableMouseAllMotion()

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	init       bool
	mouseEvent tea.MouseEvent
}

func initialize() (tea.Model, tea.Cmd) {
	return model{}, nil
}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || (msg.Type == tea.KeyRune && msg.Rune == 'q') {
			return m, tea.Quit
		}

	case tea.MouseMsg:
		m.init = true
		m.mouseEvent = tea.MouseEvent(msg)
	}

	return m, nil
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)

	s := "Do mouse stuff. When you're done press q to quit.\n\n"

	if m.init {
		e := m.mouseEvent
		s += fmt.Sprintf("(X: %d, Y: %d) %s", e.X, e.Y, e)
	}

	return s
}
