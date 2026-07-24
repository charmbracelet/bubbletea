package main

import (
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/stopwatch"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// task pairs a human-readable label with its own stopwatch. Each stopwatch is
// fully independent: it has a unique ID, so ticks are routed only to the
// stopwatch they belong to.
type task struct {
	label     string
	stopwatch stopwatch.Model
}

type keymap struct {
	toggle    key.Binding
	reset     key.Binding
	toggleAll key.Binding
	resetAll  key.Binding
	quit      key.Binding
}

type model struct {
	table  table.Model
	tasks  []task
	keymap keymap
	help   help.Model
}

func newModel() model {
	labels := []string{
		"Download assets",
		"Compile sources",
		"Run tests",
		"Build image",
		"Deploy",
	}

	tasks := make([]task, len(labels))
	for i, label := range labels {
		tasks[i] = task{
			label:     label,
			stopwatch: stopwatch.New(stopwatch.WithInterval(time.Millisecond)),
		}
	}

	columns := []table.Column{
		{Title: "Task", Width: 18},
		{Title: "Status", Width: 8},
		{Title: "Elapsed", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(len(tasks)),
		table.WithWidth(40),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return model{
		table: t,
		tasks: tasks,
		keymap: keymap{
			toggle: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start/stop"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			toggleAll: key.NewBinding(
				key.WithKeys("S"),
				key.WithHelp("S", "start/stop all"),
			),
			resetAll: key.NewBinding(
				key.WithKeys("R"),
				key.WithHelp("R", "reset all"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "q"),
				key.WithHelp("q", "quit"),
			),
		},
		help: help.New(),
	}
}

func (m model) Init() tea.Cmd {
	// Start every stopwatch so the table ticks live right away.
	cmds := make([]tea.Cmd, len(m.tasks))
	for i := range m.tasks {
		cmds[i] = m.tasks[i].stopwatch.Init()
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.toggle):
			cmds = append(cmds, m.tasks[m.table.Cursor()].stopwatch.Toggle())
		case key.Matches(msg, m.keymap.reset):
			cmds = append(cmds, m.tasks[m.table.Cursor()].stopwatch.Reset())
		case key.Matches(msg, m.keymap.toggleAll):
			for i := range m.tasks {
				cmds = append(cmds, m.tasks[i].stopwatch.Toggle())
			}
		case key.Matches(msg, m.keymap.resetAll):
			for i := range m.tasks {
				cmds = append(cmds, m.tasks[i].stopwatch.Reset())
			}
		}
	case stopwatch.TickMsg, stopwatch.StartStopMsg, stopwatch.ResetMsg:
		// Flow stopwatch messages through every stopwatch. Each one ignores
		// messages that aren't addressed to its ID, so this is safe.
		for i := range m.tasks {
			var cmd tea.Cmd
			m.tasks[i].stopwatch, cmd = m.tasks[i].stopwatch.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Let the table handle navigation (up/down, etc.).
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	m.syncRows()
	return m, tea.Batch(cmds...)
}

// syncRows refreshes the table's rows from the current stopwatch state.
func (m *model) syncRows() {
	rows := make([]table.Row, len(m.tasks))
	for i, t := range m.tasks {
		status := "stopped"
		if t.stopwatch.Running() {
			status = "running"
		}
		rows[i] = table.Row{t.label, status, formatElapsed(t.stopwatch.Elapsed())}
	}
	m.table.SetRows(rows)
}

func formatElapsed(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d.%03d",
		int(d.Minutes()),
		int(d.Seconds())%60,
		d.Milliseconds()%1000,
	)
}

func (m model) View() tea.View {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.toggle,
		m.keymap.reset,
		m.keymap.toggleAll,
		m.keymap.resetAll,
		m.keymap.quit,
	})
	return tea.NewView(baseStyle.Render(m.table.View()) + "\n" + help + "\n")
}

func main() {
	m := newModel()
	m.syncRows()
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
