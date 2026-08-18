package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	tea "charm.land/bubbletea/v2"
)

var shimmerColor color.Color

type model struct {
	text     string
	progress float64
}

var _ tea.Model = model{}

const tickInterval = 80 * time.Millisecond

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return tea.NewShimmer(m.text, tickInterval)
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.ShimmerMsg:
		m.progress = msg.Progress
		return m, tea.NewShimmer(m.text, tickInterval)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m model) View() tea.View {
	v := tea.NewView("")
	v.SetContent(fmt.Sprintf(
		"\n  %s\n\n  Press q to quit\n\n",
		tea.ShimmerText(m.text, m.progress, shimmerColor),
	))
	return v
}

func main() {
	shimmerColor = color.RGBA{R: 215, G: 119, B: 87, A: 255}

	p := tea.NewProgram(model{
		text:     "Loading data... please wait.",
		progress: 0.0,
	})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}