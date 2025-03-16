// roughly converted to Go from https://github.com/dmtrKovalenko/esp32-smooth-eye-blinking/blob/main/src/main.cpp
package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Eye dimensions (corresponding to original EYE_WIDTH and EYE_HEIGHT)
	eyeWidth   = 15
	eyeHeight  = 12 // Increased height for taller eyes
	eyeSpacing = 40

	// Blink animation timing (matching original constants)
	blinkFrames = 20
	openTimeMin = 1000
	openTimeMax = 4000
)

// Characters for drawing the eyes
const (
	eyeChar = "●"
	bgChar  = " "
)

type model struct {
	width        int
	height       int
	eyePositions [2]int
	eyeY         int
	isBlinking   bool
	blinkState   int
	lastBlink    time.Time
	openTime     time.Duration
}

type tickMsg time.Time

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}

func initialModel() model {
	m := model{
		width:      80,
		height:     24,
		isBlinking: false,
		blinkState: 0,
		lastBlink:  time.Now(),
		openTime:   time.Duration(rand.Intn(openTimeMax-openTimeMin)+openTimeMin) * time.Millisecond,
	}

	m.updateEyePositions()
	return m
}

func (m *model) updateEyePositions() {
	startX := (m.width - eyeSpacing) / 2
	m.eyeY = m.height / 2

	m.eyePositions[0] = startX
	m.eyePositions[1] = startX + eyeSpacing
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		tea.EnterAltScreen,
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateEyePositions()

	case tickMsg:
		currentTime := time.Now()

		if !m.isBlinking && currentTime.Sub(m.lastBlink) >= m.openTime {
			m.isBlinking = true
			m.blinkState = 0
		}

		if m.isBlinking {
			m.blinkState++

			if m.blinkState >= blinkFrames {
				m.isBlinking = false
				m.lastBlink = currentTime
				m.openTime = time.Duration(rand.Intn(openTimeMax-openTimeMin)+openTimeMin) * time.Millisecond

				// 10% chance of double blink (matching original logic)
				if rand.Intn(10) == 0 {
					m.openTime = 300 * time.Millisecond
				}
			}
		}
	}

	return m, tickCmd()
}

func (m model) View() string {
	// Create empty canvas
	canvas := make([][]string, m.height)
	for y := range canvas {
		canvas[y] = make([]string, m.width)
		for x := range canvas[y] {
			canvas[y][x] = bgChar
		}
	}

	// Calculate current eye height based on blink state
	currentHeight := eyeHeight
	if m.isBlinking {
		var blinkProgress float64

		if m.blinkState < blinkFrames/2 {
			// Closing eyes (with easing function from original)
			blinkProgress = float64(m.blinkState) / float64(blinkFrames/2)
			blinkProgress = 1.0 - (blinkProgress * blinkProgress)
		} else {
			// Opening eyes (with easing function from original)
			blinkProgress = float64(m.blinkState-blinkFrames/2) / float64(blinkFrames/2)
			blinkProgress = blinkProgress * (2.0 - blinkProgress)
		}

		currentHeight = int(math.Max(1, float64(eyeHeight)*blinkProgress))
	}

	// Draw both eyes
	for i := 0; i < 2; i++ {
		drawEllipse(canvas, m.eyePositions[i], m.eyeY, eyeWidth, currentHeight)
	}

	// Convert canvas to string
	var s strings.Builder
	for _, row := range canvas {
		for _, cell := range row {
			s.WriteString(cell)
		}
		s.WriteString("\n")
	}

	// Style output
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F0F0F0"))

	return style.Render(s.String())
}

func drawEllipse(canvas [][]string, x0, y0, rx, ry int) {
	// Improved ellipse drawing algorithm with better angles
	for y := -ry; y <= ry; y++ {
		// Calculate the width at this y position for a smoother ellipse
		// Use a slightly modified formula to improve the angles
		width := int(float64(rx) * math.Sqrt(1.0-math.Pow(float64(y)/float64(ry), 2.0)))

		for x := -width; x <= width; x++ {
			// Calculate canvas position
			canvasX := x0 + x
			canvasY := y0 + y

			// Make sure we're within canvas bounds
			if canvasX >= 0 && canvasX < len(canvas[0]) && canvasY >= 0 && canvasY < len(canvas) {
				canvas[canvasY][canvasX] = eyeChar
			}
		}
	}
}
