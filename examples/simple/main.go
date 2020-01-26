package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/tea"
)

// A model can be more or less any type of data. It holds all the data for a
// program, so often it's a struct. For this simple example, however, all
// we'll need is a simple integer.
type Model int

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type TickMsg struct{}

func main() {
	// Initialize our program
	p := tea.NewProgram(initialize, update, view, subscriptions)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func initialize() (tea.Model, tea.Cmd) {
	return Model(5), nil
}

// Update is called when messages are recived. The idea is that you inspect
// the message and update the model (or send back a new one) accordingly. You
// can also return a commmand, which is a function that peforms I/O and
// returns a message.
func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, _ := model.(Model)

	switch msg.(type) {
	case TickMsg:
		m -= 1
		if m <= 0 {
			return m, tea.Quit
		}
	}
	return m, nil
}

// Views take data from the model and return a string which will be rendered
// to the terminal.
func view(model tea.Model) string {
	m, _ := model.(Model)
	return fmt.Sprintf("Hi. This program will exit in %d seconds...", m)
}

// This is a subscription which we setup in NewProgram(). It waits for one
// second, sends a tick, and then restarts.
func subscriptions(_ tea.Model) tea.Subs {
	return tea.Subs{
		"tick": func(_ tea.Model) tea.Msg {
			time.Sleep(time.Second)
			return TickMsg{}
		},
	}
}
