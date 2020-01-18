package main

// A simple program that counts down from 5 and then exits.

import (
	"errors"
	"fmt"
	"log"

	"github.com/charmbracelet/tea"
	"github.com/charmbracelet/teaparty/input"
)

type Model struct {
	Input input.Model
	Error error
}

type tickMsg struct{}

func main() {
	tea.UseSysLog("tea")

	p := tea.NewProgram(
		Model{
			Input: input.DefaultModel(),
			Error: nil,
		},
		update,
		view,
		[]tea.Sub{
			// We just hand off the subscription to the input component, giving
			// it the model it expects.
			func(model tea.Model) tea.Msg {
				m, ok := model.(Model)
				if !ok {
					return tea.NewErrMsg("could not perform assertion on model")
				}
				return input.Blink(m.Input)
			},
		},
	)

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m, ok := model.(Model)
	if !ok {
		// When we encounter errors in Update we simply add the error to the
		// model so we can handle it in the view. We could also return a command
		// that does something else with the error, like logs it via IO.
		m.Error = errors.New("could not perform assertion on model")
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "break":
			fallthrough
		case "esc":
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case tea.ErrMsg:
		m.Error = msg
		return m, nil
	}

	m.Input, cmd = input.Update(msg, m.Input)
	return m, cmd
}

func view(model tea.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "Oh no: could not perform assertion on model."
	} else if m.Error != nil {
		return fmt.Sprintf("Uh oh: %s", m.Error)
	}
	return fmt.Sprintf(
		"What’s your favorite Pokémon?\n\n%s\n\n%s",
		input.View(m.Input),
		"(esc to quit)",
	)
}
