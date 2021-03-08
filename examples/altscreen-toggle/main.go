package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var (
	color   = termenv.ColorProfile().Color
	keyword = termenv.Style{}.Foreground(color("204")).Background(color("235")).Styled
)

type model struct {
	altscreen bool
	quitting  bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case " ":
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

func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	const (
		altscreenMode = " altscreen mode "
		inlineMode    = " inline mode "
	)

	var mode, otherMode string
	if m.altscreen {
		mode = altscreenMode
		otherMode = inlineMode
	} else {
		mode = inlineMode
		otherMode = altscreenMode
	}

	return fmt.Sprintf(
		"\n  You're in %s. Press %s to swich to %s.\n\n  To exit press %s.\n",
		keyword(mode), keyword(" space "), keyword(otherMode), keyword(" q "),
	)
}

func main() {
	if err := tea.NewProgram(model{}).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
