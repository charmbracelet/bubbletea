package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type model struct {
	n int
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "j":
			m.n += 1
		case "k":
			m.n -= 1
		}
	}
	return m, nil
}

func (m model) View() string {
	b := strings.Builder{}
	for i := 0; i < m.n; i++ {
		b.WriteString(fmt.Sprintf("%d. line\n", i))
	}
	return fmt.Sprintf("FirstLine\n%sLastLine", b.String())
}

func main() {
	if _, err := tea.NewProgram(model{n: 9}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func init() {
	_, _ = tea.LogToFile("tea.log", "")
}
