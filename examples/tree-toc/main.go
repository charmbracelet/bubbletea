package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/tree"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ltree "charm.land/lipgloss/v2/tree"
)

type model struct {
	tree tree.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.tree, cmd = m.tree.Update(msg)
	m.updateStyles()

	return m, cmd
}

func (m *model) updateStyles() {
	m.tree.SetStyles(tree.Styles{
		RootNodeStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		SelectedNodeStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#EE6FF8")).Bold(true),
		CursorStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("#EE6FF8")).Bold(true),
		HelpStyle:         lipgloss.NewStyle().MarginTop(1),
	})
}

func (m model) View() string {
	pageNumbers := make([]string, len(m.tree.AllNodes()))
	for i, node := range m.tree.AllNodes() {
		v := node.GivenValue()
		if page, ok := v.(page); ok {
			num := fmt.Sprintf("%d", page.page)
			if i == m.tree.YOffset() {
				num = lipgloss.NewStyle().Foreground(lipgloss.Color("#EE6FF8")).Bold(true).Render(num)
			} else {
				num = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(num)
			}
			pageNumbers[i] = num
		}
	}
	return lipgloss.NewStyle().Padding(1).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.tree.View(),
			lipgloss.JoinVertical(lipgloss.Left, pageNumbers...)),
	)
}

const (
	width           = 60
	height          = 12
	enumeratorWidth = 3
)

func enumerator(_ ltree.Children, _ int) string {
	return "  "
}

func indenter(_ ltree.Children, _ int) string {
	return "    "
}

type page struct {
	title string
	page  int
}

func (p page) String() string {
	return p.title
}

func main() {
	t := tree.New(
		tree.Root(page{"Go Mistakes", 1}).
			Enumerator(enumerator).
			Indenter(indenter).
			Child(
				tree.Root(page{"Code and Project Organization", 2}).
					Child(page{"Unintended variable shadowing", 12}).
					Child(page{"Unnecessary nested code", 22}),
			).
			Child(
				tree.Root(page{"Data Types", 23}).
					Child(page{"Creating confusion with octal literals", 28}).
					Child(page{"Neglecting integer overflows", 52}),
			).
			Child(
				tree.Root(page{"Strings", 53}).
					Child(page{"Not understaing the concept of rune", 59}).
					Child(page{"Misusing trim functions", 61}),
			),
		width,
		height,
	)
	t.SetClosedCharacter("📘")
	t.SetOpenCharacter("📖")

	if _, err := tea.NewProgram(model{tree: t}).Run(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
