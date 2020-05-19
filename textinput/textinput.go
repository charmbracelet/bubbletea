package textinput

import (
	"time"

	"github.com/charmbracelet/boba"
	"github.com/muesli/termenv"
)

var (
	// color is a helper for returning colors
	color func(s string) termenv.Color = termenv.ColorProfile().Color
)

// ErrMsg indicates there's been an error. We don't handle errors in the this
// package; we're expecting errors to be handle in the program that implements
// this text input.
type ErrMsg error

// Model is the Boba model for this text input element
type Model struct {
	Err              error
	Prompt           string
	Value            string
	Cursor           string
	BlinkSpeed       time.Duration
	Placeholder      string
	TextColor        string
	BackgroundColor  string
	PlaceholderColor string
	CursorColor      string

	// CharLimit is the maximum amount of characters this input element will
	// accept. If 0 or less, there's no limit.
	CharLimit int

	// Focus indicates whether user input focus should be on this input
	// component. When false, don't blink and ignore keyboard input.
	focus bool

	blink bool
	pos   int
}

// Focused returns the focus state on the model
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model
func (m *Model) Focus() {
	m.focus = true
	m.blink = false
}

// Blur removes the focus state on the model
func (m *Model) Blur() {
	m.focus = false
	m.blink = true
}

// colorText colorizes a given string according to the TextColor value of the
// model
func (m *Model) colorText(s string) string {
	return termenv.
		String(s).
		Foreground(color(m.TextColor)).
		Background(color(m.BackgroundColor)).
		String()
}

// colorPlaceholder colorizes a given string according to the TextColor value
// of the model
func (m *Model) colorPlaceholder(s string) string {
	return termenv.
		String(s).
		Foreground(color(m.PlaceholderColor)).
		Background(color(m.BackgroundColor)).
		String()
}

// BlinkMsg is sent when the cursor should alternate it's blinking state
type BlinkMsg struct{}

// NewModel creates a new model with default settings
func NewModel() Model {
	return Model{
		Prompt:           "> ",
		Value:            "",
		BlinkSpeed:       time.Millisecond * 600,
		Placeholder:      "",
		TextColor:        "",
		PlaceholderColor: "240",
		CursorColor:      "",
		CharLimit:        0,

		focus: false,
		blink: true,
		pos:   0,
	}
}

// Update is the Boba update loop
func Update(msg boba.Msg, m Model) (Model, boba.Cmd) {
	if !m.focus {
		m.blink = true
		return m, nil
	}

	switch msg := msg.(type) {

	case boba.KeyMsg:
		switch msg.Type {
		case boba.KeyBackspace:
			fallthrough
		case boba.KeyDelete:
			if len(m.Value) > 0 {
				m.Value = m.Value[:m.pos-1] + m.Value[m.pos:]
				m.pos--
			}
			return m, nil
		case boba.KeyLeft:
			if m.pos > 0 {
				m.pos--
			}
			return m, nil
		case boba.KeyRight:
			if m.pos < len(m.Value) {
				m.pos++
			}
			return m, nil
		case boba.KeyCtrlF: // ^F, forward one character
			fallthrough
		case boba.KeyCtrlB: // ^B, back one charcter
			fallthrough
		case boba.KeyCtrlA: // ^A, go to beginning
			m.pos = 0
			return m, nil
		case boba.KeyCtrlD: // ^D, delete char under cursor
			if len(m.Value) > 0 && m.pos < len(m.Value) {
				m.Value = m.Value[:m.pos] + m.Value[m.pos+1:]
			}
			return m, nil
		case boba.KeyCtrlE: // ^E, go to end
			m.pos = len(m.Value)
			return m, nil
		case boba.KeyCtrlK: // ^K, kill text after cursor
			m.Value = m.Value[:m.pos]
			m.pos = len(m.Value)
			return m, nil
		case boba.KeyCtrlU: // ^U, kill text before cursor
			m.Value = m.Value[m.pos:]
			m.pos = 0
			return m, nil
		case boba.KeyRune: // input a regular character
			if m.CharLimit <= 0 || len(m.Value) < m.CharLimit {
				m.Value = m.Value[:m.pos] + string(msg.Rune) + m.Value[m.pos:]
				m.pos++
			}
			return m, nil
		default:
			return m, nil
		}

	case ErrMsg:
		m.Err = msg
		return m, nil

	case BlinkMsg:
		m.blink = !m.blink
		return m, Blink(m)

	default:
		return m, nil
	}
}

// View renders the textinput in its current state
func View(model boba.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "could not perform assertion on model"
	}

	// Placeholder text
	if m.Value == "" && m.Placeholder != "" {
		return placeholderView(m)
	}

	v := m.colorText(m.Value[:m.pos])

	if m.pos < len(m.Value) {
		v += cursorView(string(m.Value[m.pos]), m)
		v += m.colorText(m.Value[m.pos+1:])
	} else {
		v += cursorView(" ", m)
	}
	return m.Prompt + v
}

// placeholderView
func placeholderView(m Model) string {
	var (
		v string
		p = m.Placeholder
	)

	// Cursor
	if m.blink && m.PlaceholderColor != "" {
		v += cursorView(
			m.colorPlaceholder(p[:1]),
			m,
		)
	} else {
		v += cursorView(p[:1], m)
	}

	// The rest of the placeholder text
	v += m.colorPlaceholder(p[1:])

	return m.Prompt + v
}

// cursorView style the cursor
func cursorView(s string, m Model) string {
	if m.blink {
		return s
	}
	return termenv.String(s).
		Foreground(color(m.CursorColor)).
		Reverse().
		String()
}

// Blink is a command used to time the cursor blinking.
func Blink(model Model) boba.Cmd {
	return func() boba.Msg {
		time.Sleep(model.BlinkSpeed)
		return BlinkMsg{}
	}
}
