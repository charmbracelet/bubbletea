package textinput

import (
	"strings"
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

	// Width is the maximum number of characters that can be displayed at once.
	// It essentially treats the text field like a horizontally scrolling
	// viewport. If 0 or less this setting is ignored.
	Width int

	// Focus indicates whether user input focus should be on this input
	// component. When false, don't blink and ignore keyboard input.
	focus bool

	// Cursor blink state
	blink bool

	// Cursor position
	pos int

	// Used to emulate a viewport when width is set and the content is
	// overflowing
	offset int
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
		case boba.KeyLeft:
			if m.pos > 0 {
				m.pos--
			}
		case boba.KeyRight:
			if m.pos < len(m.Value) {
				m.pos++
			}
		case boba.KeyCtrlF: // ^F, forward one character
			fallthrough
		case boba.KeyCtrlB: // ^B, back one charcter
			fallthrough
		case boba.KeyCtrlA: // ^A, go to beginning
			m.pos = 0
		case boba.KeyCtrlD: // ^D, delete char under cursor
			if len(m.Value) > 0 && m.pos < len(m.Value) {
				m.Value = m.Value[:m.pos] + m.Value[m.pos+1:]
			}
		case boba.KeyCtrlE: // ^E, go to end
			m.pos = len(m.Value)
		case boba.KeyCtrlK: // ^K, kill text after cursor
			m.Value = m.Value[:m.pos]
			m.pos = len(m.Value)
		case boba.KeyCtrlU: // ^U, kill text before cursor
			m.Value = m.Value[m.pos:]
			m.pos = 0
			m.offset = 0
		case boba.KeyRune: // input a regular character
			if m.CharLimit <= 0 || len(m.Value) < m.CharLimit {
				m.Value = m.Value[:m.pos] + string(msg.Rune) + m.Value[m.pos:]
				m.pos++
			}
		}

	case ErrMsg:
		m.Err = msg

	case BlinkMsg:
		m.blink = !m.blink
		return m, Blink(m)
	}

	// If a max width is defined, perform some logic to treat the visible area
	// as a horizontally scrolling mini viewport.
	if m.Width > 0 {
		overflow := max(0, len(m.Value)-m.Width)
		if overflow > 0 && m.pos < m.offset {
			m.offset = max(0, min(len(m.Value), m.pos))
		} else if overflow > 0 && m.pos >= m.offset+m.Width {
			m.offset = max(0, m.pos-m.Width)
		}
	}

	return m, nil
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

	left := m.offset
	right := 0
	if m.Width > 0 {
		right = min(len(m.Value), m.offset+m.Width+1)
	} else {
		right = len(m.Value)
	}
	value := m.Value[left:right]
	pos := m.pos - m.offset

	v := m.colorText(value[:pos])

	if pos < len(value) {
		v += cursorView(string(value[pos]), m) // cursor and text under it
		v += m.colorText(value[pos+1:])        // text after cursor
	} else {
		v += cursorView(" ", m)
	}

	// If a max width and background color were set fill the empty spaces with
	// the background color.
	if m.Width > 0 && len(m.BackgroundColor) > 0 && len(value) <= m.Width {
		padding := m.Width - len(value)
		if len(value)+padding <= m.Width && pos < len(value) {
			padding++
		}
		v += strings.Repeat(
			termenv.String(" ").Background(color(m.BackgroundColor)).String(),
			padding,
		)
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

// cursorView styles the cursor
func cursorView(s string, m Model) string {
	if m.blink {
		if m.TextColor != "" || m.BackgroundColor != "" {
			return termenv.String(s).
				Foreground(color(m.TextColor)).
				Background(color(m.BackgroundColor)).
				String()
		}
		return s
	}
	return termenv.String(s).
		Foreground(color(m.CursorColor)).
		Background(color(m.BackgroundColor)).
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
