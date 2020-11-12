package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}

	color     = termenv.ColorProfile().Color
	textStyle = termenv.Style{}.Foreground(color("252")).Styled
	helpStyle = termenv.Style{}.Foreground(color("241")).Styled
)

func main() {
	m := model{}
	m.resetSpinner()

	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

type model struct {
	index   int
	spinner spinner.Model
}

func (m model) Init() tea.Cmd {
	return spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			m.index--
			if m.index <= 0 {
				m.index = len(spinners) - 1
			}
			m.resetSpinner()
			return m, nil
		case "l", "right":
			m.index++
			if m.index >= len(spinners) {
				m.index = 0
			}
			m.resetSpinner()
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m *model) resetSpinner() {
	m.spinner = spinner.NewModel()
	m.spinner.ForegroundColor = "69"
	m.spinner.Spinner = spinners[m.index]
}

func (m model) View() (s string) {
	gap := " "
	switch m.index {
	case 1:
		gap = ""
	default:
		gap = " "
	}
	s += fmt.Sprintf("\n %s%s%s\n\n", m.spinner.View(), gap, textStyle("Spinning..."))
	s += helpStyle("j/k, ←/→: change spinner • q: exit\n")
	return
}
