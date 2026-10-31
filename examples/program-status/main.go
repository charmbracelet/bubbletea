package main

import (
	"fmt"
	"log"
	"time"

	tea "charm.land/bubbletea/v2"
)

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

var states = []tea.ProgramState{
	tea.ProgramStateIdle,
	tea.ProgramStateWorking,
	tea.ProgramStateBlocked,
	tea.ProgramStateDone,
	tea.ProgramStateError,
}

var messages = map[tea.ProgramState]string{
	tea.ProgramStateIdle:    "Waiting for instructions",
	tea.ProgramStateWorking: "Building the project",
	tea.ProgramStateBlocked: "Deploy to production?",
	tea.ProgramStateDone:    "Build finished",
	tea.ProgramStateError:   "Build failed",
}

type model struct {
	state     int
	progress  int
	supported bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.RequestProgramStatusSupport, tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.ProgramStatusSupportMsg:
		m.supported = true
	case tickMsg:
		if states[m.state] == tea.ProgramStateWorking {
			m.progress = (m.progress + 5) % 105
		}
		return m, tick()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "right", "l", "space":
			m.state = (m.state + 1) % len(states)
			m.progress = 0
		case "left", "h":
			m.state = (m.state + len(states) - 1) % len(states)
			m.progress = 0
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	state := states[m.state]
	support := "no reply yet (unsupported, or behind tmux)"
	if m.supported {
		support = "supported"
	}

	v := tea.NewView(fmt.Sprintf(
		"Program Status Protocol (OSC 7501): %s\n\nState: %s\nMessage: %s\n\nPress left/right to change state, q to quit.\n",
		support, state, messages[state],
	))
	v.ProgramStatus = &tea.ProgramStatus{
		State:       state,
		App:         "program-status-example",
		Kind:        tea.ProgramStatusKindPermission,
		Progress:    m.progress,
		HasProgress: state == tea.ProgramStateWorking,
		Message:     messages[state],
	}
	return v
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
