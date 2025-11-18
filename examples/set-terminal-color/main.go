package main

import (
	"image/color"
	"log"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

type colorType int

const (
	foreground colorType = iota + 1
	background
	cursor
)

func (c colorType) String() string {
	switch c {
	case foreground:
		return "Foreground"
	case background:
		return "Background"
	case cursor:
		return "Cursor"
	default:
		return "Unknown"
	}
}

type state int

const (
	chooseState state = iota
	inputState
)

type model struct {
	ti          textinput.Model
	choice      colorType
	state       state
	choiceIndex int
	err         error
	fg, bg, cc  color.Color
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		switch m.state {
		case chooseState:
			m.ti.Blur()
			switch msg.String() {
			case "j", "down":
				m.choiceIndex++
				if m.choiceIndex > 2 {
					m.choiceIndex = 0
				}
			case "k", "up":
				m.choiceIndex--
				if m.choiceIndex < 0 {
					m.choiceIndex = 2
				}
			case "enter":
				m.state = inputState
				m.ti.Focus()
				switch m.choiceIndex {
				case 0:
					m.choice = foreground
				case 1:
					m.choice = background
				case 2:
					m.choice = cursor
				}
			}

		case inputState:
			m.ti.Focus()
			switch msg.String() {
			case "esc":
				m.choice = 0
				m.choiceIndex = 0
				m.state = chooseState
				m.err = nil
				m.ti.Blur()
			case "enter":
				val := m.ti.Value()
				col, err := colorful.Hex(val)
				if err != nil {
					m.err = err
				} else {
					m.err = nil
					choice := m.choice
					m.choice = 0
					m.choiceIndex = 0
					m.state = chooseState

					// Reset the text input
					m.ti.Reset()

					switch choice {
					case foreground:
						m.fg = col
					case background:
						m.bg = col
					case cursor:
						m.cc = col
					}
				}

				m.ti.Blur()

			default:
				var cmd tea.Cmd
				m.ti, cmd = m.ti.Update(msg)
				return m, cmd
			}
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	var s strings.Builder
	instructions := lipgloss.NewStyle().Width(40).Render("Choose a terminal-wide color to set. All settings will be cleared on exit.")

	switch m.state {
	case chooseState:
		s.WriteString(instructions + "\n\n")
		for i, c := range []colorType{foreground, background, cursor} {
			if i == m.choiceIndex {
				s.WriteString(" > ")
			} else {
				s.WriteString("   ")
			}
			s.WriteString(c.String())
			s.WriteString("\n")
		}
	case inputState:
		s.WriteString("Enter a color in hex format:\n\n")
		s.WriteString(m.ti.View())
		s.WriteString("\n")
	}

	if m.err != nil {
		s.WriteString("\nError: ")
		s.WriteString(m.err.Error())
	}

	s.WriteString("\nPress q to quit")

	switch m.state {
	case chooseState:
		s.WriteString(", j/k to move, and enter to select")
	case inputState:
		s.WriteString(", and enter to submit, esc to go back")
	}

	s.WriteString("\n")

	v := tea.NewView(s.String())
	if m.ti.Focused() {
		v.Cursor = m.ti.Cursor()
		v.Cursor.Y += 2 // account for the prompt
		v.Cursor.Color = m.cc
	}
	v.BackgroundColor = m.bg
	v.ForegroundColor = m.fg

	return v
}

func main() {
	ti := textinput.New()
	ti.Placeholder = "#ff00ff"
	ti.CharLimit = 156
	ti.SetWidth(20)
	ti.SetVirtualCursor(false)
	p := tea.NewProgram(model{
		ti: ti,
	})

	_, err := p.Run()
	if err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
