package mouse

// A simple program that opens the alternate screen buffer and displays mouse
// coordinates and events.

import (
	"fmt"
	"log"

	tea "github.com/rprtr258/bubbletea"
)

func Main() {
	p := tea.NewProgram(model{}).WithAltScreen().WithMouseAllMotion()
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	init       bool
	mouseEvent tea.MouseEvent
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MsgKey:
		if s := msg.String(); s == "ctrl+c" || s == "q" || s == "esc" {
			return m, tea.Quit
		}

	case tea.MouseMsg:
		m.init = true
		m.mouseEvent = tea.MouseEvent(msg)
	}

	return m, nil
}

func (m model) View(r tea.Renderer) {
	s := "Do mouse stuff. When you're done press q to quit.\n\n"

	if m.init {
		e := m.mouseEvent
		s += fmt.Sprintf("(X: %d, Y: %d) %s", e.X, e.Y, e)
	}

	r.Write(s)
}
