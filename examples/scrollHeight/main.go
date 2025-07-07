package main

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

type lines struct {
	n    int
	quit bool
}

func (l lines) Init() tea.Cmd {
	return nil
}

func (l lines) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up":
			if l.n > 1 {
				l.n--
			}
		case "down":
			if l.n < 100 {
				l.n++
			}
		case "q", "ctrl+c":
			l.quit = true
			return l, tea.Quit
		}
	}
	return l, nil
}

func (l lines) View() string {
	if l.quit {
		return ""
	}
	s := "This is the first line."
	for i := 1; i < int(l.n); i++ {
		s += "\n"
	}
	return s
}

func main() {
	l := lines{n: 5}

	p := tea.NewProgram(l)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
