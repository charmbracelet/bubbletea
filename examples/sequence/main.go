package main

// A simple example illustrating how to run a series of commands in order.

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd {
	return tea.Sequence(
		tea.Batch(
			tea.Sequence(
				SleepPrintln("1-1-1", 1000),
				SleepPrintln("1-1-2", 1000),
			),
			tea.Batch(
				SleepPrintln("1-2-1", 1500),
				SleepPrintln("1-2-2", 1250),
			),
		),
		tea.Println("2"),
		tea.Sequence(
			tea.Batch(
				SleepPrintln("3-1-1", 500),
				SleepPrintln("3-1-2", 1000),
			),
			tea.Sequence(
				SleepPrintln("3-2-1", 750),
				SleepPrintln("3-2-2", 500),
			),
		),
		tea.Quit,
	)
}

// print string after stopping for a certain period of time
func SleepPrintln(s string, milisecond int) tea.Cmd {
	printCmd := tea.Println(s)
	return func() tea.Msg {
		time.Sleep(time.Duration(milisecond) * time.Millisecond)
		return printCmd()
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	return ""
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
