package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/tree"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ltree "github.com/charmbracelet/lipgloss/tree"
)

type model struct {
	tree tree.Model
	s    styles
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
	return m, cmd
}

func (m model) View() string {
	return m.s.block.Margin(1, 3).Render(m.tree.View())
}

type styles struct {
	base,
	block,
	node,
	selected,
	cursor,
	openCharacter,
	indenter,
	enumerator lipgloss.Style
}

func defaultStyles() styles {
	var s styles
	s.base = lipgloss.NewStyle().
		Background(lipgloss.Color("205"))
	s.block = s.base.
		PaddingTop(1).PaddingRight(3)
	s.cursor = s.base.Padding(1, 1, 0, 3).Foreground(lipgloss.Color("54"))
	s.openCharacter = s.base.Foreground(lipgloss.Color("54"))
	s.node = s.base.Foreground(lipgloss.Color("0"))
	s.selected = s.base.Foreground(lipgloss.Color("54")).Bold(true)

	s.enumerator = s.base.
		Foreground(lipgloss.Color("126"))

	s.indenter = s.base.
		Foreground(lipgloss.Color("126"))
	return s
}

const (
	width  = 45
	height = 13
)

func main() {
	s := defaultStyles()
	t := tree.New(tree.Root("~/charm").
		Enumerator(ltree.RoundedEnumerator).
		Child(
			"ayman",
			tree.Root("bash").
				Child(
					tree.Root("tools").
						Child(
							"zsh",
							"doom-emacs",
						),
				),
			tree.Root("carlos").
				Child(
					tree.Root("emotes").
						Child(
							"chefkiss.png",
							"kekw.png",
						),
				),
			"maas",
		), width, height)
	t.SetShowHelp(false)
	t.SetStyles(tree.Styles{
		TreeStyle:          s.block.Width(width - 5),
		CursorStyle:        s.cursor,
		NodeStyle:          s.node,
		RootNodeStyle:      s.node,
		ParentNodeStyle:    s.node,
		SelectedNodeStyle:  s.selected,
		EnumeratorStyle:    s.enumerator,
		IndenterStyle:      s.indenter,
		OpenIndicatorStyle: s.openCharacter,
	})

	if _, err := tea.NewProgram(model{tree: t, s: s}).Run(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
