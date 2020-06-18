package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	viewportTopMargin    = 2
	viewportBottomMargin = 2
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

type model struct {
	content  string
	ready    bool
	viewport viewport.Model
}

func initialize(content string) func() (tea.Model, tea.Cmd) {
	return func() (tea.Model, tea.Cmd) {
		return model{
			content: content, // keep content in the model
		}, nil
	}
}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(model)

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		viewportVerticalMargins := viewportTopMargin + viewportBottomMargin

		if !m.ready {
			m.viewport = viewport.NewModel(msg.Width, msg.Height-viewportVerticalMargins)
			m.viewport.YPosition = viewportTopMargin
			m.viewport.HighPerformanceRendering = true
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - viewportBottomMargin
		}

		// Render (or re-render) the whole viewport
		cmds = append(cmds, viewport.Sync(m.viewport))
	}

	// Because we're using the viewport's default update function (with pager-
	// style navigation) it's important that the viewport's update function:
	//
	// * Recieves messages from the Bubble Tea runtime
	// * Returns commands to the Bubble Tea runtime
	m.viewport, cmd = viewport.Update(msg, m.viewport)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func view(mdl tea.Model) string {
	m, ok := mdl.(model)
	if !ok {
		return "\n  Error: could not perform assertion on model in view."
	}

	if !m.ready {
		return "\n  Initalizing..."
	}

	return fmt.Sprintf(
		"── Mr. Pager ──\n\n%s\n\n── %3.f%% ──",
		viewport.View(m.viewport),
		m.viewport.ScrollPercent()*100,
	)
}
