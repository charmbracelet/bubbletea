package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/v2/tree"
	tea "github.com/charmbracelet/bubbletea/v2"
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
	return m, cmd
}

func (m model) View() string {
	return m.tree.View()
}

func main() {
	t := tree.New(tree.Root("~/charm").
		Child(
			"ayman",
			tree.Root("bash").
				Child(
					tree.Root("tools").
						Child("zsh",
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
		), 70, 13)

	if _, err := tea.NewProgram(model{tree: t}).Run(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
