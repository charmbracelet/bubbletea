package main

// A program demonstrating how to use the WithFilter option to intercept events.

import (
	"fmt"
	"log"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	choiceStyle   = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("241"))
	saveTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	quitViewStyle = lipgloss.NewStyle().Padding(1, 3).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170"))
)

func main() {
	p := tea.NewProgram(initialModel(), tea.WithFilter(filter))

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func filter(teaModel tea.Model, msg tea.Msg) tea.Msg {
	if _, ok := msg.(tea.QuitMsg); !ok {
		return msg
	}

	m := teaModel.(model)
	if m.hasChanges {
		return nil
	}

	return msg
}

type model struct {
	textarea   textarea.Model
	help       help.Model
	keymap     keymap
	saveText   string
	hasChanges bool
	quitting   bool
}

type keymap struct {
	save key.Binding
	quit key.Binding
}

func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "Only the best words"
	ti.Focus()

	return model{
		textarea: ti,
		help:     help.New(),
		keymap: keymap{
			save: key.NewBinding(
				key.WithKeys("ctrl+s"),
				key.WithHelp("ctrl+s", "save"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m.updatePromptView(msg)
	}

	return m.updateTextView(msg)
}

func (m model) updateTextView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m.saveText = ""
		switch {
		case key.Matches(msg, m.keymap.save):
			m.saveText = "Changes saved!"
			m.hasChanges = false
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		case len(msg.Text) > 0:
			m.saveText = ""
			m.hasChanges = true
			fallthrough
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) updatePromptView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// For simplicity's sake, we'll treat any key besides "y" as "no"
		if key.Matches(msg, m.keymap.quit) || msg.String() == "y" {
			m.hasChanges = false
			return m, tea.Quit
		}
		m.quitting = false
	}

	return m, nil
}

func (m model) View() tea.View {
	if m.quitting {
		if m.hasChanges {
			text := lipgloss.JoinHorizontal(lipgloss.Top, "You have unsaved changes. Quit without saving?", choiceStyle.Render("[yN]"))
			return tea.NewView(quitViewStyle.Render(text))
		}
		return tea.NewView("Very important. Thank you.\n")
	}

	helpView := m.help.ShortHelpView([]key.Binding{
		m.keymap.save,
		m.keymap.quit,
	})

	return tea.NewView(fmt.Sprintf(
		"Type some important things.\n%s\n %s\n %s",
		m.textarea.View(),
		saveTextStyle.Render(m.saveText),
		helpView,
	) + "\n\n")
}
