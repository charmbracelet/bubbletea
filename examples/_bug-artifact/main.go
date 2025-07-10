package main

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var (
	PrimaryColour = lipgloss.Color("#3a1577")

	Banner = lipgloss.NewStyle().Background(PrimaryColour).Height(2)
)

type model struct {
	width int
	text  string
}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(Banner.Copy().Width(m.width).Render(m.text))
	return b.String()
}

func main() {
	m := model{
		text: "hello!",
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}
