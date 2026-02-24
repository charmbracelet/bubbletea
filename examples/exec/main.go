package main

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

type editorFinishedMsg struct{ err error }

func openEditor() tea.Cmd {
	editor := cmp.Or(os.Getenv("EDITOR"), "vim")
	c := exec.Command(editor) //nolint:gosec
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

type model struct {
	altscreenActive bool
	err             error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "a":
			m.altscreenActive = !m.altscreenActive
			return m, nil
		case "e":
			return m, openEditor()
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case editorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	if m.err != nil {
		v := tea.NewView("Error: " + m.err.Error() + "\n")
		v.AltScreen = m.altscreenActive
		return v
	}
	v := tea.NewView("Press 'e' to open your EDITOR.\nPress 'a' to toggle the altscreen\nPress 'q' to quit.\n")
	v.AltScreen = m.altscreenActive
	return v
}

func main() {
	m := model{}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
