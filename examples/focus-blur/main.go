package main

// A simple program that handled losing and acquiring focus.

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model{
		// assume we start focused...
		focused:   true,
		reporting: true,
	}, tea.WithReportFocus())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	focused   bool
	reporting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.FocusMsg:
		m.focused = true
	case tea.BlurMsg:
		m.focused = false
	case tea.KeyMsg:
		switch msg.String() {
		case "t":
			m.reporting = !m.reporting
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "Hi. Focus report is currently "
	if m.reporting {
		s += "enabled"
	} else {
		s += "disabled"
	}
	s += ".\n\n"

	if m.reporting {
		if m.focused {
			s += "This program is currently focused!"
		} else {
			s += "This program is currently blurred!"
		}
	}
	return s + "\n\nTo quit sooner press ctrl-c, or t to toggle focus reporting...\n"
}
