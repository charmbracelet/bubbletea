package main

// A simple example that shows how to send messages to a Bubble Tea program
// from outside the program using Program.Send(Msg).

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	dotStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	durationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type resultMsg struct {
	food     string
	duration time.Duration
}

func (r resultMsg) String() string {
	if r.duration == 0 {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf("🍔 Ate %s %s", r.food,
		durationStyle.Render(r.duration.String()))
}

type model struct {
	results  []resultMsg
	spinner  spinner.Model
	styles   *styles
	quitting bool
}

type styles struct {
	spinnerStyle lipgloss.Style
	helpStyle    lipgloss.Style
	appStyle     lipgloss.Style
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	m.styles = &styles{
		spinnerStyle: ctx.NewStyle().Foreground(lipgloss.Color("63")),
		helpStyle:    ctx.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0),
		appStyle:     ctx.NewStyle().Margin(1, 2, 0, 2),
	}

	const numLastResults = 5
	m.spinner = spinner.New(ctx)
	m.spinner.Style = m.styles.spinnerStyle
	m.results = make([]resultMsg, numLastResults)

	return m, m.spinner.Tick
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case resultMsg:
		m.results = append(m.results[1:], msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(ctx, msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View(ctx tea.Context) string {
	var s string

	if m.quitting {
		s += "That’s all for today!"
	} else {
		s += m.spinner.View(ctx) + " Eating food..."
	}

	s += "\n\n"

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	if !m.quitting {
		s += m.styles.helpStyle.Render("Press any key to exit")
	}

	if m.quitting {
		s += "\n"
	}

	return m.styles.appStyle.Render(s)
}

func main() {
	p := tea.NewProgram(model{})

	// Simulate activity
	go func() {
		for {
			pause := time.Duration(rand.Int63n(899)+100) * time.Millisecond // nolint:gosec
			time.Sleep(pause)

			// Send the Bubble Tea program a message from outside the
			// tea.Program. This will block until it is ready to receive
			// messages.
			p.Send(resultMsg{food: randomFood(), duration: pause})
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func randomFood() string {
	food := []string{
		"an apple", "a pear", "a gherkin", "a party gherkin",
		"a kohlrabi", "some spaghetti", "tacos", "a currywurst", "some curry",
		"a sandwich", "some peanut butter", "some cashews", "some ramen",
	}
	return food[rand.Intn(len(food))] // nolint:gosec
}
