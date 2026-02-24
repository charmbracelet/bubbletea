package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
)

type model struct {
	width    int
	flip     bool
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		default:
			m.flip = !m.flip
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var view tea.View
	if m.quitting {
		return view
	}

	z := []int{0, 1}
	if m.flip {
		z = reverse(z)
	}

	footer := lipgloss.NewStyle().
		Height(13).
		Foreground(charmtone.Oyster).
		AlignVertical(lipgloss.Bottom).
		Render("Press any key to swap the cards, or q to quit.")

	cardA := newCard("Hello").Z(z[0])
	cardB := newCard("Goodbye").Z(z[1])
	comp := lipgloss.NewCompositor(
		lipgloss.NewLayer(footer),
		cardA,
		cardB.X(10).Y(2),
	)
	view.SetContent(comp.Render())

	return view
}

func newCard(str string) *lipgloss.Layer {
	return lipgloss.NewLayer(
		lipgloss.NewStyle().
			Width(20).
			Height(10).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(charmtone.Charple).
			Align(lipgloss.Center, lipgloss.Center).
			Render(str),
	)
}

// Reverse a slice, returning a new slice.
func reverse[T any](s []T) []T {
	n := len(s)
	r := make([]T, n)
	for i, v := range s {
		r[n-1-i] = v
	}
	return r
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Urgh:", err)
		os.Exit(1)
	}
}
