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

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func newModel() model {
	ti := textinput.NewModel()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	return model{input: ti}
}

type model struct {
	input textinput.Model
	err   error
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
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
			// Write the sequence to the terminal
			val := m.input.Value()
			val = "\"" + val + "\""

			// unescape the sequence
			seq, err := strconv.Unquote(val)
			if err != nil {
				m.err = err
				return m, nil
			}

			if !strings.HasPrefix(seq, "\x1b") {
				m.err = fmt.Errorf("sequence is not an ANSI escape sequence")
				return m, nil
			}

			// write the sequence to the terminal
			return m, func() tea.Msg {
				io.WriteString(os.Stdout, seq)
				return nil
			}
		}
	default:
		typ := strings.TrimPrefix(fmt.Sprintf("%T", msg), "tea.")
		if len(typ) > 0 && unicode.IsUpper(rune(typ[0])) {
			// Only log messages that are exported types
			cmds = append(cmds, tea.Printf("Received message: %T\n", msg))
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString(m.input.View())
	if m.err != nil {
		s.WriteString("\n\nError: " + m.err.Error())
	}
	s.WriteString("\n\nPress ctrl+c to quit, enter to write the sequence to terminal")
	return s.String()
}

func main() {
	p := tea.NewProgram(newModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
