package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/tree"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type model struct {
	tree   tree.Model
	choice *tree.Item
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
			m.choice = m.tree.ItemAtCurrentOffset()
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
	m.tree.SetStyles(tree.Styles{ItemStyleFunc: func(children tree.Items, i int) lipgloss.Style {
		child := children.At(i)
		w := width - enumeratorWidth*child.Depth()
		if strings.HasPrefix(child.Value(), m.tree.OpenCharacter) {
			w -= lipgloss.Width(m.tree.OpenCharacter)
		} else if strings.HasPrefix(child.Value(), m.tree.ClosedCharacter) {
			w -= lipgloss.Width(m.tree.ClosedCharacter)
		} else {
			w -= lipgloss.Width("⌯ ")
		}

		if child.YOffset() == m.tree.YOffset() {
			return lipgloss.NewStyle().Width(w).Background(lipgloss.Color("8"))
		}

		return lipgloss.NewStyle().Width(w).Background(lipgloss.Color("234"))
	}})
}

func (m model) View() string {
	return m.tree.View()
}

type file struct {
	name  string
	color string
}

func (f file) String() string {
	// TODO: can't partially apply the foreground only to the icon
	// This happens because creating the new style somehow resets the background of the selected node
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
	height          = 30
	enumeratorWidth = 3
)

func main() {
	t := tree.New(
		tree.Root(dir{"charmbracelet/lipgloss"}).
			EnumeratorStyle(lipgloss.NewStyle().Background(lipgloss.Color("234"))).
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
			Child(tree.Root(dir{"list"})).
			Child(file{"README.md", "3"}).
			Child(file{"go.mod", "255"}).
			Child(file{"go.sum", "255"}).
			Child(file{".gitignore", "255"}),
		width,
		height,
	)
	t.OpenCharacter = "📂"
	t.ClosedCharacter = "📁"
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
		// TODO: should we expose a OriginalValue() method on Item?
		fmt.Printf("---\nYou chose %s!\n", ansi.Strip(m.choice.Value()))
	}
}
