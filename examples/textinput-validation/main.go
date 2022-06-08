package main

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var alphabetRegex = regexp.MustCompile("^[a-zA-Z]+$")
var alphanumericRegex = regexp.MustCompile("^[a-zA-Z0-9]+$")

func main() {
	p := tea.NewProgram(initialModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

// validateAge implements the textinput.ValidateFunc signature
func validateAge(s string) error {
	// For an age to be valid, it must:
	//   1. Be a number
	//   2. Be between 0 and 100

	age, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		return errors.New("Age must be a number")
	}

	if strings.HasPrefix(s, "0") || age < 0 || age > 100 {
		return errors.New("Age must be between 0 and 100")
	}

	return nil
}

type model struct {
	input textinput.Model
	err   error
}

func initialModel() model {
	m := model{
		input: textinput.New(),
	}

	var ti textinput.Model
	ti = textinput.New()
	ti.Placeholder = "Enter your age"
	ti.Validate = validateAge
	ti.Focus()

	m.input = ti
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	v := "How old are you?"
	v += "\n" + m.input.View() + "\n"
	if m.input.Err != nil {
		v += lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("\nError: " + m.input.Err.Error())
	}
	return v
}
