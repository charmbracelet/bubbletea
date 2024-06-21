package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	initialInputs = 2
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

type styles struct {
	cursorStyle             lipgloss.Style
	cursorLineStyle         lipgloss.Style
	placeholderStyle        lipgloss.Style
	endOfBufferStyle        lipgloss.Style
	focusedPlaceholderStyle lipgloss.Style
	focusedBorderStyle      lipgloss.Style
	blurredBorderStyle      lipgloss.Style
}

type keymap = struct {
	next, prev, add, remove, quit key.Binding
}

func (m model) newTextarea(ctx tea.Context) textarea.Model {
	t := textarea.New(ctx)
	t.Prompt = ""
	t.Placeholder = "Type something"
	t.ShowLineNumbers = true
	t.Cursor.Style = m.styles.cursorStyle
	t.FocusedStyle.Placeholder = m.styles.focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = m.styles.placeholderStyle
	t.FocusedStyle.CursorLine = m.styles.cursorLineStyle
	t.FocusedStyle.Base = m.styles.focusedBorderStyle
	t.BlurredStyle.Base = m.styles.blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = m.styles.endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = m.styles.endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()
	return t
}

type model struct {
	width  int
	height int
	keymap keymap
	help   help.Model
	inputs []textarea.Model
	focus  int
	styles *styles
}

func newModel(ctx tea.Context) model {
	m := model{
		inputs: make([]textarea.Model, initialInputs),
		help:   help.New(ctx),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev"),
			),
			add: key.NewBinding(
				key.WithKeys("ctrl+n"),
				key.WithHelp("ctrl+n", "add an editor"),
			),
			remove: key.NewBinding(
				key.WithKeys("ctrl+w"),
				key.WithHelp("ctrl+w", "remove an editor"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}
	return m
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	m = newModel(ctx)
	m.styles = &styles{
		cursorStyle: ctx.NewStyle().Foreground(lipgloss.Color("212")),
		cursorLineStyle: ctx.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230")),
		placeholderStyle: ctx.NewStyle().
			Foreground(lipgloss.Color("238")),
		endOfBufferStyle: ctx.NewStyle().
			Foreground(lipgloss.Color("235")),
		focusedPlaceholderStyle: ctx.NewStyle().
			Foreground(lipgloss.Color("99")),
		focusedBorderStyle: ctx.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")),
		blurredBorderStyle: ctx.NewStyle().
			Border(lipgloss.HiddenBorder()),
	}
	for i := 0; i < initialInputs; i++ {
		m.inputs[i] = m.newTextarea(ctx)
	}
	m.inputs[m.focus].Focus()
	m.updateKeybindings()

	return m, textarea.Blink
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			for i := range m.inputs {
				m.inputs[i].Blur()
			}
			return m, tea.Quit
		case key.Matches(msg, m.keymap.next):
			m.inputs[m.focus].Blur()
			m.focus++
			if m.focus > len(m.inputs)-1 {
				m.focus = 0
			}
			cmd := m.inputs[m.focus].Focus()
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keymap.prev):
			m.inputs[m.focus].Blur()
			m.focus--
			if m.focus < 0 {
				m.focus = len(m.inputs) - 1
			}
			cmd := m.inputs[m.focus].Focus()
			cmds = append(cmds, cmd)
		case key.Matches(msg, m.keymap.add):
			m.inputs = append(m.inputs, m.newTextarea(ctx))
		case key.Matches(msg, m.keymap.remove):
			m.inputs = m.inputs[:len(m.inputs)-1]
			if m.focus > len(m.inputs)-1 {
				m.focus = len(m.inputs) - 1
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	m.updateKeybindings()
	m.sizeInputs()

	// Update all textareas
	for i := range m.inputs {
		newModel, cmd := m.inputs[i].Update(ctx, msg)
		m.inputs[i] = newModel
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	for i := range m.inputs {
		m.inputs[i].SetWidth(m.width / len(m.inputs))
		m.inputs[i].SetHeight(m.height - helpHeight)
	}
}

func (m *model) updateKeybindings() {
	m.keymap.add.SetEnabled(len(m.inputs) < maxInputs)
	m.keymap.remove.SetEnabled(len(m.inputs) > minInputs)
}

func (m model) View(ctx tea.Context) string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.next,
		m.keymap.prev,
		m.keymap.add,
		m.keymap.remove,
		m.keymap.quit,
	})

	var views []string
	for i := range m.inputs {
		views = append(views, m.inputs[i].View(ctx))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help
}

func main() {
	if _, err := tea.NewProgram(model{}, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
