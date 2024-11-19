package main

// A simple example that shows how to send activity to Bubble Tea in real-time
// through a channel.

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// A message used to indicate that activity has occurred. As an example, it
// will contain sample chat messages.
type responseMsg struct {
	ChatMessage string
}

// Simulate a process that sends events at an irregular interval in real time.
// In this case, we'll send events on the channel at a random interval between
// 100 to 1000 milliseconds. As a command, Bubble Tea will run this
// asynchronously.
func listenForActivity(sub chan responseMsg) tea.Cmd {
	return func() tea.Msg {
		// Create some sample chat messages that will be sent to demonstrate sending messages
		// in real-time.
		sampleChatMessages := []string{
			"Hey! How's it going?",
			"Did you catch the game last night?",
			"I'm thinking of grabbing lunch. Want to join?",
			"Have you seen the new movie that just came out?",
			"What are your plans for the weekend?",
			"I just finished reading a great book. You should check it out!",
			"Can you send me the notes from yesterday's meeting?",
			"I love this weather! Perfect for a walk.",
			"Let's go for coffee sometime this week.",
			"Did you hear about the latest update to the app?",
		}
		for {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(900)+100)) // nolint:gosec
			// Send a random chat message
			sub <- responseMsg{ChatMessage: sampleChatMessages[rand.Intn(len(sampleChatMessages))]} // nolint:gosec
		}
	}
}

// A command that waits for the activity on a channel.
func waitForActivity(sub chan responseMsg) tea.Cmd {
	return func() tea.Msg {
		return responseMsg(<-sub)
	}
}

type model struct {
	sub           chan responseMsg // where we'll receive activity notifications
	responseCount int              // how many responses we've received
	response      string           // the last response we received
	spinner       spinner.Model
	quitting      bool
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
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case responseMsg:
		m.responseCount++
		m.response = msg.ChatMessage
		return m, waitForActivity(m.sub) // wait for next event
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() string {
	s := fmt.Sprintf("\n %s Events received: %d\nLast Message: %s\n Press any key to exit\n", m.spinner.View(), m.responseCount, m.response)
	if m.quitting {
		s += "\n"
	}
	return s
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
