package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	err       error
	textInput textinput.Model
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	ti := textinput.New(ctx)
	ti.Placeholder = "Pikachu"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	m.textInput = ti
	return m, textinput.Blink
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "ctrl+c", "esc":
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(ctx, msg)
	return m, cmd
}

func (m model) View(ctx tea.Context) string {
	return fmt.Sprintf(
		"What’s your favorite Pokémon?\n\n%s\n\n%s",
		m.textInput.View(ctx),
		"(esc to quit)",
	) + "\n"
}
