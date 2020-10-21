package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"
	"log"

	input "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

type tickMsg struct{}
type errMsg error

type model struct {
	textInput input.Model
	err       error
}

func initialModel() model {
	inputModel := input.NewModel()
	inputModel.Placeholder = "Pikachu"
	inputModel.Focus()
	inputModel.CharLimit = 20

	return model{
		textInput: inputModel,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return input.Blink(m.textInput)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			fallthrough
		case tea.KeyEsc:
			fallthrough
		case tea.KeyEnter:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = input.Update(msg, m.textInput)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"What’s your favorite Pokémon?\n\n%s\n\n%s",
		input.View(m.textInput),
		"(esc to quit)",
	) + "\n"
}
