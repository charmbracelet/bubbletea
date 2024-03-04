package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
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
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
)

type Spinner struct {
	spinner spinner.Model
	index   int
}

func NewSpinner() *Spinner {
	return &Spinner{spinner: spinner.New()}
}

func (m Spinner) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Spinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			m.Next()
			m.resetSpinner()
			return m, m.spinner.Tick
		case "right", "l":
			return NextModel()
		case "left", "h":
			return PrevModel()
		}
	}
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m Spinner) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, spinnerStyle.Render(m.spinner.View()), HelpMenu("spinner"))
}

func (m *Spinner) resetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinners[m.index]
}

func (m *Spinner) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func (m *Spinner) Prev() {
	if m.index == 0 {
		m.index = len(spinners) - 1
	} else {
		m.index--
	}
}
