package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/calendar"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cal        calendar.Model
	chosenDate *time.Time
}

func NewModel() model {
	return model{
		cal: calendar.NewModel(),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "n":
			m.cal.CurrentDate = m.cal.CurrentDate.AddDate(0, -1, 0)
		case "p":
			m.cal.CurrentDate = m.cal.CurrentDate.AddDate(0, 1, 0)

		case "k":
			m.cal.CurrentDate = m.cal.CurrentDate.AddDate(0, 0, -7)
		case "j":
			m.cal.CurrentDate = m.cal.CurrentDate.AddDate(0, 0, 7)

		case "h":
			m.cal.CurrentDate = m.cal.CurrentDate.AddDate(0, 0, -1)
		case "l":
			m.cal.CurrentDate = m.cal.CurrentDate.AddDate(0, 0, 1)

		case "enter":
			m.chosenDate = &m.cal.CurrentDate

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	m.cal.Update(msg)

	return m, nil
}

func (m model) View() string {
	s := fmt.Sprintf("Date Picker:\n%s %d\n\n", m.cal.CurrentDate.Month(), m.cal.CurrentDate.Year())
	s += m.cal.View()

	if m.chosenDate != nil {
		s += fmt.Sprintf("You chose %s.\n", m.chosenDate.Format("02 January 2006"))
	}

	s += "\nPress hjkl to move around.\n"
	s += "Press q to quit.\n"

	return s
}

func main() {
	p := tea.NewProgram(NewModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
