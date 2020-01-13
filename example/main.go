package main

import (
	"fmt"
	"tea"
	"time"
)

type Model struct {
	Choice int
	Ticks  int
}

type TickMsg struct{}

const tpl = `What to do today?

%s

Elapsed: %d seconds.

(press j/k or up/down to select, q or esc to quit)`

func main() {
	p := tea.NewProgram(Model{0, 0}, update, view, []tea.Sub{tick})
	if err := p.Start(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, _ := model.(Model)

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg {
		case "j":
			fallthrough
		case "down":
			m.Choice += 1
			if m.Choice > 3 {
				m.Choice = 3
			}
		case "k":
			fallthrough
		case "up":
			m.Choice -= 1
			if m.Choice < 0 {
				m.Choice = 0
			}
		case "q":
			fallthrough
		case "esc":
			fallthrough
		case "ctrl+c":
			return m, tea.Quit
		}

	case TickMsg:
		m.Ticks += 1
	}

	return m, nil
}

// Subscription
func tick(_ tea.Model) tea.Msg {
	time.Sleep(time.Second * 1)
	return TickMsg{}
}

func view(model tea.Model) string {
	m, _ := model.(Model)
	c := m.Choice

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		checkbox("Plant carrots", c == 0),
		checkbox("Go to the market", c == 1),
		checkbox("Read something", c == 2),
		checkbox("See friends", c == 3),
	)

	return fmt.Sprintf(tpl, choices, m.Ticks)
}

func checkbox(label string, checked bool) string {
	check := " "
	if checked {
		check = "x"
	}
	return fmt.Sprintf("[%s] %s", check, label)
}
