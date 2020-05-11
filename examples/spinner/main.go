package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/tea"
	"github.com/charmbracelet/teaparty/spinner"
	"github.com/muesli/termenv"
)

var (
	color = termenv.ColorProfile()
)

type Model struct {
	spinner spinner.Model
	err     error
}

type errMsg error

func main() {
	p := tea.NewProgram(initialize, update, view, subscriptions)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func initialize() (tea.Model, tea.Cmd) {
	m := spinner.NewModel()
	m.Type = spinner.Dot

	return Model{
		spinner: m,
	}, nil
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
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	default:
		m.spinner, _ = spinner.Update(msg, m.spinner)
		return m, nil
	}

}

func view(model tea.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "could not perform assertion on model in view"
	}
	if m.err != nil {
		return m.err.Error()
	}
	s := termenv.
		String(spinner.View(m.spinner)).
		Foreground(color.Color("205")).
		String()
	return fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", s)
}

func subscriptions(model tea.Model) tea.Subs {
	m, ok := model.(Model)
	if !ok {
		return nil
	}

	sub, err := spinner.MakeSub(m.spinner)
	if err != nil {
		return nil
	}
	return tea.Subs{
		"tick": sub,
	}
}
