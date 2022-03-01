package main

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	p *tea.Program
)

type model struct {
	altscreenActive bool
	err             error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			m.altscreenActive = !m.altscreenActive
			cmd := tea.EnterAltScreen
			if !m.altscreenActive {
				cmd = tea.ExitAltScreen
			}
			return m, cmd
		case "e":
			if err := p.ReleaseTerminal(); err != nil {
				m.err = err
				return m, nil
			}

			c := exec.Command(os.Getenv("EDITOR")) //nolint:gosec
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			m.err = c.Run()

			if err := p.RestoreTerminal(); err != nil {
				m.err = err
			}

			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error()
	}
	return "Press 'e' to open your EDITOR.\nPress 'a' to toggle the altscreen\nPress 'q' to quit.\n"
}

func main() {
	m := model{}
	p = tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
