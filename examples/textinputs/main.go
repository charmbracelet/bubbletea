package main

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	styles     *styles
}

type styles struct {
	focusedStyle        lipgloss.Style
	blurredStyle        lipgloss.Style
	cursorStyle         lipgloss.Style
	noStyle             lipgloss.Style
	helpStyle           lipgloss.Style
	cursorModeHelpStyle lipgloss.Style

	focusedButton string
	blurredButton string
}

func initialModel(ctx tea.Context) model {
	m := model{
		inputs: make([]textinput.Model, 3),
	}

	m.styles = &styles{
		focusedStyle:        ctx.NewStyle().Foreground(lipgloss.Color("205")),
		blurredStyle:        ctx.NewStyle().Foreground(lipgloss.Color("240")),
		cursorStyle:         ctx.NewStyle().Foreground(lipgloss.Color("205")),
		noStyle:             ctx.NewStyle(),
		helpStyle:           ctx.NewStyle().Foreground(lipgloss.Color("240")),
		cursorModeHelpStyle: ctx.NewStyle().Foreground(lipgloss.Color("244")),
	}

	m.styles.focusedButton = m.styles.focusedStyle.Copy().Render("[ Submit ]")
	m.styles.blurredButton = fmt.Sprintf("[ %s ]", m.styles.blurredStyle.Render("Submit"))

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New(ctx)
		t.Cursor.Style = m.styles.cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Nickname"
			t.Focus()
			t.PromptStyle = m.styles.focusedStyle
			t.TextStyle = m.styles.focusedStyle
		case 1:
			t.Placeholder = "Email"
			t.CharLimit = 64
		case 2:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) Init(ctx tea.Context) (tea.Model, tea.Cmd) {
	m = initialModel(ctx)
	return m, textinput.Blink
}

func (m model) Update(ctx tea.Context, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = m.styles.focusedStyle
					m.inputs[i].TextStyle = m.styles.focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = m.styles.noStyle
				m.inputs[i].TextStyle = m.styles.noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(ctx, msg)

	return m, cmd
}

func (m *model) updateInputs(ctx tea.Context, msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(ctx, msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View(ctx tea.Context) string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View(ctx))
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &m.styles.blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &m.styles.focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(m.styles.helpStyle.Render("cursor mode is "))
	b.WriteString(m.styles.cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(m.styles.helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
