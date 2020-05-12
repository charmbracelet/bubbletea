package spinner

import (
	"errors"
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

	assertionErr = errors.New("could not perform assertion on model to what the spinner expects. are you sure you passed the right value?")

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
type TickMsg time.Time

// Update is the Boba update function
func Update(msg boba.Msg, m Model) (Model, boba.Cmd) {
	switch msg.(type) {
	case TickMsg:
		m.frame++
		if m.frame >= len(spinners[m.Type]) {
			m.frame = 0
		}
		return m, nil
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

// GetSub creates the subscription that allows the spinner to spin. Remember
// that you need to execute this function in order to get the subscription
// you'll need.
func MakeSub(model boba.Model) (boba.Sub, error) {
	m, ok := model.(Model)
	if !ok {
		return nil, assertionErr
	}
	return boba.Tick(time.Second/time.Duration(m.FPS), func(t time.Time) boba.Msg {
		return TickMsg(t)
	}), nil
}
