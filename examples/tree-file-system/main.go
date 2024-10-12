package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/tree"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ltree "charm.land/lipgloss/v2/tree"
	"github.com/charmbracelet/x/ansi"
)

type model struct {
	tree   tree.Model
	choice *tree.Node
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "e":
			m.choice = m.tree.NodeAtCurrentOffset()
			return m, tea.Quit
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.tree, cmd = m.tree.Update(msg)
	m.updateStyles()

	return m, cmd
}

func (m *model) updateStyles() {
	dimmed := lipgloss.Color("239")
	base := lipgloss.NewStyle()
	m.tree.SetStyles(tree.Styles{
		TreeStyle: base.
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("236")).
			BorderBackground(base.GetBackground()),
		RootNodeStyle:      base,
		NodeStyle:          base,
		ParentNodeStyle:    base,
		OpenIndicatorStyle: base,
		SelectedNodeStyle:  base.Bold(true).Background(lipgloss.Color("8")),
		HelpStyle:          base.MarginTop(1),
		EnumeratorStyle:    base.Foreground(dimmed),
		IndenterStyle:      base.Foreground(dimmed),
	})
}

func (m model) View() string {
	return m.tree.View()
}

type file struct {
	name  string
	color string
}

func (f file) String() string {
	return "⌯ " + lipgloss.NewStyle().Foreground(lipgloss.Color(f.color)).Render(f.name)
}

type dir struct {
	name string
}

func (d dir) String() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(d.name)
}

const (
	width           = 50
	height          = 21
	enumeratorWidth = 3
)

func main() {
	t := tree.New(
		tree.Root(dir{"charmbracelet/lipgloss"}).
			Indenter(func(_ ltree.Children, _ int) string {
				return "│ "
			}).
			Enumerator(func(_ ltree.Children, _ int) string {
				return "│ "
			}).
			Child(
				tree.Root(dir{"tree"}).
					Child(file{"tree.go", "6"}).
					Child(file{"renderer.go", "6"}),
			).
			Child(
				tree.Root(dir{"table"}).
					Child(
						tree.Root(dir{"utils"}).
							Child(file{"utils.go", "6"}),
					),
			).
			Child(tree.Root(dir{"list"}).Child(lipgloss.NewStyle().Faint(true).Render("(empty)"))).
			Child(file{"README.md", "3"}).
			Child(file{"go.mod", "255"}).
			Child(file{"go.sum", "255"}).
			Child(file{".gitignore", "255"}),
		width,
		height,
	)
	t.SetCursorCharacter("")
	t.SetOpenCharacter("📂")
	t.SetClosedCharacter("📁")
	kb := []key.Binding{
		key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "select")),
	}
	t.AdditionalShortHelpKeys = func() []key.Binding {
		return kb
	}
	t.AdditionalFullHelpKeys = func() []key.Binding {
		return kb
	}

	p := tea.NewProgram(model{tree: t})
	m, err := p.Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}

	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := m.(model); ok && m.choice != nil {
		fmt.Printf("---\nYou chose %s!\n", ansi.Strip(m.choice.Value()))
	}
}
