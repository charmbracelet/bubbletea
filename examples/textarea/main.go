package main

// A simple program demonstrating the textarea component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/v2/textarea"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
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
	ti.VirtualCursor = false
	ti.Focus()

	return model{
		textarea: ti,
		err:      nil,
	}
}

func (m model) Init() (tea.Model, tea.Cmd) {
	return m, tea.Batch(textarea.Blink, tea.RequestBackgroundColor)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.textarea.Styles = textarea.DefaultStyles(msg.IsDark())

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
		header = "Tell me a story.\n"
		footer = "\n(ctrl+c to quit)\n"
	)

	f := tea.NewFrame(strings.Join([]string{
		header,
		m.textarea.View(),
		footer,
	}, "\n"))

	if !m.textarea.VirtualCursor {
		f.Cursor = m.textarea.Cursor()

		// Set the y offset of the cursor based on the position of the textarea
		// in the application.
		f.Cursor.Position.Y += lipgloss.Height(header)
	}

	return f
}
