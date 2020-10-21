package main

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"os"

	input "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	te "github.com/muesli/termenv"
)

const focusedTextColor = "205"

var (
	color               = te.ColorProfile().Color
	focusedPrompt       = te.String("> ").Foreground(color("205")).String()
	blurredPrompt       = "> "
	focusedSubmitButton = "[ " + te.String("Submit").Foreground(color("205")).String() + " ]"
	blurredSubmitButton = "[ " + te.String("Submit").Foreground(color("240")).String() + " ]"
)

func main() {
	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}

type model struct {
	index         int
	nameInput     input.Model
	nickNameInput input.Model
	emailInput    input.Model
	submitButton  string
}

func initialModel() model {
	name := input.NewModel()
	name.Placeholder = "Name"
	name.Focus()
	name.Prompt = focusedPrompt
	name.TextColor = focusedTextColor
	name.CharLimit = 32

	nickName := input.NewModel()
	nickName.Placeholder = "Nickname"
	nickName.Prompt = blurredPrompt
	nickName.CharLimit = 32

	email := input.NewModel()
	email.Placeholder = "Email"
	email.Prompt = blurredPrompt
	email.CharLimit = 64

	return model{0, name, nickName, email, blurredSubmitButton}

}
func (m model) Init() tea.Cmd {
	return tea.Batch(
		input.Blink(m.nameInput),
		input.Blink(m.nickNameInput),
		input.Blink(m.emailInput),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit

		// Cycle between inputs
		case "tab", "shift+tab", "enter", "up", "down":

			inputs := []input.Model{
				m.nameInput,
				m.nickNameInput,
				m.emailInput,
			}

			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.index == len(inputs) {
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.index--
			} else {
				m.index++
			}

			if m.index > len(inputs) {
				m.index = 0
			} else if m.index < 0 {
				m.index = len(inputs)
			}

			for i := 0; i <= len(inputs)-1; i++ {
				if i == m.index {
					// Set focused state
					inputs[i].Focus()
					inputs[i].Prompt = focusedPrompt
					inputs[i].TextColor = focusedTextColor
					continue
				}
				// Remove focused state
				inputs[i].Blur()
				inputs[i].Prompt = blurredPrompt
				inputs[i].TextColor = ""
			}

			m.nameInput = inputs[0]
			m.nickNameInput = inputs[1]
			m.emailInput = inputs[2]

			if m.index == len(inputs) {
				m.submitButton = focusedSubmitButton
			} else {
				m.submitButton = blurredSubmitButton
			}

			return m, nil
		}
	}

	// Handle character input and blinks
	m, cmd = updateInputs(msg, m)
	return m, cmd
}

// Pass messages and models through to text input components. Only text inputs
// with Focus() set will respond, so it's safe to simply update all of them
// here without any further logic.
func updateInputs(msg tea.Msg, m model) (model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.nameInput, cmd = input.Update(msg, m.nameInput)
	cmds = append(cmds, cmd)

	m.nickNameInput, cmd = input.Update(msg, m.nickNameInput)
	cmds = append(cmds, cmd)

	m.emailInput, cmd = input.Update(msg, m.emailInput)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	s := "\n"

	inputs := []string{
		input.View(m.nameInput),
		input.View(m.nickNameInput),
		input.View(m.emailInput),
	}

	for i := 0; i < len(inputs); i++ {
		s += inputs[i]
		if i < len(inputs)-1 {
			s += "\n"
		}
	}

	s += "\n\n" + m.submitButton + "\n"
	return s
}
