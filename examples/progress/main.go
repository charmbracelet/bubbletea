package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	fps              = 60
	stepSize float64 = 1.0 / (float64(fps) * 2.0)
	padding          = 2
	maxWidth         = 80
)

func main() {
	prog, err := progress.NewModel(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
	if err != nil {
		fmt.Println("Could not initialize progress model:", err)
		os.Exit(1)
	}

	if err = tea.NewProgram(example{progress: prog}).Start(); err != nil {
		fmt.Println("Oh no!", err)
		os.Exit(1)
	}
}

type tickMsg time.Time

type example struct {
	percent  float64
	progress *progress.Model
}

func (e example) Init() tea.Cmd {
	return tickCmd()
}

func (e example) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return e, tea.Quit
		default:
			return e, nil
		}

	case tea.WindowSizeMsg:
		e.progress.Width = msg.Width - padding*2 - 4
		if e.progress.Width > maxWidth {
			e.progress.Width = maxWidth
		}
		return e, nil

	case tickMsg:
		e.percent += stepSize
		if e.percent > 1.0 {
			e.percent = 1.0
			return e, tea.Quit
		}
		return e, tickCmd()

	default:
		return e, nil
	}
}

func (e example) View() string {
	pad := strings.Repeat(" ", padding)
	return "\n" + pad + e.progress.View(e.percent) + "\n\n"
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/fps, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
