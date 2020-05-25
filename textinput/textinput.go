package textinput

import (
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

const (
	defaultBlinkSpeed = time.Millisecond * 600
)

var (
	// color is a helper for returning colors
	color func(s string) termenv.Color = termenv.ColorProfile().Color
)

// ErrMsg indicates there's been an error. We don't handle errors in the this
// package; we're expecting errors to be handle in the program that implements
// this text input.
type ErrMsg error

// Model is the Tea model for this text input element.
type Model struct {
	Err              error
	Prompt           string
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

	// Underlying text value
	value string

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

// SetValue sets the value of the text input.
func (m *Model) SetValue(s string) {
	if m.CharLimit > 0 && len(s) > m.CharLimit {
		m.value = s[:m.CharLimit]
	} else {
		m.value = s
	}
	if m.pos > len(m.value) {
		m.pos = len(m.value)
	}
	m.handleOverflow()
}

// Value returns the value of the text input.
func (m Model) Value() string {
	return m.value
}

// Cursor start moves the cursor to the given position. If the position is out
// of bounds the cursor will be moved to the start or end accordingly.
func (m *Model) SetCursor(pos int) {
	m.pos = max(0, min(len(m.value), pos))
	m.handleOverflow()
}

// CursorStart moves the cursor to the start of the field.
func (m *Model) CursorStart() {
	m.pos = 0
	m.handleOverflow()
}

// CursorEnd moves the cursor to the end of the field.
func (m *Model) CursorEnd() {
	m.pos = len(m.value)
	m.handleOverflow()
}

// Focused returns the focus state on the model.
func (m Model) Focused() bool {
	return m.focus
}

// Focus sets the focus state on the model.
func (m *Model) Focus() {
	m.focus = true
	m.blink = false
}

// Blur removes the focus state on the model.
func (m *Model) Blur() {
	m.focus = false
	m.blink = true
}

// Reset sets the input to its default state with no input.
func (m *Model) Reset() {
	m.value = ""
	m.offset = 0
	m.pos = 0
	m.blink = false
}

// If a max width is defined, perform some logic to treat the visible area
// as a horizontally scrolling viewport.
func (m *Model) handleOverflow() {
	if m.Width > 0 {
		overflow := max(0, len(m.value)-m.Width)

		if overflow > 0 && m.pos < m.offset {
			m.offset = max(0, min(len(m.value), m.pos))
		} else if overflow > 0 && m.pos >= m.offset+m.Width {
			m.offset = max(0, m.pos-m.Width)
		}
	}
}

// colorText colorizes a given string according to the TextColor value of the
// model.
func (m *Model) colorText(s string) string {
	return termenv.
		String(s).
		Foreground(color(m.TextColor)).
		Background(color(m.BackgroundColor)).
		String()
}

// colorPlaceholder colorizes a given string according to the TextColor value
// of the model.
func (m *Model) colorPlaceholder(s string) string {
	return termenv.
		String(s).
		Foreground(color(m.PlaceholderColor)).
		Background(color(m.BackgroundColor)).
		String()
}

func (m *Model) wordLeft() {
	if m.pos == 0 || len(m.value) == 0 {
		return
	}

	i := m.pos - 1

	for i >= 0 {
		if unicode.IsSpace(rune(m.value[i])) {
			m.pos--
			i--
		} else {
			break
		}
	}

	for i >= 0 {
		if !unicode.IsSpace(rune(m.value[i])) {
			m.pos--
			i--
		} else {
			break
		}
	}
}

func (m *Model) wordRight() {
	if m.pos >= len(m.value) || len(m.value) == 0 {
		return
	}

	i := m.pos

	for i < len(m.value) {
		if unicode.IsSpace(rune(m.value[i])) {
			m.pos++
			i++
		} else {
			break
		}
	}

	for i < len(m.value) {
		if !unicode.IsSpace(rune(m.value[i])) {
			m.pos++
			i++
		} else {
			break
		}
	}
}

// BlinkMsg is sent when the cursor should alternate it's blinking state.
type BlinkMsg struct{}

// NewModel creates a new model with default settings.
func NewModel() Model {
	return Model{
		Prompt:           "> ",
		BlinkSpeed:       defaultBlinkSpeed,
		Placeholder:      "",
		TextColor:        "",
		PlaceholderColor: "240",
		CursorColor:      "",
		CharLimit:        0,

		value: "",
		focus: false,
		blink: true,
		pos:   0,
	}
}

// Update is the Tea update loop.
func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	if !m.focus {
		m.blink = true
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			fallthrough
		case tea.KeyDelete:
			if len(m.value) > 0 {
				m.value = m.value[:m.pos-1] + m.value[m.pos:]
				m.pos--
			}
		case tea.KeyLeft:
			if msg.Alt { // alt+left arrow, back one word
				m.wordLeft()
				break
			}
			if m.pos > 0 {
				m.pos--
			}
		case tea.KeyRight:
			if msg.Alt { // alt+right arrow, forward one word
				m.wordRight()
				break
			}
			if m.pos < len(m.value) {
				m.pos++
			}
		case tea.KeyCtrlF: // ^F, forward one character
			fallthrough
		case tea.KeyCtrlB: // ^B, back one charcter
			fallthrough
		case tea.KeyCtrlA: // ^A, go to beginning
			m.CursorStart()
		case tea.KeyCtrlD: // ^D, delete char under cursor
			if len(m.value) > 0 && m.pos < len(m.value) {
				m.value = m.value[:m.pos] + m.value[m.pos+1:]
			}
		case tea.KeyCtrlE: // ^E, go to end
			m.CursorEnd()
		case tea.KeyCtrlK: // ^K, kill text after cursor
			m.value = m.value[:m.pos]
			m.pos = len(m.value)
		case tea.KeyCtrlU: // ^U, kill text before cursor
			m.value = m.value[m.pos:]
			m.pos = 0
			m.offset = 0
		case tea.KeyRune: // input a regular character

			if msg.Alt {
				if msg.Rune == 'b' { // alt+b, back one word
					m.wordLeft()
					break
				}
				if msg.Rune == 'f' { // alt+f, forward one word
					m.wordRight()
					break
				}
			}

			// Input a regular character
			if m.CharLimit <= 0 || len(m.value) < m.CharLimit {
				m.value = m.value[:m.pos] + string(msg.Rune) + m.value[m.pos:]
				m.pos++
			}
		}

	case ErrMsg:
		m.Err = msg

	case BlinkMsg:
		m.blink = !m.blink
		return m, Blink(m)
	}

	m.handleOverflow()

	return m, nil
}

// View renders the textinput in its current state.
func View(model tea.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "could not perform assertion on model"
	}

	// Placeholder text
	if m.value == "" && m.Placeholder != "" {
		return placeholderView(m)
	}

	left := m.offset
	right := 0
	if m.Width > 0 {
		right = min(len(m.value), m.offset+m.Width+1)
	} else {
		right = len(m.value)
	}
	value := m.value[left:right]
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

// cursorView styles the cursor.
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
func Blink(model Model) tea.Cmd {
	return func() tea.Msg {
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
