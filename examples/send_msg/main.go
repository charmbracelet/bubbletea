package send_msg

// A simple example that shows how to send messages to a Bubble Tea program
// from outside the program using Program.Send(Msg).

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/rprtr258/bubbletea"
	"github.com/rprtr258/bubbletea/bubbles/spinner"
	"github.com/rprtr258/bubbletea/lipgloss"
)

var (
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle      = helpStyle.Copy().UnsetMargins()
	durationStyle = dotStyle.Copy()
	appStyle      = lipgloss.NewStyle().Margin(1, 2, 0, 2)
)

type resultMsg struct {
	duration time.Duration
	food     string
}

func (r resultMsg) String() string {
	if r.duration == 0 {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf("🍔 Ate %s %s", r.food,
		durationStyle.Render(r.duration.String()))
}

type model struct {
	spinner  spinner.Model
	results  []resultMsg
	quitting bool
}

func newModel() model {
	const numLastResults = 5
	s := spinner.New()
	s.Style = spinnerStyle
	return model{
		spinner: s,
		results: make([]resultMsg, numLastResults),
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MsgKey:
		m.quitting = true
		return m, tea.Quit
	case resultMsg:
		m.results = append(m.results[1:], msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View(r tea.Renderer) {
	var s string

	if m.quitting {
		s += "That’s all for today!"
	} else {
		s += m.spinner.View() + " Eating food..."
	}

	s += "\n\n"

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	if !m.quitting {
		s += helpStyle.Render("Press any key to exit")
	}

	if m.quitting {
		s += "\n"
	}

	r.Write(appStyle.Render(s))
}

func Main() {
	p := tea.NewProgram(newModel())

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
