package main

import (
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// this is an enum for Go
type sessionState uint

const (
	defaultTime              = time.Minute
	timerModel  sessionState = iota
	spinnerModel
)

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	current   = timerModel
	models    []tea.Model
)

type MainModel struct{}

// New: Create a new main model
func New(timeout time.Duration) MainModel {
	// initialize your model; timerView is the first "view" we want to see
	models = []tea.Model{}
	models = append(models, NewTimer(timeout))
	models = append(models, NewSpinner())
	m := MainModel{}
	return m
}

func (m MainModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(models[current].Init())
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return models[current].Update(msg)
}

func (m MainModel) View() string {
	var s string
	s += models[current].View() + "\n"
	s += helpStyle("right: next - left: prev - enter: new spinner")
	return s
}

func NextModel() {
	if int(current) == len(models)-1 {
		current = 0
	} else {
		current++
	}
}

func PrevModel() {
	if int(current) == 0 {
		current = sessionState(len(models) - 1)
	} else {
		current--
	}
}

func main() {
	p := tea.NewProgram(New(defaultTime))

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
