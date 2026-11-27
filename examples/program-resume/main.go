// A counter that the terminal can bring back, count included, after it
// restarts. Build it first: a binary from "go run" is deleted when it exits,
// so the terminal would have nothing to resume.
//
//	go build -o counter . && ./counter
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	count     int
	exe, dir  string
	supported bool
	quitting  bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.EnvMsg:
		m.supported = msg.Getenv("OSC7577_TOKEN") != ""
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k", "+":
			m.count++
		case "down", "j", "-":
			m.count--
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	support := "not supported by this terminal"
	if m.supported {
		support = "supported"
	}

	v := tea.NewView(fmt.Sprintf(
		"Program Resume Protocol (OSC 7577): %s\n\nCount: %d\n\nPress up/down to count. Restart the terminal to bring it back, or q to quit.\n",
		support, m.count,
	))

	// The user quit, so there is nothing to resume: leaving it out clears
	// the record.
	if !m.quitting {
		// The command carries the count, so the resumed program picks up
		// where this one left off.
		v.ProgramResume = &tea.ProgramResume{
			Args: []string{m.exe, "-count", strconv.Itoa(m.count)},
			Dir:  m.dir,
		}
	}
	return v
}

func main() {
	count := flag.Int("count", 0, "the count to start from")
	flag.Parse()

	// An absolute path, so the terminal finds the program without PATH.
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tea.NewProgram(model{count: *count, exe: exe, dir: dir}).Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
