package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Available spinners
var spinners = []spinner.Spinner{
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

type styles struct {
	textStyle    lipgloss.Style
	spinnerStyle lipgloss.Style
	helpStyle    lipgloss.Style
}

func main() {
	m := model{}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

type model struct {
	index   int
	spinner spinner.Model
	styles  *styles
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	m.styles = &styles{
		textStyle:    ctx.NewStyle().Foreground(lipgloss.Color("252")),
		spinnerStyle: ctx.NewStyle().Foreground(lipgloss.Color("69")),
		helpStyle:    ctx.NewStyle().Foreground(lipgloss.Color("241")),
	}

	m.resetSpinner(ctx)

	return m, m.spinner.Tick
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "h", "left":
			m.index--
			if m.index < 0 {
				m.index = len(spinners) - 1
			}
			m.resetSpinner(ctx)
			return m, m.spinner.Tick
		case "l", "right":
			m.index++
			if m.index >= len(spinners) {
				m.index = 0
			}
			m.resetSpinner(ctx)
			return m, m.spinner.Tick
		default:
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(ctx, msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m *model) resetSpinner(ctx tea.Context) {
	m.spinner = spinner.New(ctx)
	m.spinner.Style = m.styles.spinnerStyle
	m.spinner.Spinner = spinners[m.index]
}

func (m model) View(ctx tea.Context) (s string) {
	var gap string
	switch m.index {
	case 1:
		gap = ""
	default:
		gap = " "
	}

	s += fmt.Sprintf("\n %s%s%s\n\n", m.spinner.View(ctx), gap, m.styles.textStyle.Render("Spinning..."))
	s += m.styles.helpStyle.Render("h/l, ←/→: change spinner • q: exit\n")
	return
}
