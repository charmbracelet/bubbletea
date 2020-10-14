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

var (
	color = termenv.ColorProfile()
)

type Model struct {
	spinner  spinner.Model
	quitting bool
	err      error
}

type errMsg error

func main() {
	p := tea.NewProgram(initialize, update, view)
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initialize() (tea.Model, tea.Cmd) {
	s := spinner.NewModel()
	s.Frames = spinner.Dot

	return Model{
		spinner: s,
	}, spinner.Tick(s)
}

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, ok := model.(Model)
	if !ok {
		return model, nil
	}

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
		m.spinner, cmd = spinner.Update(msg, m.spinner)
		return m, cmd
	}

}

func view(model tea.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "could not perform assertion on model in view\n"
	}
	if m.err != nil {
		return m.err.Error()
	}
	s := termenv.
		String(spinner.View(m.spinner)).
		Foreground(color.Color("205")).
		String()
	str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", s)
	if m.quitting {
		return str + "\n"
	}
	return str
}
