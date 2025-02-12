package main

import (
	"github.com/charmbracelet/bubbles/v2/table"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func NewTable() table.Model {
	rows := []table.Row{
		{"1", "issue", "v1.2.3", "24/11/22", "EnterAltScreen"},
	}
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "NAME", Width: 28},
		{Title: "VERSION", Width: 18},
		{Title: "DATE", Width: 15},
		{Title: "REMARK", Width: 4},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
}

type Model struct {
	currentModel   table.Model
	enterAltScreen bool
}

func (m Model) Init() tea.Cmd {
	if m.enterAltScreen {
		return tea.EnterAltScreen
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			{
				if m.enterAltScreen {
					m.enterAltScreen = false
					cmds = append(cmds, tea.ExitAltScreen)
				} else {
					m.enterAltScreen = true
					cmds = append(cmds, tea.EnterAltScreen)
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.currentModel, cmd = m.currentModel.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return baseStyle.Render(m.currentModel.View()) + "\n" + m.currentModel.HelpView() + "\n"
}

func NewModel() (model Model, err error) {
	return Model{
		currentModel:   NewTable(),
		enterAltScreen: true,
	}, err
}

func main() {
	model, err := NewModel()
	if err != nil {
		return
	}
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return
	}
}

func init() {
	_, _ = tea.LogToFile("tea.log", "")
}
