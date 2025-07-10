package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type execCmd struct {
	w io.Writer
}

type execDoneMsg struct{}

func (e *execCmd) Run() error {
	for i := range 8 {
		e.w.Write([]byte(fmt.Sprintf("exec %v\n", i)))
	}

	return nil
}

func (e *execCmd) SetStdin(r io.Reader) {
}

func (e *execCmd) SetStdout(w io.Writer) {
	e.w = w
}

func (e *execCmd) SetStderr(w io.Writer) {
}

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case execDoneMsg:
		cmds = append(cmds, tea.Quit)
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var sb strings.Builder

	for i := range 5 {
		sb.WriteString(fmt.Sprintf("view %v\n", i))
	}

	return sb.String()
}

func main() {
	initialModel := model{}

	p := tea.NewProgram(initialModel)

	go func() {
		p.Send(
			tea.Exec(&execCmd{}, func(err error) tea.Msg {
				return execDoneMsg{}
			})(),
		)
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
