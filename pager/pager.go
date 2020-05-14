package pager

import (
	"errors"
	"os"
	"strings"

	"github.com/charmbracelet/boba"
	"golang.org/x/crypto/ssh/terminal"
)

// MSG

type terminalSizeMsg struct {
	width  int
	height int
}

type errMsg error

// MODEL

type State int

const (
	StateInit State = iota
	StateReady
)

type Model struct {
	Err        error
	Standalone bool
	State      State
	Width      int
	Height     int
	Y          int

	lines []string
}

func (m *Model) PageUp() {
	m.Y = max(0, m.Y-m.Height)
}

func (m *Model) PageDown() {
	m.Y = min(len(m.lines)-m.Height, m.Y+m.Height)
}

// Content adds text content to the model
func (m *Model) Content(s string) {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\r\n", "\n", -1) // normalize line endings
	m.lines = strings.Split(s, "\n")
}

func NewModel() Model {
	return Model{
		State: StateInit,
	}
}

// INIT

func Init(initialContent string) func() (boba.Model, boba.Cmd) {
	m := NewModel()
	m.Standalone = true
	m.Content(initialContent)
	return func() (boba.Model, boba.Cmd) {
		return m, GetTerminalSize
	}
}

// UPDATE

func Update(msg boba.Msg, model boba.Model) (boba.Model, boba.Cmd) {
	m, ok := model.(Model)
	if !ok {
		return Model{
			Err: errors.New("could not perform assertion on model in update in pager; are you sure you passed the correct model?"),
		}, nil
	}

	switch msg := msg.(type) {

	case boba.KeyMsg:
		switch msg.String() {
		case "q":
			fallthrough
		case "ctrl+c":
			if m.Standalone {
				return m, boba.Quit
			}

		// Up one page
		case "pgup":
			fallthrough
		case "b":
			m.Y = max(0, m.Y-m.Height)
			return m, nil

		// Down one page
		case "pgdown":
			fallthrough
		case "space":
			fallthrough
		case "f":
			m.Y = min(len(m.lines)-m.Height, m.Y+m.Height)
			return m, nil

		// Up half page
		case "u":
			m.Y = max(0, m.Y-m.Height/2)
			return m, nil

		// Down half page
		case "d":
			m.Y = min(len(m.lines)-m.Height, m.Y+m.Height/2)
			return m, nil

		// Up one line
		case "up":
			fallthrough
		case "k":
			m.Y = max(0, m.Y-1)
			return m, nil

		// Down one line
		case "down":
			fallthrough
		case "j":
			m.Y = min(len(m.lines)-m.Height, m.Y+1)
			return m, nil

		// Re-render
		case "ctrl+l":
			return m, GetTerminalSize

		}

	case errMsg:
		m.Err = msg
		return m, nil

	case terminalSizeMsg:
		m.Width = msg.width
		m.Height = msg.height
		m.State = StateReady
		return m, nil
	}

	return model, nil
}

// VIEW

func View(model boba.Model) string {
	m, ok := model.(Model)
	if !ok {
		return "could not perform assertion on model in view in pager; are you sure you passed the correct model?"
	}

	if m.Err != nil {
		return m.Err.Error()
	}

	if len(m.lines) == 0 {
		return "(Buffer empty)"
	}

	if m.State == StateReady {
		// Render viewport
		top := max(0, m.Y)
		bottom := min(len(m.lines), m.Y+m.Height)
		lines := m.lines[top:bottom]
		return "\n" + strings.Join(lines, "\n")
	}

	return ""
}

// CMD

func GetTerminalSize() boba.Msg {
	w, h, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return errMsg(err)
	}
	return terminalSizeMsg{w, h}
}

// ETC

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
