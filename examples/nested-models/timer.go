package main

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Timer struct {
	timer timer.Model
}

func NewTimer(timeout time.Duration) *Timer {
	return &Timer{timer: timer.New(timeout)}
}

func (m Timer) Init() tea.Cmd {
	return m.timer.Init()
}

func (m Timer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l":
			return NextModel()
		case "left", "h":
			return PrevModel()
		}
	}
	m.timer, cmd = m.timer.Update(msg)
	return m, cmd
}

func (m Timer) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.timer.View(), HelpMenu())
}
