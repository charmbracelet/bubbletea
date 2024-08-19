package main

import (
	"image/color"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	bg          color.Color
	fg          color.Color
	cursor      color.Color
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.BackgroundColor,
		tea.ForegroundColor,
		tea.CursorColor,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			var cmds []tea.Cmd
			if m.fg != nil {
				cmds = append(cmds, tea.SetForegroundColor(m.fg))
			}
			if m.bg != nil {
				cmds = append(cmds, tea.SetBackgroundColor(m.bg))
			}
			if m.cursor != nil {
				cmds = append(cmds, tea.SetCursorColor(m.cursor))
			}
			cmds = append(cmds, tea.Quit)
			return m, tea.Batch(cmds...)
		}

		switch m.state {
		case chooseState:
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
			switch msg.String() {
			case "esc":
				m.choice = 0
				m.choiceIndex = 0
				m.state = chooseState
				m.err = nil
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

					switch choice {
					case foreground:
						return m, tea.SetForegroundColor(col)
					case background:
						return m, tea.SetBackgroundColor(col)
					case cursor:
						return m, tea.SetCursorColor(col)
					}
				}

			default:
				var cmd tea.Cmd
				m.ti, cmd = m.ti.Update(msg)
				return m, cmd
			}
		}

	case tea.BackgroundColorMsg:
		m.bg = msg.Color

	case tea.ForegroundColorMsg:
		m.fg = msg.Color

	case tea.CursorColorMsg:
		m.cursor = msg.Color

	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	switch m.state {
	case chooseState:
		s.WriteString("Choose a color to set:\n\n")
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

	return s.String()
}

func main() {
	ti := textinput.New()
	ti.Placeholder = "#ff00ff"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	p := tea.NewProgram(model{
		ti: ti,
	})

	_, err := p.Run()
	if err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
