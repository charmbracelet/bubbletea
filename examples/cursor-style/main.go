package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type model struct {
	cursor tea.Cursor
	blink  bool
}

func (m model) Init() (tea.Model, tea.Cmd) {
	m.blink = true
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "h", "left":
			m.cursor.Shape--
			if m.cursor.Shape < tea.CursorBlock {
				m.cursor.Shape = tea.CursorBar
			}
		case "l", "right":
			m.cursor.Shape++
			if m.cursor.Shape > tea.CursorBar {
				m.cursor.Shape = tea.CursorBlock
			}
		}
	}
	m.blink = !m.blink
	return m, nil
}

func (m model) View() fmt.Stringer {
	f := tea.NewFrame("Press left/right to change the cursor style, q or ctrl+c to quit." +
		"\n\n" +
		"  <- This is the cursor (a " + m.describeCursor() + ")")
	f.Cursor = tea.NewCursor(0, 2)
	f.Cursor.Shape = m.cursor.Shape
	f.Cursor.Blink = m.blink
	return f
}

func (m model) describeCursor() string {
	var adj, noun string

	if m.blink {
		adj = "blinking"
	} else {
		adj = "steady"
	}

	switch m.cursor.Shape {
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
