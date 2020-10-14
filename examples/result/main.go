package main

// A simple example that shows how to retrieve a value from a Bubble Tea
// program after the Bubble Tea has exited.
//
// Thanks to Treilik for this one.

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	choices = []string{"Taro", "Coffee", "Lychee"}
)

type model struct {
	cursor int
	choice chan string
}

func main() {
	// This is where we'll listen for the choice the user makes in the Bubble
	// Tea program.
	result := make(chan string, 1)

	// Pass the channel to the initialize function so our Bubble Tea program
	// can send the final choice along when the time comes.
	if err := tea.NewProgram(initialize(result), update, view).Start(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}

	// Print out the final choice.
	if r := <-result; r != "" {
		fmt.Printf("\n---\nYou chose %s!\n", r)
	}
}

// Pass a channel to the model to listen to the result value. This is a
// function that returns the initialize function and is typically how you would
// pass arguments to a tea.Init function.
func initialize(choice chan string) func() (tea.Model, tea.Cmd) {
	return func() (tea.Model, tea.Cmd) {
		return model{cursor: 0, choice: choice}, nil
	}
}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			close(m.choice) // If we're quitting just chose the channel.
			return m, tea.Quit

		case "enter":
			// Send the choice on the channel and exit.
			m.choice <- choices[m.cursor]
			return m, tea.Quit

		case "down", "j":
			m.cursor++
			if m.cursor >= len(choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(choices) - 1
			}
		}

	}

	return m, nil
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)

	s := strings.Builder{}
	s.WriteString("What kind of Bubble Tea would you like to order?\n\n")

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}
	s.WriteString("\n(press q to quit)\n")

	return s.String()
}
