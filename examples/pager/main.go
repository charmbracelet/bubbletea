package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

const (
	useHighPerformanceRenderer = true

	headerHeight = 3
	footerHeight = 3
)

func main() {

	// Load some text to render
	content, err := ioutil.ReadFile("artichoke.md")
	//content, err := ioutil.ReadFile("menagerie.txt")
	if err != nil {
		fmt.Println("could not load file:", err)
		os.Exit(1)
	}

	// Set PAGER_LOG to a path to log to a file. For example,
	//
	//     export PAGER_LOG=debug.log
	//
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
		verticalMargins := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.NewModel(msg.Width, msg.Height-verticalMargins)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargins
		}

		// Render (or re-render) the whole viewport
		if useHighPerformanceRenderer {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	// Because we're using the viewport's default update function (with pager-
	// style navigation) it's important that the viewport's update function:
	//
	// * Recieves messages from the Bubble Tea runtime
	// * Returns commands to the Bubble Tea runtime
	m.viewport, cmd = viewport.Update(msg, m.viewport)
	if useHighPerformanceRenderer {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func view(mdl tea.Model) string {
	m, _ := mdl.(model)

	if !m.ready {
		return "\n  Initalizing..."
	}

	headerTop := "╭───────────╮"
	headerMid := "│ Mr. Pager ├"
	headerBot := "╰───────────╯"
	headerMid += strings.Repeat("─", m.viewport.Width-runewidth.StringWidth(headerMid))
	header := fmt.Sprintf("%s\n%s\n%s", headerTop, headerMid, headerBot)

	footerTop := "╭──────╮"
	footerMid := fmt.Sprintf("┤ %3.f%% │", m.viewport.ScrollPercent()*100)
	footerBot := "╰──────╯"
	gapSize := m.viewport.Width - runewidth.StringWidth(footerMid)
	footerTop = strings.Repeat(" ", gapSize) + footerTop
	footerMid = strings.Repeat("─", gapSize) + footerMid
	footerBot = strings.Repeat(" ", gapSize) + footerBot
	footer := footerTop + "\n" + footerMid + "\n" + footerBot

	return fmt.Sprintf("%s\n%s\n%s", header, viewport.View(m.viewport), footer)
}
