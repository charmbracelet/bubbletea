package main

// TODO: This code feels messy. Clean it up.

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/tea"
	"github.com/fogleman/ease"
)

// Model contains the data for our application.
type Model struct {
	Choice   int
	Chosen   bool
	Ticks    int
	Frames   int
	Progress float64
	Loaded   bool
}

type tickMsg struct{}

type frameMsg struct{}

func main() {
	p := tea.NewProgram(
		initialize,
		update,
		view,
		subscriptions,
	)
	if err := p.Start(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

// INIT

func initialize() (tea.Model, tea.Cmd) {
	return Model{0, false, 10, 0, 0, false}, nil
}

// SUBSCRIPTIONS

func subscriptions(model tea.Model) tea.Subs {
	m, _ := model.(Model)
	if !m.Chosen || m.Loaded {
		return tea.Subs{
			"tick": tick,
		}
	}
	return tea.Subs{
		"frame": frame,
	}
}

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}

func frame() tea.Msg {
	time.Sleep(time.Second / 60)
	return frameMsg{}
}

// UPDATES

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, _ := model.(Model)

	if !m.Chosen {
		return updateChoices(msg, m)
	}
	return updateChosen(msg, m)
}

func updateChoices(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
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
			return m, nil
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
	}

	return m, nil
}

func updateChosen(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
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
			m.Progress = ease.OutBounce(float64(m.Frames) / float64(120))
			if m.Progress >= 1 {
				m.Progress = 1
				m.Loaded = true
				m.Ticks = 3
			}
		}

	case tickMsg:
		if m.Loaded {
			if m.Ticks == 0 {
				return m, tea.Quit
			}
			m.Ticks -= 1
		}
	}

	return m, nil
}

// VIEWS

func view(model tea.Model) string {
	m, _ := model.(Model)
	if !m.Chosen {
		return choicesView(m)
	}
	return chosenView(m)
}

const choicesTpl = `What to do today?

%s

Program quits in %d seconds.

(press j/k or up/down to select, enter to choose, and q or esc to quit)`

func choicesView(m Model) string {
	c := m.Choice

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		checkbox("Plant carrots", c == 0),
		checkbox("Go to the market", c == 1),
		checkbox("Read something", c == 2),
		checkbox("See friends", c == 3),
	)

	return fmt.Sprintf(choicesTpl, choices, m.Ticks)
}

func chosenView(m Model) string {
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
