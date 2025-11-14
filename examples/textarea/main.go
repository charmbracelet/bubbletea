package main

// A simple program demonstrating the textarea component from the Bubbles
// component library.

import (
	"log"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	ti.SetVirtualCursor(false)
	ti.SetStyles(textarea.DefaultStyles(true)) // default to dark styles.
	ti.Focus()

	return model{
		textarea: ti,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.RequestBackgroundColor)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// Update styling now that we know the background color.
		m.textarea.SetStyles(textarea.DefaultStyles(msg.IsDark()))

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

func (m model) headerView() string {
	return "Tell me a story.\n"
}

func (m model) View() tea.View {
	const (
		footer = "\n(ctrl+c to quit)\n"
	)

	var c *tea.Cursor
	if !m.textarea.VirtualCursor() {
		c = m.textarea.Cursor()

		// Set the y offset of the cursor based on the position of the textarea
		// in the application.
		offset := lipgloss.Height(m.headerView())
		c.Y += offset
	}

	f := strings.Join([]string{
		m.headerView(),
		m.textarea.View(),
		footer,
	}, "\n")

	v := tea.NewView(f)
	v.Cursor = c
	return v
}
