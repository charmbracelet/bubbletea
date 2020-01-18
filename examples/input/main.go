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
		nil,
		//[]tea.Sub{input.Blink},
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

	case input.CursorBlinkMsg:
		return input.Update(msg, m.input)
	}

	m.input, cmd = input.Update(msg, m.input)
	return m, cmd
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	help := "(esc to exit)"

	return fmt.Sprintf(
		"What’s your favorite Pokémon?\n\n%s\n\n%s",
		input.View(m.input),
		help,
	)
}
