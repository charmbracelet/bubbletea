package main

// A simple program demonstrating the spinner component from the Bubbles
// component library.

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type errMsg error

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s}
}

func (m model) Init() (model, tea.Cmd) {
	return m, m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() fmt.Stringer {
	if m.err != nil {
		return tea.NewFrame(m.err.Error())
	}
	str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", m.spinner.View())
	if m.quitting {
		return tea.NewFrame(str + "\n")
	}
	return tea.NewFrame(str)
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
