package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	input textinput.Model
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.input.View()
}

func main() {
	var m model
	m.input = textinput.New()
	m.input.Placeholder = "What are your favorite fruits?"
	m.input.Focus()
	m.input.Prompt = "Fruits> "
	m.input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	m.input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
	m.input.ShowSuggestions = true
	m.input.SetSuggestions([]string{
		"apple",
		"banana",
		"cherry",
		"date",
		"elderberry",
		"fig",
		"grape",
		"honeydew",
		"kiwi",
		"lemon",
		"mango",
		"nectarine",
		"orange",
		"pear",
		"quince",
		"raspberry",
		"strawberry",
		"tangerine",
		"ugli fruit",
		"vanilla bean",
		"watermelon",
		"ximenia caffra",
		"yellow passion fruit",
		"zucchini",
	})

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v", err)
	}
}
