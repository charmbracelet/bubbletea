package main

// An example demonstrating setting a terminal window title and restoring it on
// exit via the terminal title stack.

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	initialTitle = "Initial Window Title"
	appTitle     = "Bubble Tea Window Title"
)

type model struct {
	count int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "space", "enter":
			m.count++
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	title := appTitle
	if m.count > 0 {
		title = fmt.Sprintf("%s (%d)", appTitle, m.count)
	}
	wrap := lipgloss.NewStyle().Width(78).Render
	content := fmt.Sprintf(
		"Current window title: %q\n\n"+
			"The initial title was saved onto the terminal's title stack.\n"+
			"On exit, the initial title is restored.\n\n"+
			"Press space or enter to update the title\n"+
			"Press q or esc to quit",
		title,
	)
	v := tea.NewView(wrap(content))
	v.WindowTitle = title
	return v
}

func main() {
	fmt.Printf("Setting initial title to: %q\n", initialTitle)
	fmt.Print(ansi.SetWindowTitle(initialTitle))
	fmt.Println("Starting tea app to update title and push initial title onto the stack in 5 seconds...")
	time.Sleep(5 * time.Second)
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
	fmt.Printf("App exited and popped title back to: %q\n", initialTitle)
	fmt.Println("Example exiting in 5 seconds, prompt will probably set the title back to $USER@$HOST $")
	time.Sleep(5 * time.Second)
}
