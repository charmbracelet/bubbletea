package main

// TODO: The views feel messy. Clean 'em up.

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/boba"
	"github.com/fogleman/ease"
)

func main() {
	p := boba.NewProgram(
		initialize,
		update,
		view,
	)
	if err := p.Start(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

// MSG

type tickMsg struct{}

type frameMsg struct{}

// MODEL

// Model contains the data for our application.
type Model struct {
	Choice   int
	Chosen   bool
	Ticks    int
	Frames   int
	Progress float64
	Loaded   bool
}

// INIT

func initialize() (boba.Model, boba.Cmd) {
	return Model{0, false, 10, 0, 0, false}, tick
}

// CMDS

func tick() boba.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}

func frame() boba.Msg {
	time.Sleep(time.Second / 60)
	return frameMsg{}
}

// UPDATE

func update(msg boba.Msg, model boba.Model) (boba.Model, boba.Cmd) {
	m, _ := model.(Model)

	if !m.Chosen {
		return updateChoices(msg, m)
	}
	return updateChosen(msg, m)
}

func updateChoices(msg boba.Msg, m Model) (boba.Model, boba.Cmd) {
	switch msg := msg.(type) {

	case boba.KeyMsg:
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
			return m, boba.Quit
		}

	case tickMsg:
		if m.Ticks == 0 {
			return m, boba.Quit
		}
		m.Ticks -= 1
	}

	return m, tick
}

func updateChosen(msg boba.Msg, m Model) (boba.Model, boba.Cmd) {
	switch msg := msg.(type) {

	case boba.KeyMsg:
		switch msg.String() {
		case "q":
			fallthrough
		case "esc":
			fallthrough
		case "ctrl+c":
			return m, boba.Quit
		}

	case frameMsg:
		if !m.Loaded {
			m.Frames += 1
			m.Progress = ease.OutBounce(float64(m.Frames) / float64(160))
			if m.Progress >= 1 {
				m.Progress = 1
				m.Loaded = true
				m.Ticks = 3
			}
		}

	case tickMsg:
		if m.Loaded {
			if m.Ticks == 0 {
				return m, boba.Quit
			}
			m.Ticks -= 1
		}
	}

	return m, frame
}

// VIEW

func view(model boba.Model) string {
	m, _ := model.(Model)
	if !m.Chosen {
		return choicesView(m) + "\n"
	}
	return chosenView(m) + "\n"
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
