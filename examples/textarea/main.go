package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	p := tea.NewProgram(initialModel())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

type tickMsg struct{}
type errMsg error

type model struct {
	textarea textarea.Model
	err      error
}

func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "Once upon a time..."
	ti.Focus()
	ti.CharLimit = 400
	ti.Width = 20
	ti.Height = 3
	ti.LineLimit = 10
	ti.CursorLineStyle = lipgloss.NewStyle().Background(lipgloss.Color("0"))
	ti.ShowLineNumbers = true

	return model{
		textarea: ti,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"Tell me a story.\n\n%s\n\nColumn: %d\n\n\n%s",
		m.textarea.View(),
		m.textarea.Cursor(),
		"(esc to quit)",
	) + "\n\n"
}
