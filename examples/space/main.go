package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

// An example to show the FPS count of a moving space-like background.
//
// This was ported from the talented Orhun Parmaksız (@orhun)'s space example
// from his blog post "Why stdout is faster than stderr?".

type fps struct {
	frameCount  int
	lastInstant time.Time
	fps         *float64
}

func (f *fps) tick() {
	f.frameCount++
	elapsed := time.Since(f.lastInstant)
	// Update FPS every second if we have at least 2 frames
	if elapsed > time.Second && f.frameCount > 2 {
		fps := float64(f.frameCount) / elapsed.Seconds()
		f.fps = &fps
		f.frameCount = 0
		f.lastInstant = time.Now()
	}
}

type model struct {
	colors     [][]color.Color
	lastWidth  int
	lastHeight int
	fps        fps
	frameCount int
	width      int
	height     int
}

func initialModel() model {
	return model{
		fps: fps{
			lastInstant: time.Now(),
		},
	}
}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width != m.lastWidth || m.height != m.lastHeight {
			m.setupColors()
			m.lastWidth = m.width
			m.lastHeight = m.height
		}

	case tickMsg:
		m.frameCount++
		m.fps.tick()
		return m, tickCmd()
	}

	return m, nil
}

func (m *model) setupColors() {
	height := m.height * 2 // double height for half blocks
	m.colors = make([][]color.Color, height)

	for y := 0; y < height; y++ {
		m.colors[y] = make([]color.Color, m.width)
		randomnessFactor := float64(height-y) / float64(height)

		for x := 0; x < m.width; x++ {
			baseValue := randomnessFactor * (float64(height-y) / float64(height))
			randomOffset := (rand.Float64() * 0.2) - 0.1
			value := clamp(baseValue+randomOffset, 0, 1)

			// Convert value to grayscale color (0-255)
			gray := uint8(value * 255)
			m.colors[y][x] = lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", gray, gray, gray))
		}
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (m model) View() string {
	// Title and FPS display
	title := lipgloss.NewStyle().Bold(true).Render("Space")
	fpsText := ""
	if m.fps.fps != nil {
		fpsText = fmt.Sprintf("%.1f fps", *m.fps.fps)
	}
	header := fmt.Sprintf("%s %s", title, fpsText)

	// Color display
	var s strings.Builder
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			xi := (x + m.frameCount) % m.width
			fg := m.colors[y*2][xi]
			bg := m.colors[y*2+1][xi]
			st := lipgloss.NewStyle().Foreground(fg).Background(bg)
			s.WriteString(st.Render("▀"))
		}
		s.WriteByte('\n')
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, s.String())
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithFPS(120))

	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
