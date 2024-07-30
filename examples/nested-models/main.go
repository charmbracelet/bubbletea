package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// this is an enum for Go
type sessionState int

const (
	defaultTime = time.Minute
	first       = 0
)

const (
	timerModel sessionState = iota
	spinnerModel
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	current   = timerModel
	models    []tea.Model
)

func HelpMenu(view ...string) string {
	if len(view) != 0 {
		return helpStyle(fmt.Sprintf("right/l: next • left/h: previous • enter: new %s", view[first]))
	}
	return helpStyle("right/l: next • left/h: previous")
}

func NextModel() (tea.Model, tea.Cmd) {
	if int(current) == len(models)-1 {
		current = first
	} else {
		current++
	}
	return models[current], models[current].Init()
}

func PrevModel() (tea.Model, tea.Cmd) {
	if int(current) == first {
		current = sessionState(len(models) - 1)
	} else {
		current--
	}
	return models[current], models[current].Init()
}

func main() {
	models = []tea.Model{}
	models = append(models, NewTimer(defaultTime))
	models = append(models, NewSpinner())

	p := tea.NewProgram(models[current])

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
