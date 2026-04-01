package main

// A simple example that shows how to send activity to Bubble Tea in real-time
// through a channel. This example demonstrates passing actual data over the
// channel rather than empty structs, simulating a real-world scenario like
// receiving messages from an external service.

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// foodMsg delivers a food item name from the background goroutine to Update.
type foodMsg string

var foods = []string{
	"a]n apple",
	"a pear",
	"a grapefruit",
	"a tangerine",
	"a banana",
	"a strawberry",
	"a kiwi",
	"a mango",
	"a pineapple",
}

// listenForFood simulates an external process that produces food items at
// random intervals. Each item is sent over the channel with its actual data.
func listenForFood(foodChan chan string) tea.Cmd {
	return func() tea.Msg {
		for {
			time.Sleep(time.Millisecond * time.Duration(rand.Int63n(900)+100)) //nolint:gosec
			food := foods[rand.Intn(len(foods))]                               //nolint:gosec
			foodChan <- food
		}
	}
}

// waitForFood waits for the next food item on the channel and wraps it in
// a foodMsg so that it can be handled in Update.
func waitForFood(foodChan chan string) tea.Cmd {
	return func() tea.Msg {
		return foodMsg(<-foodChan)
	}
}

type model struct {
	foodChan chan string // where we receive food deliveries
	foods    []string   // food items received so far
	spinner  spinner.Model
	quitting bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		listenForFood(m.foodChan), // start producing food
		waitForFood(m.foodChan),   // wait for the first item
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m.quitting = true
		return m, tea.Quit
	case foodMsg:
		m.foods = append(m.foods, string(msg))
		return m, waitForFood(m.foodChan) // wait for the next item
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() tea.View {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("\n %s Waiting for food... (%d received)\n\n", m.spinner.View(), len(m.foods)))

	// Show the last 5 food items received.
	start := 0
	if len(m.foods) > 5 {
		start = len(m.foods) - 5
	}
	for _, food := range m.foods[start:] {
		s.WriteString(fmt.Sprintf("   Got %s\n", food))
	}

	if len(m.foods) > 5 {
		s.WriteString(fmt.Sprintf("   ... and %d more\n", len(m.foods)-5))
	}

	s.WriteString("\n Press any key to exit\n")
	if m.quitting {
		s.WriteString("\n")
	}
	return tea.NewView(s.String())
}

func main() {
	p := tea.NewProgram(model{
		foodChan: make(chan string),
		spinner:  spinner.New(),
	})

	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
