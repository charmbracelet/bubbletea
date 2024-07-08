package main

// A simple program demonstrating the textarea component from the Bubbles
// component library.

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type errMsg error

type model struct {
	textarea textarea.Model
	err      error
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	m.textarea = textarea.New(ctx)
	m.textarea.Placeholder = "Once upon a time..."
	m.textarea.Focus()

	return m, textarea.Blink
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
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

	m.textarea, cmd = m.textarea.Update(ctx, msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View(ctx tea.Context) string {
	return fmt.Sprintf(
		"Tell me a story.\n\n%s\n\n%s",
		m.textarea.View(ctx),
		"(ctrl+c to quit)",
	) + "\n\n"
}
