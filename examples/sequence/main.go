package sequence

// A simple example illustrating how to run a series of commands in order.

import (
	"fmt"
	"os"

	tea "github.com/rprtr258/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd {
	return tea.Sequence(
		tea.Batch(
			tea.Println("A"),
			tea.Println("B"),
			tea.Println("C"),
		),
		tea.Println("Z"),
		tea.Quit,
	)
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg.(type) {
	case tea.MsgKey:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View(r tea.Renderer) {
}

func Main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
