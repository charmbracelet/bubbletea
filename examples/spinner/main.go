package main

// A simple program demonstrating the spinner component from the Bubbles
// component library.

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var term = termenv.ColorProfile()

type errMsg error

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initialModel() model {
	s := spinner.NewModel()
	s.Frames = spinner.Dot
	return model{spinner: s}
}

func (m model) Init() tea.Cmd {
	return spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			fallthrough
		case "esc":
			fallthrough
		case "ctrl+c":
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

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	s := termenv.
		String(m.spinner.View()).
		Foreground(term.Color("205")).
		String()
	str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", s)
	if m.quitting {
		return str + "\n"
	}
	return str
}
