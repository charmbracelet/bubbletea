package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/muesli/reflow/indent"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var (
		daemonMode bool
		showHelp   bool
		opts       []tea.ProgramOption
	)

	flag.BoolVar(&daemonMode, "d", false, "run as a daemon")
	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if daemonMode || !isatty.IsTerminal(os.Stdout.Fd()) {
		// If we're in daemon mode don't render the TUI
		opts = []tea.ProgramOption{tea.WithoutRenderer()}
	} else {
		// If we're in TUI mode, discard log output
		log.SetOutput(ioutil.Discard)
	}

	p := tea.NewProgram(newModel(), opts...)
	if err := p.Start(); err != nil {
		fmt.Println("Error starting Bubble Tea program:", err)
		os.Exit(1)
	}
}

type model struct {
	spinner  spinner.Model
	results  []time.Duration
	quitting bool
}

func newModel() model {
	const showLastResults = 5

	return model{
		spinner: spinner.NewModel(),
		results: make([]time.Duration, showLastResults),
	}
}

func (m model) Init() tea.Cmd {
	log.Println("Starting work...")
	return tea.Batch(
		spinner.Tick,
		runPretendProcess,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case processFinishedMsg:
		d := time.Duration(msg)
		log.Printf("Finished job in %s", d)
		m.results = append(m.results[1:], d)
		return m, runPretendProcess
	default:
		return m, nil
	}
}

func (m model) View() string {
	s := "\n" + m.spinner.View() + " Doing some work...\n\n"

	for _, dur := range m.results {
		if dur == 0 {
			s += ".....................\n"
		} else {
			s += fmt.Sprintf("Job finished in %s\n", dur)
		}
	}

	s += "\nPress any key to exit\n"

	if m.quitting {
		s += "\n"
	}

	return indent.String(s, 1)
}

// processFinishedMsg is send when a pretend process completes.
type processFinishedMsg time.Duration

// pretendProcess simulates a long-running process.
func runPretendProcess() tea.Msg {
	pause := time.Duration(rand.Int63n(899)+100) * time.Millisecond
	time.Sleep(pause)
	return processFinishedMsg(pause)
}
