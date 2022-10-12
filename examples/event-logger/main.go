package main

// An event logger that displays tea window-size, key, and mouse events.

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

var (
	typeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	msgStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD"))
	timeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF06B7"))
	tooltipStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
)

func main() {
	p := tea.NewProgram(&model{lastEvent: time.Now()},
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type event struct {
	timestamp time.Time
	msg       tea.Msg
}

type model struct {
	width  int
	height int

	lastEvent      time.Time
	lastMouseEvent tea.MouseEvent

	events []event
}

func (m *model) Init() tea.Cmd {
	return tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.MouseMsg:
		m.lastMouseEvent = tea.MouseEvent(msg)

		if m.lastMouseEvent.Type == tea.MouseMotion {
			// don't log motion events as it becomes spamy
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tick
	}

	m.lastEvent = time.Now()
	m.events = append(m.events, event{
		timestamp: time.Now(),
		msg:       msg,
	})

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	h := m.height - 3
	evs := m.events
	if m.height > 0 && len(evs) > h {
		evs = evs[len(evs)-h:]
	}

	var s string
	for _, event := range evs {
		s += fmt.Sprintf("[%s] %s\n",
			timeStyle.Render(event.timestamp.Format("15:04:05")),
			msgView(event.msg),
		)
	}

	lastEvent := fmt.Sprintf("Waiting for event... (last %s)", humanize.Time(m.lastEvent))
	cursor := fmt.Sprintf("X: %d, Y: %d", m.lastMouseEvent.X, m.lastMouseEvent.Y)

	return fmt.Sprintf("%s\n%s%s\n%s",
		lipgloss.NewStyle().Height(h+1).Render(s),
		lipgloss.NewStyle().Width(m.width-16).Render(lastEvent),
		lipgloss.NewStyle().Width(16).Align(lipgloss.Right).Render(cursor),
		tooltipStyle.Render("Press Ctrl+C to quit"),
	)
}

func msgView(msg tea.Msg) string {
	switch ev := msg.(type) {
	case tea.KeyMsg:
		return fmt.Sprintf("%s %s",
			typeStyle.Render("KeyMsg"),
			msgStyle.Render(
				fmt.Sprintf("%s (alt: %t)",
					ev.String(),
					ev.Alt,
				),
			),
		)
	case tea.MouseMsg:
		return fmt.Sprintf("%s %s",
			typeStyle.Render("MouseMsg"),
			msgStyle.Render(
				fmt.Sprintf("X: %d, Y: %d, type: %s",
					ev.X,
					ev.Y,
					tea.MouseEvent(ev),
				),
			),
		)
	case tea.WindowSizeMsg:
		return fmt.Sprintf("%s %s",
			typeStyle.Render("WindowSizeMsg"),
			msgStyle.Render(
				fmt.Sprintf("%d x %d",
					ev.Width,
					ev.Height,
				),
			),
		)

	default:
		return "Unknown event"
	}
}

type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
