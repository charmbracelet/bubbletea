package main

import (
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/tree"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	root := tree.Root("🛂 Passport expiration date")
	thisYear := time.Now().Year()
	for year := thisYear; year < thisYear+10; year++ {
		yRoot := tree.Root(fmt.Sprintf("%d", year)).Close()
		for month := 1; month <= 12; month++ {
			mRoot := tree.Root(time.Month(month).String()).Close().RootStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("1")))
			for day := 1; day < daysIn(time.Month(month), year); day++ {
				mRoot.Child(fmt.Sprintf("%d", day))
			}
			yRoot.Child(mRoot)
		}
		root.Child(yRoot)
	}

	t := tree.New(root, 80, 30)

	if _, err := tea.NewProgram(model{tree: t}).Run(); err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
