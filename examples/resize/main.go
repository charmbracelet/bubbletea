package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	windowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}).
			Align(lipgloss.Center)

	keywordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F25D94"))
)

type model struct {
	width, height int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// The window dimensions are sent to update when the program first
		// starts as well as after a resize.
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			return m, nil
		}
	default:
		return m, nil
	}
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Waiting for dimensions..."
	}

	s := fmt.Sprintf(
		"Window is %s x %s cells.\n\nResize to update. Press q to exit.",
		keywordStyle.Render(strconv.Itoa(m.width)),
		keywordStyle.Render(strconv.Itoa(m.height)),
	)
	s = strings.Repeat("\n", m.height/2-lipgloss.Height(s)) + s

	return windowStyle.Copy().
		Width(m.width - windowStyle.GetHorizontalBorderSize()).
		Height(m.height - windowStyle.GetVerticalBorderSize()).
		Render(s)
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
