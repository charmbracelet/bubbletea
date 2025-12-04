package main

import (
	tea "charm.land/bubbletea/v2"
)

// This example demonstrates using Bubbletea without visual output.
// By using WithoutOutput(), the renderer is disabled but raw mode is still
// enabled for proper keyboard input handling (arrow keys work without Enter).
// Output processing is preserved so that println and logging libraries work
// correctly without needing manual "\r\n" handling.
//
// This is useful for non-TUI applications that still need Bubbletea's
// event handling and state management but want to use standard output
// functions like println or logging libraries.
func main() {
	p := tea.NewProgram(model(5), tea.WithoutOutput())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

type model int

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			m++
			println(m)
		case "down":
			m--
			println(m)
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	return tea.NewView("")
}
