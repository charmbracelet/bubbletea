package main

import (
	"fmt"
	"tea"
)

type Model int

func main() {
	p := tea.NewProgram(0, update, view)
	p.Start()
}

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, _ := model.(Model)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg {

		case "j":
			m += 1
			if m > 3 {
				m = 3
			}

		case "k":
			m -= 1
			if m < 0 {
				m = 0
			}

		case "q":
			return m, tea.Quit

		}
	}

	return m, nil
}

func view(model tea.Model) string {
	m, _ := model.(Model)

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		checkbox("Plant carrots", m == 0),
		checkbox("Go to the market", m == 1),
		checkbox("Read something", m == 2),
		checkbox("See friends", m == 3),
	)

	return fmt.Sprintf(
		"What to do today?\n\n%s.\n\n(press j/k to select or q to quit)",
		choices,
	)
}

func checkbox(label string, checked bool) string {
	check := " "
	if checked {
		check = "x"
	}
	return fmt.Sprintf("[%s] %s", check, label)
}
