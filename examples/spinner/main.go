package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/boba"
	"github.com/charmbracelet/boba/spinner"
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
	p := boba.NewProgram(initialize, update, view, subscriptions)
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initialize() (boba.Model, boba.Cmd) {
	m := spinner.NewModel()
	m.Type = spinner.Dot

	return Model{
		spinner: m,
	}, nil
}

func update(msg boba.Msg, model boba.Model) (boba.Model, boba.Cmd) {
	m, ok := model.(Model)
	if !ok {
		return model, nil
	}

	switch msg := msg.(type) {

	case boba.KeyMsg:
		switch msg.String() {
		case "q":
			fallthrough
		case "esc":
			fallthrough
		case "ctrl+c":
			m.quitting = true
			return m, boba.Quit
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

func view(model boba.Model) string {
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

func subscriptions(model boba.Model) boba.Subs {
	m, ok := model.(Model)
	if !ok {
		return nil
	}

	sub, err := spinner.MakeSub(m.spinner)
	if err != nil {
		return nil
	}
	return boba.Subs{
		"tick": sub,
	}
}
