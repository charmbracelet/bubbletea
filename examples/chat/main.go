package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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
	viewport    viewport.Model
	messages    []string
	textInput   textinput.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Send a message..."
	ti.Focus()
	ti.Prompt = "┃ "
	ti.CharLimit = 280
	ti.Width = 30
	ti.Height = 3
	ti.LineLimit = 10

	vp := viewport.New(30, 10)
	vp.SetContent(`Welcome to the Bubbles multi-line text input!
Try typing any message and pressing ENTER.
If you write a long message, it will automatically wrap :D
	`)

	return model{
		textInput:   ti,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textInput.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textInput.Value())
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textInput.Reset()
			m.viewport.GotoBottom()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textInput.View(),
	) + "\n\n"
}
