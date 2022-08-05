package main

import (
	"log"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// this is an enum for Go
type sessionState uint

const (
	defaultTime              = time.Minute
	timerView   sessionState = iota
	spinnerView
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

type mainModel struct {
	state   sessionState
	timer   timer.Model
	spinner spinner.Model
	index   int
}

// New: Create a new main model
func New(timeout time.Duration) mainModel {
	// initialize your model; timerView is the first "view" we want to see
	m := mainModel{state: timerView}
	m.timer = timer.New(timeout)
	m.spinner = spinner.New()
	return m
}

func (m mainModel) Init() tea.Cmd {
	// start the timer and spinner on program start
	return tea.Batch(m.timer.Init(), m.spinner.Tick)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle IO -> keypress, WindowSizeMSg
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if m.state == timerView {
				m.state = spinnerView
			} else {
				m.state = timerView
			}
			return m, nil
		case "n":
			var cmd tea.Cmd
			if m.state == timerView {
				m.timer = timer.New(defaultTime)
				cmd = m.timer.Init()
			} else {
				m.Next()
				m.resetSpinner()
				cmd = spinner.Tick
			}
			return m, cmd
		}
		switch m.state {
		// update whichever model is focused
		case spinnerView:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		default:
			// another way to do it
			newModel, cmd := m.timer.Update(msg)
			m.timer = newModel
			// if this were your own model, you would need to wrap the type
			// before assignment because its type would be tea.Model
			// i.e. m.list = newModel.(list.Model)
			return m, cmd
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m mainModel) View() string {
	var s string
	switch m.state {
	case spinnerView:
		s += m.spinner.View() + "\n"
	default:
		s += m.timer.View() + "\n"
	}
	s += helpStyle("enter: change view • n: new spinner/timer • q: exit\n")
	return s
}

func (m *mainModel) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func (m *mainModel) resetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinners[m.index]
}

func main() {
	p := tea.NewProgram(New(defaultTime))

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
