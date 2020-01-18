package input

import (
	"tea"
	"time"
)

type Model struct {
	Prompt       string
	Value        string
	Cursor       string
	HiddenCursor string
	Blink        bool
	BlinkSpeed   time.Duration
}

type CursorBlinkMsg struct{}

func DefaultModel() Model {
	return Model{
		Prompt:       "> ",
		Value:        "",
		Cursor:       "█",
		HiddenCursor: " ",
		Blink:        false,
		BlinkSpeed:   time.Second,
	}
}

func Update(msg tea.Msg, m Model) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyBackspace:
			if len(m.Value) > 0 {
				m.Value = m.Value[:len(m.Value)-1]
			}
		case tea.KeyRune:
			m.Value = m.Value + msg.String()
		}

	case CursorBlinkMsg:
		m.Blink = !m.Blink
	}

	return m, nil
}

func View(model tea.Model) string {
	m, _ := model.(Model)
	cursor := m.Cursor
	if m.Blink {
		cursor = m.HiddenCursor
	}
	return m.Prompt + m.Value + cursor
}

func Blink(model tea.Model) tea.Msg {
	m, _ := model.(Model)
	time.Sleep(m.BlinkSpeed)
	return CursorBlinkMsg{}
}
