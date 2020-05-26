package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	// Load some text to render
	content, err := ioutil.ReadFile("artichoke.md")
	if err != nil {
		fmt.Println("could not load file:", err)
		os.Exit(1)
	}

	// Set PAGER_LOG to a path to log to a file. For example,
	// export PAGER_LOG=debug.log
	if os.Getenv("PAGER_LOG") != "" {
		p := os.Getenv("PAGER_LOG")
		f, err := tea.LogToFile(p, "pager")
		if err != nil {
			fmt.Printf("Could not open file %s: %v", p, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	// Use the full size of the terminal in its "Alternate Screen Buffer"
	tea.AltScreen()
	defer tea.ExitAltScreen()

	if err := tea.NewProgram(
		initialize(string(content)),
		update,
		view,
	).Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

type terminalSizeMsg struct {
	width  int
	height int
	err    error
}

func (t terminalSizeMsg) Size() (int, int) { return t.width, t.height }
func (t terminalSizeMsg) Error() error     { return t.err }

type resizeMsg struct{}

type model struct {
	err      error
	content  string
	ready    bool
	viewport viewport.Model
}

func initialize(content string) func() (tea.Model, tea.Cmd) {
	return func() (tea.Model, tea.Cmd) {
		return model{
				content: content, // keep content in the model
			}, tea.Batch(

				// Get terminal size asynchronously
				getTerminalSize(),

				// We're not doing anything with it in this example, but this
				// is now you'd listen for resizes.
				tea.OnResize(func() tea.Msg { return resizeMsg{} }),
			)
	}
}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		m.viewport, _ = viewport.Update(msg, m.viewport)

	case terminalSizeMsg:
		if msg.Error() != nil {
			m.err = msg.Error()
			break
		}
		w, h := msg.Size()
		if !m.ready {
			m.viewport = viewport.NewModel(w, h)
			m.viewport.SetContent(m.content)
			m.ready = true
		}

	case resizeMsg:
		return m, getTerminalSize()
	}

	return m, nil
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)
	if m.err != nil {
		return "\nError:" + m.err.Error()
	} else if m.ready {
		return "\n" + viewport.View(m.viewport)
	}
	return "\nInitalizing..."
}

func getTerminalSize() tea.Cmd {
	return tea.GetTerminalSize(func(w, h int, err error) tea.TerminalSizeMsg {
		return terminalSizeMsg{width: w, height: h, err: err}
	})
}
