package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"tea"

	"tea/input"
)

type model struct {
	input input.Model
}

type tickMsg struct{}

func main() {
	tea.UseSysLog("tea")

	p := tea.NewProgram(
		model{
			input: input.DefaultModel(),
		},
		update,
		view,
		[]tea.Sub{
			// Just hand off the subscription to the input component
			func(m tea.Model) tea.Msg {
				if m, ok := m.(model); ok {
					return input.Blink(m.input)
				}
				// TODO: return error
				return nil
			},
		},
	)

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m, _ := mdl.(model)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "break":
			fallthrough
		case "esc":
			return m, tea.Quit
		}
	}

	m.input, cmd = input.Update(msg, m.input)
	return m, cmd
}

func view(m tea.Model) string {
	if m, ok := m.(model); ok {
		help := "(esc to exit)"

		return fmt.Sprintf(
			"What’s your favorite Pokémon?\n\n%s\n\n%s",
			input.View(m.input),
			help,
		)
	}
	// TODO: return error
	return ""
}
