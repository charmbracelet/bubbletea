package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbles"
)

// A model can be more or less any type of data. It holds all the data for a
// program, so often it's a struct. For this simple example, however, all
// we'll need is a simple integer.
type model int

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type tickMsg time.Time

func main() {
	// Initialize our program
	p := tea.NewProgram(initialize, update, view)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func initialize() (tea.Model, tea.Cmd) {
	return model(5), tick
}

// Update is called when messages are recived. The idea is that you inspect
// the message and update the model (or send back a new one) accordingly. You
// can also return a commmand, which is a function that peforms I/O and
// returns a message.
func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tickMsg:
		m -= 1
		if m <= 0 {
			return m, tea.Quit
		}
		return m, tick
	}
	return m, nil
}

// Views take data from the model and return a string which will be rendered
// to the terminal.
func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	return fmt.Sprintf("Hi. This program will exit in %d seconds. To quit sooner press any key.\n", m)
}

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
