package main

// An example Bubble Tea server. This will put an ssh session into alt screen
// and continually print up to date terminal information.

import (
	"fmt"

	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

const (
	host = "0.0.0.0"
	port = "1337"
)

func main() {
	m := NewModel()
	tea.NewProgram(m).Run()
}

type LI string

func (i LI) Title() string {
	return string(i)
}

func (i LI) Description() string {
	return string(i)
}

func (i LI) FilterValue() string {
	return string(i)
}

func NewModel() model {
	return model{
		txtStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		quitStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		bg:        "light",
		list: list.New(
			[]list.Item{
				LI("Item 1"),
				LI("Item 2"),
				LI("Item 3"),
				LI("Item 4"),
				LI("Item 5"),
				LI("Item 6"),
				LI("Item 7"),
				LI("Item 8"),
			}, list.NewDefaultDelegate(), 80, 20,
		),
	}
}

// Just a generic tea.Model to demo terminal information of ssh.
type model struct {
	term      string
	profile   string
	width     int
	height    int
	bg        string
	txtStyle  lipgloss.Style
	quitStyle lipgloss.Style
	list      list.Model
}

func (m model) Init() tea.Cmd {
	// default values
	return tea.Batch(
		tea.EnterAltScreen,
		tea.RequestBackgroundColor,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.ColorProfileMsg:
		m.profile = msg.String()
	case tea.BackgroundColorMsg:
		if msg.IsDark() {
			m.bg = "dark"
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.list.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := fmt.Sprintf("  Your term is %s\n  Your window size is %dx%d\n\n\n  Background: %s\n  Color Profile: %s", "", m.width, m.height, m.bg, m.profile)
	s = m.txtStyle.Render(s) + "\n\n" + m.quitStyle.Render("Press 'q' to quit\n")
	s += "\n" + lipgloss.NewStyle().Render(m.list.View())
	return s
}
