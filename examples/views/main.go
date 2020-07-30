package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fogleman/ease"
)

func main() {
	p := tea.NewProgram(
		initialize,
		update,
		view,
	)
	if err := p.Start(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

// MESSAGES

type tickMsg struct{}
type frameMsg struct{}

// MODEL

type model struct {
	Choice   int
	Chosen   bool
	Ticks    int
	Frames   int
	Progress float64
	Loaded   bool
}

func initialize() (tea.Model, tea.Cmd) {
	return model{0, false, 10, 0, 0, false}, tick()
}

// CMDS

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

// UPDATE

// Main update function. This just hands off the message and model to the
// approprate update loop based on the current state.
func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	if !m.Chosen {
		return updateChoices(msg, m)
	}
	return updateChosen(msg, m)
}

// Update loop for the first view where you're choosing a task
func updateChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
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
		case "enter":
			m.Chosen = true
			return m, frame()
		case "q":
			fallthrough
		case "esc":
			fallthrough
		case "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		if m.Ticks == 0 {
			return m, tea.Quit
		}
		m.Ticks -= 1
		return m, tick()
	}

	return m, nil
}

// Update loop for the second view after a choice has been made
func updateChosen(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			fallthrough
		case "esc":
			fallthrough
		case "ctrl+c":
			return m, tea.Quit
		}

	case frameMsg:
		if !m.Loaded {
			m.Frames += 1
			m.Progress = ease.OutBounce(float64(m.Frames) / float64(100))
			if m.Progress >= 1 {
				m.Progress = 1
				m.Loaded = true
				m.Ticks = 3
				return m, tick()
			}
			return m, frame()
		}

	case tickMsg:
		if m.Loaded {
			if m.Ticks == 0 {
				return m, tea.Quit
			}
			m.Ticks -= 1
			return m, tick()
		}
	}

	return m, nil
}

// VIEW

// The main view, which just calls the approprate sub-view
func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	if !m.Chosen {
		return choicesView(m) + "\n"
	}
	return chosenView(m) + "\n"
}

// The first view, where you're choosing a task
func choicesView(m model) string {
	c := m.Choice

	tpl := "What to do today?\n\n"
	tpl += "%s\n\n"
	tpl += "Program quits in %d seconds\n\n"
	tpl += "(press j/k or up/down to select, enter to choose, and q or esc to quit)`"

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		checkbox("Plant carrots", c == 0),
		checkbox("Go to the market", c == 1),
		checkbox("Read something", c == 2),
		checkbox("See friends", c == 3),
	)

	return fmt.Sprintf(tpl, choices, m.Ticks)
}

// The second view, after a task has been chosen
func chosenView(m model) string {
	var msg string

	switch m.Choice {
	case 0:
		msg = "Carrot planting?\n\nCool, we'll need libgarden and vegeutils..."
	case 1:
		msg = "A trip to the market?\n\nOkay, then we should install marketkit and libshopping..."
	case 2:
		msg = "Reading time?\n\nOkay, cool, then we’ll need a library. Yes, an actual library."
	default:
		msg = "It’s always good to see friends.\n\nFetching social-skills and conversationutils..."
	}

	label := "Downloading..."
	if m.Loaded {
		label = fmt.Sprintf("Downloaded. Exiting in %d...", m.Ticks)
	}

	return msg + "\n\n " + label + "\n" + progressbar(80, m.Progress) + "%"
}

func checkbox(label string, checked bool) string {
	check := " "
	if checked {
		check = "x"
	}
	return fmt.Sprintf("[%s] %s", check, label)
}

func progressbar(width int, percent float64) string {
	metaChars := 7
	w := float64(width - metaChars)
	fullSize := int(math.Round(w * percent))
	emptySize := int(w) - fullSize
	fullCells := strings.Repeat("#", fullSize)
	emptyCells := strings.Repeat(".", emptySize)
	return fmt.Sprintf("|%s%s| %3.0f", fullCells, emptyCells, math.Round(percent*100))
}
