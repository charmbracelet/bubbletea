package main

// A simple program demonstrating the textarea component from the Bubbles
// component library.

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/v2/cursor"
	"github.com/charmbracelet/bubbles/v2/textarea"
	tea "github.com/charmbracelet/bubbletea/v2"
)

func main() {
	// p := tea.NewProgram(initialModel())
	m := initialModel()
	p := &tea.Program[model]{
		Init: m.Init,
		Update: func(m model, msg tea.Msg) (model, tea.Cmd) {
			return m.Update(msg)
		},
		View: func(m model) fmt.Stringer {
			return m.View()
		},
	}

	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type errMsg error

type model struct {
	textarea textarea.Model
	err      error
}

func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "Once upon a time..."
	ti.Focus()
	ti.VirtualCursor.SetMode(cursor.CursorHide)

	return model{
		textarea: ti,
		err:      nil,
	}
}

func (m model) Init() (model, tea.Cmd) {
	return m, tea.Batch(
		textarea.Blink,
	)
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case "ctrl+c":
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() fmt.Stringer {
	const (
		xOffset = 6
		yOffset = 2
	)
	f := tea.NewFrame(fmt.Sprintf(
		"Tell me a story.\n\n%s\n\n%s",
		m.textarea.View(),
		"(ctrl+c to quit)",
	) + "\n\n")

	cur := m.textarea.Cursor()
	x, y := cur.Position.X, cur.Position.Y
	f.Cursor = tea.NewCursor(x+xOffset, y+yOffset)

	return f
}
