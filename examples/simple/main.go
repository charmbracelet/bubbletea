package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	// Log to a file. Useful in debugging since you can't really log to stdout.
	// Not required.
	logfilePath := os.Getenv("BUBBLETEA_LOG")
	if logfilePath != "" {
		if _, err := tea.LogToFile(logfilePath, "simple"); err != nil {
			log.Fatal(err)
		}
	}

	// Declare our program.
	p := tea.Program[model]{
		// Init optionally returns an initial command we should run. In this
		// case we want to start the timer.
		Init: func() (model, tea.Cmd) {
			return model(5), tea.Batch(tick)
		},

		// Update is called when messages are received. The idea is that you
		// inspect the message and send back an updated model accordingly. You
		// can also return a command, which is a function that performs I/O and
		// returns a message.
		Update: func(m model, msg tea.Msg) (model, tea.Cmd) {
			switch msg := msg.(type) {
			case tea.KeyPressMsg:
				switch msg.String() {
				case "ctrl+c", "q":
					return m, tea.Quit
				case "ctrl+z":
					return m, tea.Suspend
				}

			case tickMsg:
				m--
				if m <= 0 {
					return m, tea.Quit
				}
				return m, tick
			}
			return m, nil
		},

		// View returns a string based on data in the model. That string which
		// will be rendered to the terminal.
		View: func(m model) fmt.Stringer {
			return tea.NewFrame(fmt.Sprintf("Hi. This program will exit in %d seconds.\n\nTo quit sooner press ctrl-c, or press ctrl-z to suspend...\n", m))
		},
	}

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// A model can be more or less any type of data. It holds all the data for a
// program, so often it's a struct. For this simple example, however, all
// we'll need is a simple integer.
type model int

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
