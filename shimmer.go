package tea

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// ShimmerMsg is sent by [NewShimmer] to indicate a shimmer animation tick.
// It carries the current progress of the shimmer sweep from 0.0 (start)
// to 1.0 (end).
type ShimmerMsg struct {
	// Progress is the current shimmer position, between 0.0 and 1.0.
	Progress float64
}

// NewShimmer returns a [Cmd] that fires a [ShimmerMsg] after the given
// interval. To create a continuous shimmer animation, return NewShimmer
// from your Update handler when you receive a ShimmerMsg:
//
//	func (m model) Init() tea.Cmd {
//	    return tea.NewShimmer("Thinking...", 80*time.Millisecond)
//	}
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case tea.ShimmerMsg:
//	        m.progress = msg.Progress
//	        return m, tea.NewShimmer(m.text, 80*time.Millisecond)
//	    }
//	    return m, nil
//	}
//
//	func (m model) View() tea.View {
//	    s := tea.ShimmerText(
//	        m.text,
//	        m.progress,
//	        color.RGBA{R: 215, G: 119, B: 87, A: 255},
//	    )
//	    return tea.NewView(s)
//	}
func NewShimmer(text string, interval time.Duration) Cmd {
	return func() Msg {
		time.Sleep(interval)
		return ShimmerMsg{Progress: 0.03}
	}
}

// ShimmerText renders text with a shimmer/glow effect. A narrow highlight
// window sweeps across the text from left to right, similar to the shimmer
// effect seen in Claude Code's terminal UI.
//
// progress must be between 0.0 (shimmer at start) and 1.0 (shimmer at end).
// The shimmer color is applied to the portion of text within the sweep zone.
//
// width controls the width of the shimmer zone in character cells. Defaults
// to 3.0 when 0 is passed.
//
// Returns a styled string with ANSI escape codes for the highlighted portion.
func ShimmerText(text string, progress float64, shimmerColor color.Color, opts ...float64) string {
	if text == "" {
		return ""
	}

	width := 3.0
	if len(opts) > 0 {
		width = opts[0]
	}

	textWidth := ansi.StringWidth(text)
	if textWidth <= 0 {
		return text
	}

	if progress < 0 {
		progress = 0
	} else if progress > 1 {
		progress = 1
	}

	center := progress * float64(textWidth)

	colorSeq := colorToAnsi(shimmerColor)
	resetSeq := "\x1b[0m"
	if colorSeq == "" {
		return text
	}

	var sb strings.Builder
	col := 0
	inShimmer := false

	for _, r := range []rune(text) {
		rw := ansi.StringWidth(string(r))
		if rw <= 0 {
			rw = 1
		}

		centerPos := float64(col) + float64(rw)/2.0
		if math.Abs(centerPos-center) < width {
			if !inShimmer {
				sb.WriteString(colorSeq)
				inShimmer = true
			}
			sb.WriteRune(r)
		} else {
			if inShimmer {
				sb.WriteString(resetSeq)
				inShimmer = false
			}
			sb.WriteRune(r)
		}
		col += rw
	}

	if inShimmer {
		sb.WriteString(resetSeq)
	}
	return sb.String()
}

// ShimmerState is a convenience type that encapsulates shimmer animation
// state and eliminates boilerplate.
//
// Example:
//
//	type model struct {
//	    shimmer tea.ShimmerState
//	}
//
//	func (m model) Init() tea.Cmd {
//	    m.shimmer = tea.NewShimmerState("Thinking...", 80*time.Millisecond, myColor)
//	    return tea.NewShimmer("Thinking...", 80*time.Millisecond)
//	}
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case tea.ShimmerMsg:
//	        m.shimmer.Tick(msg.Progress)
//	        return m, m.shimmer.NextCmd()
//	    }
//	    return m, nil
//	}
//
//	func (m model) View() tea.View {
//	    return tea.NewView(m.shimmer.Render())
//	}
type ShimmerState struct {
	text     string
	interval time.Duration
	color    color.Color
	progress float64
	width    float64
}

// NewShimmerState creates a new [ShimmerState] with the given text, interval,
// and shimmer color. The initial progress is 0.0.
func NewShimmerState(text string, interval time.Duration, color color.Color, opts ...float64) ShimmerState {
	s := ShimmerState{
		text:     text,
		interval: interval,
		color:    color,
		progress: 0.0,
	}
	if len(opts) > 0 {
		s.width = opts[0]
	}
	return s
}

// Tick advances the shimmer progress. It should be called when a [ShimmerMsg]
// is received.
func (s *ShimmerState) Tick(progress float64) {
	s.progress += progress
	for s.progress >= 1.0 {
		s.progress -= 1.0
	}
}

// NextCmd returns a [Cmd] that will fire the next shimmer tick. Return this
// from your Update handler to continue the animation.
func (s *ShimmerState) NextCmd() Cmd {
	return NewShimmer(s.text, s.interval)
}

// Render returns the text with the current shimmer effect applied.
func (s ShimmerState) Render() string {
	if s.width > 0 {
		return ShimmerText(s.text, s.progress, s.color, s.width)
	}
	return ShimmerText(s.text, s.progress, s.color)
}

// Text returns the underlying text being shimmered.
func (s ShimmerState) Text() string {
	return s.text
}

// SetText updates the text being shimmered.
func (s *ShimmerState) SetText(text string) {
	s.text = text
}

// Reset resets the shimmer progress to 0.0.
func (s *ShimmerState) Reset() {
	s.progress = 0.0
}

// colorToAnsi converts an image/color.Color to an ANSI foreground escape
// sequence using the ECMA-48 24-bit color format: \x1b[38;2;R;G;Bm
func colorToAnsi(c color.Color) string {
	if c == nil {
		return ""
	}

	switch col := c.(type) {
	case color.RGBA:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", col.R, col.G, col.B)
	case color.NRGBA:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", col.R, col.G, col.B)
	case color.RGBA64:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", uint8(col.R>>8), uint8(col.G>>8), uint8(col.B>>8))
	case color.NRGBA64:
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", uint8(col.R>>8), uint8(col.G>>8), uint8(col.B>>8))
	case color.Gray:
		v := col.Y
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", v, v, v)
	default:
		r, g, b, _ := c.RGBA()
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", uint8(r>>8), uint8(g>>8), uint8(b>>8))
	}
}