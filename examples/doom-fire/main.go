package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// This Doom Fire implementation was ported from @const-void's Node version.
// See https://github.com/const-void/DOOM-fire-node

var whiteFg = lipgloss.NewStyle().Foreground(lipgloss.White)

type model struct {
	screenBuf   []int
	width       int
	height      int
	firePalette []int
	startTime   time.Time
}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tickMsg:
		m.spreadFire()
		return m, tick
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height * 2 // Double height for half-block characters
		m.screenBuf = make([]int, m.width*m.height)
		// Initialize the bottom row with white (maximum intensity)
		for i := 0; i < m.width; i++ {
			m.screenBuf[(m.height-1)*m.width+i] = len(m.firePalette) - 1
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var s strings.Builder

	for y := 0; y < m.height-2; y += 2 {
		for x := 0; x < m.width; x++ {
			pixelHi := m.screenBuf[y*m.width+x]
			pixelLo := m.screenBuf[(y+1)*m.width+x]

			// Convert palette indices to xterm-256 colors
			hiColor := m.firePalette[pixelHi]
			loColor := m.firePalette[pixelLo]

			s.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.ANSIColor(hiColor)).
				Background(lipgloss.ANSIColor(loColor)).
				Render("▀"))
		}
		if y < m.height-2 {
			s.WriteByte('\n')
		}
	}

	elapsed := time.Since(m.startTime)
	s.WriteString(whiteFg.Render("Press q or ctrl+c to quit. " + fmt.Sprintf("Elapsed: %s", elapsed.Round(time.Second))))
	return s.String()
}

func (m *model) spreadFire() {
	for x := 0; x < m.width; x++ {
		for y := 0; y < m.height; y++ {
			m.spreadPixel(y*m.width + x)
		}
	}
}

func (m *model) spreadPixel(idx int) {
	if idx < m.width {
		return
	}

	pixel := m.screenBuf[idx]
	if pixel == 0 {
		m.screenBuf[idx-m.width] = 0
		return
	}

	rnd := rand.Intn(3)
	dst := idx - rnd + 1
	if dst-m.width >= 0 && dst-m.width < len(m.screenBuf) {
		decay := rnd & 1
		newValue := pixel - decay
		if newValue < 0 {
			newValue = 0
		}
		m.screenBuf[dst-m.width] = newValue
	}
}

type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Millisecond * 50)
	return tickMsg(time.Now())
}

func initialModel() model {
	// Same color palette as the original
	palette := []int{0, 233, 234, 52, 53, 88, 89, 94, 95, 96, 130, 131, 132, 133, 172, 214, 215, 220, 220, 221, 3, 226, 227, 230, 231, 7}

	return model{
		firePalette: palette,
		startTime:   time.Now(),
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
