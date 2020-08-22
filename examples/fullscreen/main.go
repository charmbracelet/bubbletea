package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model int

type tickMsg time.Time

func main() {
	p := tea.NewProgram(initialize, update, view)

	p.EnterAltScreen()
	err := p.Start()
	p.ExitAltScreen()

	if err != nil {
		log.Fatal(err)
	}
}

func initialize() (tea.Model, tea.Cmd) {
	return model(5), tick()
}

func update(message tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg := message.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			fallthrough
		case "esc":
			fallthrough
		case "q":
			return m, tea.Quit
		}

	case tickMsg:
		m -= 1
		if m <= 0 {
			return m, tea.Quit
		}
		return m, tick()

	}

	return m, nil
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	return fmt.Sprintf("\n\n     Hi. This program will exit in %d seconds...", m)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
