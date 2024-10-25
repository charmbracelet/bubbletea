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
	return m.s.block.Render(m.tree.View())
}

type styles struct {
	base,
	block,
	node,
	selected,
	enumerator lipgloss.Style
}

func defaultStyles() styles {
	var s styles
	s.base = lipgloss.NewStyle().
		Background(lipgloss.Color("205"))
	s.block = s.base.
		Padding(1, 3).
		Margin(1, 3)
	s.node = s.base.Foreground(lipgloss.Color("0"))
	s.selected = s.base.Foreground(lipgloss.Color("54")).Bold(true).Underline(true)

	s.enumerator = s.base.
		Foreground(lipgloss.Color("126")).
		PaddingRight(1)
	return s
}

const (
	width  = 35
	height = 15
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
		TreeStyle:          s.block,
		CursorStyle:        s.base.PaddingRight(1),
		NodeStyle:          s.node,
		RootNodeStyle:      s.node,
		ParentNodeStyle:    s.node,
		SelectedNodeStyle:  s.selected,
		EnumeratorStyle:    s.enumerator,
		OpenIndicatorStyle: s.base,
	})

	if _, err := tea.NewProgram(model{tree: t, s: s}).Run(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
