package prevent_quit

// A program demonstrating how to use the WithFilter option to intercept events.

import (
	"fmt"
	"log"

	tea "github.com/rprtr258/bubbletea"
	"github.com/rprtr258/bubbletea/bubbles/help"
	"github.com/rprtr258/bubbletea/bubbles/key"
	"github.com/rprtr258/bubbletea/bubbles/textarea"
	"github.com/rprtr258/bubbletea/lipgloss"
)

var (
	choiceStyle   = lipgloss.NewStyle().PaddingLeft(1).Foreground(lipgloss.Color("241"))
	saveTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	quitViewStyle = lipgloss.NewStyle().Padding(1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("170"))
)

func Main() {
	p := tea.NewProgram(initialModel()).WithFilter(filter)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func filter(m model, msg tea.Msg) tea.Msg {
	if _, ok := msg.(tea.QuitMsg); !ok {
		return msg
	}

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

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	if m.quitting {
		return m.updatePromptView(msg)
	}

	return m.updateTextView(msg)
}

func (m model) updateTextView(msg tea.Msg) (model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.MsgKey:
		m.saveText = ""
		switch {
		case key.Matches(msg, m.keymap.save):
			m.saveText = "Changes saved!"
			m.hasChanges = false
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		case msg.Type == tea.KeyRunes:
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

func (m model) updatePromptView(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MsgKey:
		// For simplicity's sake, we'll treat any key besides "y" as "no"
		if key.Matches(msg, m.keymap.quit) || msg.String() == "y" {
			m.hasChanges = false
			return m, tea.Quit
		}
		m.quitting = false
	}

	return m, nil
}

func (m model) View(r tea.Renderer) {
	if m.quitting {
		if m.hasChanges {
			text := lipgloss.JoinHorizontal(lipgloss.Top, "You have unsaved changes. Quit without saving?", choiceStyle.Render("[yn]"))
			r.Write(quitViewStyle.Render(text))
			return
		}

		r.Write("Very important, thank you\n")
		return
	}

	helpView := m.help.ShortHelpView([]key.Binding{
		m.keymap.save,
		m.keymap.quit,
	})

	r.Write(fmt.Sprintf(
		"\nType some important things.\n\n%s\n\n %s\n %s",
		m.textarea.View(),
		saveTextStyle.Render(m.saveText),
		helpView,
	) + "\n\n")
}
