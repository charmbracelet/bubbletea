package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/tea"
)

type model int

type tickMsg struct{}

func main() {
	tea.Fullscreen()
	defer tea.ExitFullscreen()
	err := tea.NewProgram(initialize, update, view, subscriptions).Start()
	if err != nil {
		log.Fatal(err)
	}
}

func initialize() (tea.Model, tea.Cmd) {
	return model(5), nil
}

func update(message tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg := message.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "break":
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

	}

	return m, nil
}

func subscriptions(_ tea.Model) tea.Subs {
	return tea.Subs{
		"tick": func(_ tea.Model) tea.Msg {
			time.Sleep(time.Second)
			return tickMsg{}
		},
	}
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	return fmt.Sprintf("\n\n     Hi. This program will exit in %d seconds...", m)
}
