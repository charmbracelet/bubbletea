package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type model struct{}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		return m, tea.Printf("You pressed a key! %T", msg)
	case tea.MouseMsg:
		return m, tea.Printf("You clicked a mouse button! %T", msg)
	}
	return m, nil
}

// View implements tea.Model.
func (m model) View() string {
	return "Hi"
}

var _ tea.Model = model{}

func main() {
	rootCmd := &cobra.Command{
		Use:   "game",
		Short: "Tic-Tac-Toe game",
		Long:  "A simple Tic-Tac-Toe game written in Go using the Bubble Tea library and the Lip Gloss library.",
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(model{})
			p.EnableMouseAllMotion()
			p.SetWindowTitle("Welcome to the game")
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
