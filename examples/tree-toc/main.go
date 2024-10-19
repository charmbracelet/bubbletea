package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/tree"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ltree "github.com/charmbracelet/lipgloss/tree"
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

func (m *model) childWidth(child *tree.Node) int {
	w := width - enumeratorWidth*child.Depth() + 1
	if strings.HasPrefix(child.Value(), m.tree.OpenCharacter) {
		w += lipgloss.Width(m.tree.OpenCharacter)
	} else if strings.HasPrefix(child.Value(), m.tree.ClosedCharacter) {
		w += lipgloss.Width(m.tree.ClosedCharacter)
	}
	return w
}

func (m *model) updateStyles() {
	m.tree.SetStyles(tree.Styles{
		NodeStyleFunc: func(children tree.Nodes, i int) lipgloss.Style {
			child := children.At(i)
			w := m.childWidth(child)
			s := lipgloss.NewStyle().Width(w).MaxWidth(w).Inline(true)
			// TODO: should this be defined in a RootStyle?
			if child.Children().Length() > 0 {
				return s.Bold(true)
			}

			return s
		},
		SelectedNodeStyleFunc: func(children tree.Nodes, i int) lipgloss.Style {
			child := children.At(i)
			w := m.childWidth(child)

			return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("92")).Width(w).MaxWidth(w).Inline(true)
		},
		HelpStyle: lipgloss.NewStyle().MarginTop(1),
	})
}

func (m model) View() string {
	pageNumbers := make([]string, len(m.tree.FlatNodes()))
	for i, node := range m.tree.FlatNodes() {
		v := node.GivenValue()
		// check if v is page
		if page, ok := v.(page); ok {
			num := fmt.Sprintf("%d", page.page)
			if i == m.tree.YOffset() {
				num = lipgloss.NewStyle().Foreground(lipgloss.Color("92")).Render(fmt.Sprintf("%d", page.page))
			}
			pageNumbers[i] = num
		} else {
			pageNumbers[i] = fmt.Sprintf("%T", v)
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
	height          = 30
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
	// TODO: overcompensating the number of dots, should I support a ValueFunc?
	return p.title + strings.Repeat(".", width-len(p.title))
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
	t.ClosedCharacter = "📘"
	t.OpenCharacter = "📖"

	if _, err := tea.NewProgram(model{tree: t}).Run(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
