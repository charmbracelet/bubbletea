package spinner

import (
	"time"

	"github.com/charmbracelet/boba"
	"github.com/muesli/termenv"
)

// Spinner denotes a type of spinner
type Spinner = int

// Available types of spinners
const (
	Line Spinner = iota
	Dot
)

var (
	// Spinner frames
	spinners = map[Spinner][]string{
		Line: {"|", "/", "-", "\\"},
		Dot:  {"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "},
	}

	color = termenv.ColorProfile().Color
)

// Model contains the state for the spinner. Use NewModel to create new models
// rather than using Model as a struct literal.
type Model struct {
	Type            Spinner
	FPS             int
	ForegroundColor string
	BackgroundColor string

	frame int
}

// NewModel returns a model with default values
func NewModel() Model {
	return Model{
		Type:  Line,
		FPS:   9,
		frame: 0,
	}
}

// TickMsg indicates that the timer has ticked and we should render a frame
type TickMsg struct{}

// Update is the Boba update function
func Update(msg boba.Msg, m Model) (Model, boba.Cmd) {
	switch msg.(type) {
	case TickMsg:
		m.frame++
		if m.frame >= len(spinners[m.Type]) {
			m.frame = 0
		}
		return m, Tick(m)
	default:
		return m, nil
	}
}

// View renders the model's view
func View(model Model) string {
	s := spinners[model.Type]
	if model.frame >= len(s) {
		return "[error]"
	}

	str := s[model.frame]

	if model.ForegroundColor != "" || model.BackgroundColor != "" {
		return termenv.
			String(str).
			Foreground(color(model.ForegroundColor)).
			Background(color(model.BackgroundColor)).
			String()
	}

	return str
}

// Tick is the command used to advance the spinner one frame.
func Tick(model Model) boba.Cmd {
	return func() boba.Msg {
		time.Sleep(time.Second / time.Duration(model.FPS))
		return TickMsg{}
	}
}
