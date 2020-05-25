package spinner

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

// Spinner is a set of frames used in animating the spinner.
type Spinner = int

// Available types of spinners
const (
	Line Spinner = iota
	Dot
)

const (
	defaultFPS = 9
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

	// Type is the set of frames to use. See Spinner.
	Type Spinner

	// FPS is the speed at which the ticker should tick
	FPS int

	// ForegroundColor sets the background color of the spinner. It can be a
	// hex code or one of the 256 ANSI colors. If the terminal emulator can't
	// doesn't support the color specified it will automatically degrade
	// (per github.com/muesli/termenv).
	ForegroundColor string

	// BackgroundColor sets the background color of the spinner. It can be a
	// hex code or one of the 256 ANSI colors. If the terminal emulator can't
	// doesn't support the color specified it will automatically degrade
	// (per github.com/muesli/termenv).
	BackgroundColor string

	// CustomMsgFunc can be used to a custom message on tick. This can be
	// useful when you have spinners in different parts of your application and
	// want to differentiate between the messages for clarity and simplicity.
	// If nil, this setting is ignored.
	CustomMsgFunc func() tea.Msg

	frame int
}

// NewModel returns a model with default values.
func NewModel() Model {
	return Model{
		Type: Line,
		FPS:  defaultFPS,
	}
}

// TickMsg indicates that the timer has ticked and we should render a frame.
type TickMsg struct{}

// Update is the Tea update function. This will advance the spinner one frame
// every time it's called, regardless the message passed, so be sure the logic
// is setup so as not to call this Update needlessly.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	m.frame++
	if m.frame >= len(spinners[m.Type]) {
		m.frame = 0
	}
	if m.CustomMsgFunc != nil {
		return m, Tick(m)
	}
	return m, Tick(m)
}

// View renders the model's view.
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
func Tick(model Model) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Second / time.Duration(model.FPS))
		if model.CustomMsgFunc != nil {
			return model.CustomMsgFunc()
		}
		return TickMsg{}
	}
}
