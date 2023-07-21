package altscreen_toggle

import (
	"fmt"
	"log"

	"github.com/muesli/termenv"
	tea "github.com/rprtr258/bubbletea"
)

var (
	color   = termenv.EnvColorProfile().Color
	keyword = termenv.Style{}.Foreground(color("204")).Background(color("235")).Styled
	help    = termenv.Style{}.Foreground(color("241")).Styled
)

type model struct {
	altscreen bool
	quitting  bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MsgKey:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case " ":
			var cmd tea.Cmd
			if m.altscreen {
				cmd = tea.ExitAltScreen
			} else {
				cmd = tea.EnterAltScreen
			}
			m.altscreen = !m.altscreen
			return m, cmd
		}
	}
	return m, nil
}

func (m model) View(r tea.Renderer) {
	if m.quitting {
		r.Write("Bye!\n")
		return
	}

	const (
		altscreenMode = " altscreen mode "
		inlineMode    = " inline mode "
	)

	var mode string
	if m.altscreen {
		mode = altscreenMode
	} else {
		mode = inlineMode
	}

	r.Write(fmt.Sprintf("\n\n  You're in %s\n\n\n%s", keyword(mode), help("  space: switch modes • q: exit\n")))
}

func Main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		log.Fatal("Error running program: ", err.Error())
	}
}
