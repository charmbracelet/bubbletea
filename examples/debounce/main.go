package main

// This example illustrates how to debounce commands.
//
// When the user presses a key we increment the "tag" value on the model and,
// after a short delay, we include that tag value in the message produced
// by the Tick command.
//
// In a subsequent Update, if the tag in the Msg matches current tag on the
// model's state we know that the debouncing is complete and we can proceed as
// normal. If not, we simply ignore the inbound message.

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const debounceDuration = time.Second

type exitMsg int

type model struct {
	tag int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Increment the tag on the model...
		m.tag++
		return m, tea.Tick(debounceDuration, func(_ time.Time) tea.Msg {
			// ...and include a copy of that tag value in the message.
			return exitMsg(m.tag)
		})
	case exitMsg:
		// If the tag in the message doesn't match the tag on the model then we
		// know that this message was not the last one sent and another is on
		// the way. If that's the case we know, we can ignore this message.
		// Otherwise, the debounce timeout has passed and this message is a
		// valid debounced one.
		if int(msg) == m.tag {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Key presses: %d", m.tag) +
		"\nTo exit press any key, then wait for one second without pressing anything."
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("uh oh:", err)
		os.Exit(1)
	}
}
