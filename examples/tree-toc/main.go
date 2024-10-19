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
	w := width - enumeratorWidth*child.Depth()
	if strings.HasPrefix(child.Value(), m.tree.OpenCharacter) {
		w -= lipgloss.Width(m.tree.OpenCharacter)
	} else if strings.HasPrefix(child.Value(), m.tree.ClosedCharacter) {
		w -= lipgloss.Width(m.tree.ClosedCharacter)
	} else {
		w -= lipgloss.Width("    ")
	}

	return w
}

func (m *model) updateStyles() {
	m.tree.SetStyles(tree.Styles{
		TreeStyle: lipgloss.NewStyle().Padding(1),
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
	return m.tree.View()
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
}

func (p page) String() string {
	// TODO: overcompensating the number of dots
	return p.title + strings.Repeat(".", width-len(p.title))
}

func main() {
	t := tree.New(
		tree.Root("Go Mistakes").
			Enumerator(enumerator).
			Indenter(indenter).
			Child(
				tree.Root(page{"Code and Project Organization"}).
					Child(page{"Unintended variable shadowing"}).
					Child(page{"Unnecessary nested code"}),
			).
			Child(
				tree.Root(page{"Data Types"}).
					Child(page{"Creating confusion with octal literals"}).
					Child(page{"Neglecting integer overflows"}),
			).
			Child(
				tree.Root(page{"Strings"}).
					Child(page{"Not understaing the concept of rune"}).
					Child(page{"Misusing trim functions"}),
			),
		width,
		height,
	)
	t.OpenCharacter = "📖"
	t.ClosedCharacter = "📘"

	if _, err := tea.NewProgram(model{tree: t}).Run(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
