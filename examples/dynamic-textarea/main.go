package main

import (
	"fmt"
	"log"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type model struct {
	textarea textarea.Model
}

func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "Schnrr..."
	ti.ShowLineNumbers = true
	ti.DynamicHeight = true
	ti.MinHeight = 3
	ti.MaxHeight = 15
	ti.MaxContentHeight = 20
	ti.SetWidth(60)
	ti.SetVirtualCursor(false)
	ti.Focus()

	return model{textarea: ti}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.RequestBackgroundColor)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.textarea.SetStyles(textarea.DefaultStyles(msg.IsDark()))
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) statusView() string {
	return fmt.Sprintf(
		"\nHeight: %d · Lines: %d · Cursor: (%d, %d) · Scroll: %.0f%%",
		m.textarea.Height(),
		m.textarea.LineCount(),
		m.textarea.Line(),
		m.textarea.Column(),
		m.textarea.ScrollPercent()*100,
	)
}

func (m model) View() tea.View {
	const gap = 1

	var c *tea.Cursor
	if !m.textarea.VirtualCursor() {
		c = m.textarea.Cursor()
		c.Y += gap
	}

	f := strings.Repeat("\n", gap)
	f += strings.Join([]string{
		m.textarea.View(),
		m.statusView(),
		"\n(ctrl+c to quit)",
	}, "\n")

	v := tea.NewView(f)
	v.Cursor = c
	return v
}
