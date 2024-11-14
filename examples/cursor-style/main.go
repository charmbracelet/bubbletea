package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type model struct {
	style  tea.CursorStyle
	steady bool
}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tea.Batch(tea.ShowCursor, tea.SetCursorStyle(m.style, m.steady))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+q", "q":
			return m, tea.Quit
		case "h", "left":
			if m.style == tea.CursorBlock && !m.steady {
				break
			}
			if !m.steady {
				m.style--
			}
			m.steady = !m.steady
			cmd = tea.SetCursorStyle(m.style, m.steady)
		case "l", "right":
			if m.style == tea.CursorBar && m.steady {
				break
			}
			if m.steady {
				m.style++
			}
			m.steady = !m.steady
			cmd = tea.SetCursorStyle(m.style, m.steady)
		}
	}
	return m, cmd
}

func (m model) View() string {
	return "Press left/right to change the cursor style, q or ctrl+c to quit." +
		"\n\n" +
		"  <- This is the cursor (a " + m.describeCursor() + ")"
}

func (m model) describeCursor() string {
	var (
		adj, noun string
	)

	if m.steady {
		adj = "steady"
	} else {
		adj = "blinking"
	}

	switch m.style {
	case tea.CursorBlock:
		noun = "block"
	case tea.CursorUnderline:
		noun = "underline"
	case tea.CursorBar:
		noun = "bar"
	}

	return fmt.Sprintf("%s %s", adj, noun)
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
