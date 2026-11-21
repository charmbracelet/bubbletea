package main

// A simple example that shows how to send activity to Bubble Tea in real-time
// through a channel.

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// A message carrying the data that arrived on the channel. In the real world
// (for example, chat) this would hold the incoming chat message.
type responseMsg struct {
	sentAt time.Time
}

// Simulate a process that sends events at an irregular interval in real time.
// In this case, we'll send events on the channel at a random interval between
// 100 to 1000 milliseconds. As a command, Bubble Tea will run this
// asynchronously.
func listenForActivity(sub chan responseMsg) tea.Cmd {
	return func() tea.Msg {
		for {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(900)+100)) // nolint:gosec
			sub <- responseMsg{sentAt: time.Now()}
		}
	}
}

// A command that waits for the activity on a channel.
func waitForActivity(sub chan responseMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

type model struct {
	sub       chan responseMsg // where we'll receive activity
	responses int              // how many responses we've received
	spinner   spinner.Model
	quitting  bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		listenForActivity(m.sub), // generate activity
		waitForActivity(m.sub),   // wait for activity
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m.quitting = true
		return m, tea.Quit
	case responseMsg:
		m.responses++ // record external activity
		return m, tea.Batch(
			tea.Printf("Event %d sent at %s", m.responses, msg.sentAt.Format("15:04:05.000")),
			waitForActivity(m.sub), // wait for next event
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() tea.View {
	s := fmt.Sprintf("\n %s Events received: %d\n\n Press any key to exit\n", m.spinner.View(), m.responses)
	if m.quitting {
		s += "\n"
	}
	return tea.NewView(s)
}

func main() {
	p := tea.NewProgram(model{
		sub:     make(chan responseMsg),
		spinner: spinner.New(),
	})

	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
