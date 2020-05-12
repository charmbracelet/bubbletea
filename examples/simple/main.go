package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/boba"
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
	p := boba.NewProgram(initialize, update, view)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func initialize() (boba.Model, boba.Cmd) {
	return model(5), tick
}

// Update is called when messages are recived. The idea is that you inspect
// the message and update the model (or send back a new one) accordingly. You
// can also return a commmand, which is a function that peforms I/O and
// returns a message.
func update(msg boba.Msg, mdl boba.Model) (boba.Model, boba.Cmd) {
	m, _ := mdl.(model)

	switch msg.(type) {
	case boba.KeyMsg:
		return m, boba.Quit
	case tickMsg:
		m -= 1
		if m <= 0 {
			return m, boba.Quit
		}
		return m, tick
	}
	return m, nil
}

// Views take data from the model and return a string which will be rendered
// to the terminal.
func view(mdl boba.Model) string {
	m, _ := mdl.(model)
	return fmt.Sprintf("Hi. This program will exit in %d seconds. To quit sooner press any key.\n", m)
}

func tick() boba.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
