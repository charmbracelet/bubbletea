package main

// This example demonstrates two-way communication between a Bubble Tea
// program and a background goroutine. The TUI sends requests to a worker
// goroutine and receives results back, showing how to build interactive
// programs that delegate work to background processes.

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// request is sent from the TUI to the background worker.
type request struct {
	query string
}

// resultMsg is sent from the background worker back to the TUI via a Cmd.
type resultMsg struct {
	query    string
	answer   string
	duration time.Duration
}

// worker runs in a background goroutine. It reads requests from the requests
// channel, simulates processing, and sends results back on the results channel.
func worker(requests <-chan request, results chan<- resultMsg) {
	for req := range requests {
		// Simulate some processing time.
		d := time.Duration(rand.Int63n(500)+200) * time.Millisecond //nolint:gosec
		time.Sleep(d)

		answers := []string{
			"Yes, absolutely!",
			"No way.",
			"Maybe, ask again later.",
			"It is certain.",
			"Very doubtful.",
			"Signs point to yes.",
			"Better not tell you now.",
		}
		answer := answers[rand.Intn(len(answers))] //nolint:gosec

		results <- resultMsg{
			query:    req.query,
			answer:   answer,
			duration: d,
		}
	}
}

// waitForResult returns a Cmd that waits for the next result from the worker.
func waitForResult(results <-chan resultMsg) tea.Cmd {
	return func() tea.Msg {
		return <-results
	}
}

type model struct {
	requests chan<- request
	results  <-chan resultMsg
	input    string
	history  []resultMsg
	waiting  bool
	quitting bool
}

func newModel() model {
	requests := make(chan request)
	results := make(chan resultMsg)

	// Start the background worker.
	go worker(requests, results)

	return model{
		requests: requests,
		results:  results,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			close(m.requests)
			return m, tea.Quit

		case "enter":
			if m.waiting || len(m.input) == 0 {
				return m, nil
			}
			// Send the question to the worker.
			m.requests <- request{query: m.input}
			m.waiting = true
			m.input = ""
			return m, waitForResult(m.results)

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil

		default:
			if msg.Text != "" {
				m.input += msg.Text
			}
			return m, nil
		}

	case resultMsg:
		m.history = append(m.history, msg)
		m.waiting = false
		return m, nil

	default:
		return m, nil
	}
}

func (m model) View() tea.View {
	var b strings.Builder

	b.WriteString("  Two-Way Goroutine Communication\n")
	b.WriteString("  Ask the oracle a question!\n\n")

	// Show history (last 5 entries).
	start := 0
	if len(m.history) > 5 {
		start = len(m.history) - 5
	}
	for _, r := range m.history[start:] {
		b.WriteString(fmt.Sprintf("  Q: %s\n", r.query))
		b.WriteString(fmt.Sprintf("  A: %s (%s)\n\n", r.answer, r.duration))
	}

	if m.waiting {
		b.WriteString("  Thinking...\n\n")
	} else {
		b.WriteString(fmt.Sprintf("  > %s_\n\n", m.input))
	}

	b.WriteString("  enter: ask  esc: quit\n")

	if m.quitting {
		b.WriteString("\n  Goodbye!\n")
	}

	return tea.NewView(b.String())
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
