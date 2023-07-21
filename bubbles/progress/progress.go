package progress

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/harmonica"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/termenv"
	"github.com/rprtr258/bubbletea/lipgloss"

	tea "github.com/rprtr258/bubbletea"
)

// Internal ID management. Used during animating to assure that frame messages
// can only be received by progress components that sent them.
var (
	lastID int
	idMtx  sync.Mutex
)

// Return the next ID we should use on the model.
func nextID() int {
	idMtx.Lock()
	defer idMtx.Unlock()
	lastID++
	return lastID
}

const (
	fps              = 60
	defaultWidth     = 40
	defaultFrequency = 18.0
	defaultDamping   = 1.0
)

// Option is used to set options in New. For example:
//
//	    progress := New(
//		       WithRamp("#ff0000", "#0000ff"),
//		       WithoutPercentage(),
//	    )
type Option func(*Model)

// WithDefaultGradient sets a gradient fill with default colors.
func WithDefaultGradient() Option {
	return WithGradient("#5A56E0", "#EE6FF8")
}

// WithGradient sets a gradient fill blending between two colors.
func WithGradient(colorA, colorB string) Option {
	return func(m *Model) {
		m.setRamp(colorA, colorB, false)
	}
}

// WithDefaultScaledGradient sets a gradient with default colors, and scales the
// gradient to fit the filled portion of the ramp.
func WithDefaultScaledGradient() Option {
	return WithScaledGradient("#5A56E0", "#EE6FF8")
}

// WithScaledGradient scales the gradient to fit the width of the filled portion of
// the progress bar.
func WithScaledGradient(colorA, colorB string) Option {
	return func(m *Model) {
		m.setRamp(colorA, colorB, true)
	}
}

// WithSolidFill sets the progress to use a solid fill with the given color.
func WithSolidFill(color string) Option {
	return func(m *Model) {
		m.FullColor = color
		m.useRamp = false
	}
}

// WithoutPercentage hides the numeric percentage.
func WithoutPercentage() Option {
	return func(m *Model) {
		m.ShowPercentage = false
	}
}

// WithWidth sets the initial width of the progress bar. Note that you can also
// set the width via the Width property, which can come in handy if you're
// waiting for a tea.WindowSizeMsg.
func WithWidth(w int) Option {
	return func(m *Model) {
		m.Width = w
	}
}

// WithSpringOptions sets the initial frequency and damping options for the
// progress bar's built-in spring-based animation. Frequency corresponds to
// speed, and damping to bounciness. For details see:
//
// https://github.com/charmbracelet/harmonica
func WithSpringOptions(frequency, damping float64) Option {
	return func(m *Model) {
		m.SetSpringOptions(frequency, damping)
		m.springCustomized = true
	}
}

// WithColorProfile sets the color profile to use for the progress bar.
func WithColorProfile(p termenv.Profile) Option {
	return func(m *Model) {
		m.colorProfile = p
	}
}

// FrameMsg indicates that an animation step should occur.
type FrameMsg struct {
	id  int
	tag int
}

// Model stores values we'll use when rendering the progress bar.
type Model struct {
	// An identifier to keep us from receiving messages intended for other
	// progress bars.
	id int

	// An identifier to keep us from receiving frame messages too quickly.
	tag int

	// Total width of the progress bar, including percentage, if set.
	Width int

	// "Filled" sections of the progress bar.
	Full      rune
	FullColor string

	// "Empty" sections of the progress bar.
	Empty      rune
	EmptyColor string

	// Settings for rendering the numeric percentage.
	ShowPercentage  bool
	PercentFormat   string // a fmt string for a float
	PercentageStyle lipgloss.Style

	// Members for animated transitions.
	spring           harmonica.Spring
	springCustomized bool
	percentShown     float64 // percent currently displaying
	targetPercent    float64 // percent to which we're animating
	velocity         float64

	// Gradient settings
	useRamp    bool
	rampColorA colorful.Color
	rampColorB colorful.Color

	// When true, we scale the gradient to fit the width of the filled section
	// of the progress bar. When false, the width of the gradient will be set
	// to the full width of the progress bar.
	scaleRamp bool

	// Color profile for the progress bar.
	colorProfile termenv.Profile
}

// New returns a model with default values.
func New(opts ...Option) Model {
	m := Model{
		id:             nextID(),
		Width:          defaultWidth,
		Full:           '█',
		FullColor:      "#7571F9",
		Empty:          '░',
		EmptyColor:     "#606060",
		ShowPercentage: true,
		PercentFormat:  " %3.0f%%",
		colorProfile:   termenv.ColorProfile(),
	}
	if !m.springCustomized {
		m.SetSpringOptions(defaultFrequency, defaultDamping)
	}

	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// NewModel returns a model with default values.
//
// Deprecated: use [New] instead.
var NewModel = New

// Init exists to satisfy the tea.Model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is used to animate the progress bar during transitions. Use
// SetPercent to create the command you'll need to trigger the animation.
//
// If you're rendering with ViewAs you won't need this.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FrameMsg:
		if msg.id != m.id || msg.tag != m.tag {
			return m, nil
		}

		// If we've more or less reached equilibrium, stop updating.
		if !m.IsAnimating() {
			return m, nil
		}

		m.percentShown, m.velocity = m.spring.Update(m.percentShown, m.velocity, m.targetPercent)
		return m, m.nextFrame()

	default:
		return m, nil
	}
}

// SetSpringOptions sets the frequency and damping for the current spring.
// Frequency corresponds to speed, and damping to bounciness. For details see:
//
// https://github.com/charmbracelet/harmonica
func (m *Model) SetSpringOptions(frequency, damping float64) {
	m.spring = harmonica.NewSpring(harmonica.FPS(fps), frequency, damping)
}

// Percent returns the current visible percentage on the model. This is only
// relevant when you're animating the progress bar.
//
// If you're rendering with ViewAs you won't need this.
func (m Model) Percent() float64 {
	return m.targetPercent
}

// SetPercent sets the percentage state of the model as well as a command
// necessary for animating the progress bar to this new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *Model) SetPercent(p float64) tea.Cmd {
	m.targetPercent = math.Max(0, math.Min(1, p))
	m.tag++
	return m.nextFrame()
}

// IncrPercent increments the percentage by a given amount, returning a command
// necessary to animate the progress bar to the new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *Model) IncrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() + v)
}

// DecrPercent decrements the percentage by a given amount, returning a command
// necessary to animate the progress bar to the new percentage.
//
// If you're rendering with ViewAs you won't need this.
func (m *Model) DecrPercent(v float64) tea.Cmd {
	return m.SetPercent(m.Percent() - v)
}

// View renders an animated progress bar in its current state. To render
// a static progress bar based on your own calculations use ViewAs instead.
func (m Model) View() string {
	return m.ViewAs(m.percentShown)
}

// ViewAs renders the progress bar with a given percentage.
func (m Model) ViewAs(percent float64) string {
	b := strings.Builder{}
	percentView := m.percentageView(percent)
	m.barView(&b, percent, ansi.PrintableRuneWidth(percentView))
	b.WriteString(percentView)
	return b.String()
}

func (m *Model) nextFrame() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(fps), func(time.Time) tea.Msg {
		return FrameMsg{id: m.id, tag: m.tag}
	})
}

func (m Model) barView(b *strings.Builder, percent float64, textWidth int) {
	var (
		tw = max(0, m.Width-textWidth)                // total width
		fw = int(math.Round((float64(tw) * percent))) // filled width
		p  float64
	)

	fw = max(0, min(tw, fw))

	if m.useRamp {
		// Gradient fill
		for i := 0; i < fw; i++ {
			switch {
			case fw == 1:
				// this is up for debate: in a gradient of width=1, should the
				// single character rendered be the first color, the last color
				// or exactly 50% in between? I opted for 50%
				p = 0.5
			case m.scaleRamp:
				p = float64(i) / float64(fw-1)
			default:
				p = float64(i) / float64(tw-1)
			}
			c := m.rampColorA.BlendLuv(m.rampColorB, p).Hex()
			b.WriteString(termenv.
				String(string(m.Full)).
				Foreground(m.color(c)).
				String(),
			)
		}
	} else {
		// Solid fill
		s := termenv.String(string(m.Full)).Foreground(m.color(m.FullColor)).String()
		b.WriteString(strings.Repeat(s, fw))
	}

	// Empty fill
	e := termenv.String(string(m.Empty)).Foreground(m.color(m.EmptyColor)).String()
	n := max(0, tw-fw)
	b.WriteString(strings.Repeat(e, n))
}

func (m Model) percentageView(percent float64) string {
	if !m.ShowPercentage {
		return ""
	}
	percent = math.Max(0, math.Min(1, percent))
	percentage := fmt.Sprintf(m.PercentFormat, percent*100) //nolint:gomnd
	percentage = m.PercentageStyle.Inline(true).Render(percentage)
	return percentage
}

func (m *Model) setRamp(colorA, colorB string, scaled bool) {
	// In the event of an error colors here will default to black. For
	// usability's sake, and because such an error is only cosmetic, we're
	// ignoring the error.
	a, _ := colorful.Hex(colorA)
	b, _ := colorful.Hex(colorB)

	m.useRamp = true
	m.scaleRamp = scaled
	m.rampColorA = a
	m.rampColorB = b
}

func (m Model) color(c string) termenv.Color {
	return m.colorProfile.Color(c)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IsAnimating returns false if the progress bar reached equilibrium and is no longer animating.
func (m *Model) IsAnimating() bool {
	dist := math.Abs(m.percentShown - m.targetPercent)
	return !(dist < 0.001 && m.velocity < 0.01)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
