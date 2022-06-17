package main

/*
With each keypress you bump the tag on the `model` and send that tag along in a `Msg` after delaying with a `tea.Tick`.
If the tag in the `Msg` matches the one on the model you know the debouncing is done and you can then return your command

We do this in the [spinner](https://github.com/charmbracelet/bubbles/blob/master/spinner/spinner.go) Bubble to keep it from spinning out of control in the event that it accidentally gets too many spin Cmds.
*/

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
