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
	err error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
	return "Press e to open Vim. Press q to quit."
}

func main() {
	m := model{}
	p = tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
