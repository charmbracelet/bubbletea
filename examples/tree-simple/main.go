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

const (
	width = 50
)

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
	enumerator lipgloss.Style
}

func defaultStyles() styles {
	var s styles
	s.base = lipgloss.NewStyle().
		Foreground(lipgloss.Color("225"))
	s.block = s.base.
		Background(lipgloss.Color("205")).
		Padding(1, 3).
		Margin(1, 3).
		Width(width)

	s.enumerator = s.base.
		Background(lipgloss.Color("205")).
		Foreground(lipgloss.Color("126")).
		PaddingRight(1)
	return s
}

func main() {
	s := defaultStyles()
	t := tree.New(tree.Root("~/charm").
		Enumerator(ltree.RoundedEnumerator).
		EnumeratorStyle(s.enumerator).
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
		), 30, 10)
	t.SetShowHelp(false)
	t.SetStyles(tree.Styles{
		TreeStyle:         lipgloss.NewStyle().Background(lipgloss.Color("205")),
		NodeStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("205")),
		SelectedNodeStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("54")).Background(lipgloss.Color("205")).Bold(true).Underline(true),
	})

	if _, err := tea.NewProgram(model{tree: t, s: s}).Run(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
