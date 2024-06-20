package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	altscreen bool
	quitting  bool
	styles    *styles
}

type styles struct {
	keywordStyle lipgloss.Style
	helpStyle    lipgloss.Style
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	m.styles = &styles{
		keywordStyle: ctx.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235")),
		helpStyle:    ctx.NewStyle().Foreground(lipgloss.Color("241")),
	}
	return m, nil
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "space":
			var cmd tea.Cmd
			if m.altscreen {
				cmd = tea.ExitAltScreen
			} else {
				cmd = tea.EnterAltScreen
			}
			m.altscreen = !m.altscreen
			return m, cmd
		}
	}
	return m, nil
}

func (m model) View(ctx tea.Context) string {
	if m.quitting {
		return "Bye!\n"
	}

	const (
		altscreenMode = " altscreen mode "
		inlineMode    = " inline mode "
	)

	var mode string
	if m.altscreen {
		mode = altscreenMode
	} else {
		mode = inlineMode
	}

	return fmt.Sprintf("\n\n  You're in %s\n\n\n", m.styles.keywordStyle.Render(mode)) +
		m.styles.helpStyle.Render("  space: switch modes • q: exit\n")
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
