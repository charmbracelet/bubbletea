package main

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

const (
	padding  = 2
	maxWidth = 80
)

type progressMsg float64

type progressErrMsg struct{ err error }

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

type model struct {
	pw       *progressWriter
	progress progress.Model
	err      error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.progress.SetWidth(msg.Width - padding*2 - 4)
		if m.progress.Width() > maxWidth {
			m.progress.SetWidth(maxWidth)
		}
		return m, nil

	case progressErrMsg:
		m.err = msg.err
		return m, tea.Quit

	case progressMsg:
		var cmds []tea.Cmd

		if msg >= 1.0 {
			cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
		}

		cmds = append(cmds, m.progress.SetPercent(float64(msg)))
		return m, tea.Batch(cmds...)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() tea.View {
	if m.err != nil {
		return tea.NewView("Error downloading: " + m.err.Error() + "\n")
	}

	pad := strings.Repeat(" ", padding)
	return tea.NewView("\n" +
		pad + m.progress.View() + "\n\n" +
		pad + helpStyle("Press any key to quit"))
}
