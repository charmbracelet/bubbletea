package main

// An example of copying to and pasting from the system clipboard.
//
// Copying uses OSC52 when the terminal supports it. On terminals that don't
// support OSC52, such as Apple's Terminal.app, Bubble Tea bridges the
// clipboard through the operating system's clipboard tools (pbcopy/pbpaste on
// macOS, wl-clipboard or xclip on Linux) instead.
//
// The same bridge applies to programs started with Exec and ExecProcess: when
// the terminal can't handle OSC52, they run on a pseudo-terminal whose
// clipboard traffic is intercepted and bridged.

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

type model struct {
	input  string
	copied string
	pasted string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			m.copied = m.input
			return m, tea.SetClipboard(m.input)
		case "ctrl+v":
			return m, tea.ReadClipboard
		case "backspace":
			if n := len(m.input); n > 0 {
				m.input = m.input[:n-1]
			}
		default:
			if text := msg.Key().Text; len(text) > 0 {
				m.input += text
			}
		}

	case tea.ClipboardMsg:
		m.pasted = msg.Content
	}

	return m, nil
}

func (m model) View() tea.View {
	var s strings.Builder
	s.WriteString("Type something, then:\n\n")
	s.WriteString(helpStyle.Render("  enter   copy the input to the clipboard") + "\n")
	s.WriteString(helpStyle.Render("  ctrl+v  paste the clipboard") + "\n")
	s.WriteString(helpStyle.Render("  esc     quit") + "\n\n")
	s.WriteString("Input:  " + m.input + "\n")
	s.WriteString("Copied: " + m.copied + "\n")
	s.WriteString("Pasted: " + m.pasted + "\n")
	return tea.NewView(s.String())
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
