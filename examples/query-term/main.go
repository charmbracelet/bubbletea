// This example uses a textinput to send the terminal ANSI sequences to query
// it for capabilities.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

func newModel() model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.SetWidth(20)
	ti.SetVirtualCursor(false)
	return model{input: ti}
}

type model struct {
	input textinput.Model
	err   error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m.err = nil
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			// Write the sequence to the terminal.
			val := m.input.Value()
			val = "\"" + val + "\""

			// Unescape the sequence.
			seq, err := strconv.Unquote(val)
			if err != nil {
				m.err = err
				return m, nil
			}

			if !strings.HasPrefix(seq, "\x1b") {
				m.err = fmt.Errorf("sequence is not an ANSI escape sequence")
				return m, nil
			}

			m.input.SetValue("")

			// Write the sequence to the terminal.
			return m, func() tea.Msg {
				io.WriteString(os.Stdout, seq)
				return nil
			}
		}
	default:
		_, typ, ok := strings.Cut(fmt.Sprintf("%T", msg), ".")
		if ok && unicode.IsUpper(rune(typ[0])) {
			// Only log messages that are exported types.
			cmds = append(cmds, tea.Printf("Received message: %T %+v", msg, msg))
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	var s strings.Builder
	s.WriteString(m.input.View())
	if m.err != nil {
		s.WriteString("\n\nError: " + m.err.Error())
	}
	s.WriteString("\n\nPress ctrl+c to quit, enter to write the sequence to terminal")
	v := tea.NewView(s.String())
	v.Cursor = m.input.Cursor()
	return v
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
