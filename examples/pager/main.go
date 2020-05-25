package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/charmbracelet/boba"
)

func main() {

	// Load some text to render
	content, err := ioutil.ReadFile("artichoke.md")
	if err != nil {
		fmt.Println("could not load file:", err)
		os.Exit(1)
	}

	// Use the full size of the terminal in its "Alternate Screen Buffer"
	boba.AltScreen()
	defer boba.ExitAltScreen()

	if err := boba.NewProgram(
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

type model struct {
	err     error
	content string
	ready   bool
	pager   pager.Model
}

func initialize(content string) func() (boba.Model, boba.Cmd) {
	return func() (boba.Model, boba.Cmd) {
		return model{
			content: content,
		}, getTerminalSize()
	}
}

func update(msg boba.Msg, mdl boba.Model) (boba.Model, boba.Cmd) {
	m, _ := mdl.(model)

	switch msg := msg.(type) {
	case boba.KeyMsg:
		if msg.Type == boba.KeyCtrlC {
			return m, boba.Quit
		}
		m.pager, _ = pager.Update(msg, m.pager)
	case terminalSizeMsg:
		if msg.Error() != nil {
			m.err = msg.Error()
			break
		}
		w, h := msg.Size()
		m.pager = pager.NewModel(w, h)
		m.pager.SetContent(m.content)
		m.ready = true
	}

	return m, nil
}

func view(mdl boba.Model) string {
	m, _ := mdl.(model)
	if m.err != nil {
		return "\nError:" + m.err.Error()
	} else if m.ready {
		return "\n" + pager.View(m.pager)
	}
	return "\nInitalizing..."
}

func getTerminalSize() boba.Cmd {
	return boba.GetTerminalSize(func(w, h int, err error) boba.TerminalSizeMsg {
		return terminalSizeMsg{width: w, height: h, err: err}
	})
}
