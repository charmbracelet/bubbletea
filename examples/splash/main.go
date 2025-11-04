package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// This example was ported from the awesome Textualize project by @willmcgugan.
// Check it out here:
// https://github.com/Textualize/textual/blob/main/examples/splash.py

// Color gradient
var colors = []color.Color{
	lipgloss.Color("#881177"),
	lipgloss.Color("#aa3355"),
	lipgloss.Color("#cc6666"),
	lipgloss.Color("#ee9944"),
	lipgloss.Color("#eedd00"),
	lipgloss.Color("#99dd55"),
	lipgloss.Color("#44dd88"),
	lipgloss.Color("#22ccbb"),
	lipgloss.Color("#00bbcc"),
	lipgloss.Color("#0099cc"),
	lipgloss.Color("#3366bb"),
	lipgloss.Color("#663399"),
}

type model struct {
	width  int
	height int
	rate   int64
}

func (m model) Init() tea.Cmd {
	return tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		return m, tick
	}
	return m, nil
}

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	if m.width == 0 {
		v.SetContent("Initializing...")
		return v
	}

	v.SetContent(m.gradient())
	return v
}

func (m model) gradient() string {
	// Time-based angle for animation
	t := float64(time.Now().UnixNano()*m.rate) / float64(time.Second)
	angleRadians := -t * math.Pi / 180.0
	sinAngle := math.Sin(angleRadians)
	cosAngle := math.Cos(angleRadians)

	centerX := float64(m.width) / 2
	centerY := float64(m.height)

	var output strings.Builder

	for lineY := range m.height {
		pointY := float64(lineY)*2 - centerY
		pointX := 0.0 - centerX

		x1 := (centerX + (pointX*cosAngle - pointY*sinAngle)) / float64(m.width)
		x2 := (centerX + (pointX*cosAngle - (pointY+1.0)*sinAngle)) / float64(m.width)
		pointX = float64(m.width) - centerX
		endX1 := (centerX + (pointX*cosAngle - pointY*sinAngle)) / float64(m.width)
		deltaX := (endX1 - x1) / float64(m.width)

		if math.Abs(deltaX) < 0.0001 {
			// Special case for verticals
			color1 := getGradientColor(x1)
			color2 := getGradientColor(x2)
			style := lipgloss.NewStyle().
				Foreground(color1).
				Background(color2)
			output.WriteString(style.Render(strings.Repeat("▀", m.width)))
		} else {
			// Render each column in the row
			for x := range m.width {
				pos1 := x1 + float64(x)*deltaX
				pos2 := x2 + float64(x)*deltaX
				color1 := getGradientColor(pos1)
				color2 := getGradientColor(pos2)
				style := lipgloss.NewStyle().
					Foreground(color1).
					Background(color2)
				output.WriteString(style.Render("▀"))
			}
		}
		if lineY < m.height-1 {
			output.WriteString("\n")
		}
	}

	return output.String()
}

func getGradientColor(position float64) color.Color {
	// Normalize position to [0,1]
	if position <= 0 {
		position = 0
	}
	if position >= 1 {
		position = 1
	}

	// Calculate the color index
	idx := position * float64(len(colors)-1)
	i1 := int(math.Floor(idx))
	i2 := int(math.Ceil(idx))

	// Ensure indices are within bounds
	i1 = i1 % len(colors)
	i2 = i2 % len(colors)
	if i1 < 0 {
		i1 += len(colors)
	}
	if i2 < 0 {
		i2 += len(colors)
	}

	// Interpolate between colors
	t := idx - float64(i1)
	return interpolateColors(colors[i1], colors[i2], t)
}

func interpolateColors(color1, color2 color.Color, t float64) color.Color {
	// Parse hex colors
	r1, g1, b1, _ := color1.RGBA()
	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2, _ := color2.RGBA()
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	// Interpolate
	r := int(float64(r1)*(1-t) + float64(r2)*t)
	g := int(float64(g1)*(1-t) + float64(g2)*t)
	b := int(float64(b1)*(1-t) + float64(b2)*t)

	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

type tickMsg time.Time

func tick() tea.Msg {
	return tickMsg(time.Now())
}

func main() {
	p := tea.NewProgram(
		model{rate: 90},
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
