package main

// The tag is incremented with each keypress which sends that tag along in a Msg after a `tea.Tick` delay
// If the tag in the Msg matches the model's state, we can be sure that the debouncing is complete and that we can return our command.
// This is useful in cases such as the Spinner bubble to prevent accidentally spinning it too fast.
// https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type exitMsg int

type model struct {
	tag int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.tag++
		return m, tea.Tick(time.Second, func(_ time.Time) tea.Msg {
			return exitMsg(m.tag)
		})
	case exitMsg:
		if int(msg) == m.tag {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	return "To exit press any key, then wait for one second without pressing anything."
}

func main() {
	if err := tea.NewProgram(model{}).Start(); err != nil {
		fmt.Println("uh oh:", err)
		os.Exit(1)
	}
}
