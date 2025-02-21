package main

import (
	"log"

	_ "image/jpeg"
	_ "image/png"

	bimage "github.com/charmbracelet/bubbles/v2/image"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type model struct {
	m    bimage.Model
	w, h int
}

var _ tea.Model = model{}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.m.Init(),
	)
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	im, cmd := m.m.Update(msg)
	m.m = im.(bimage.Model)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

var slashes = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))

// View implements tea.Model.
func (m model) View() string {
	s := lipgloss.Place(
		m.w,
		m.h,
		lipgloss.Center,
		lipgloss.Center,
		m.m.View(),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceStyle(slashes),
	)
	return s
}

func main() {
	area := bimage.Rect(0, 0, 12, 5)
	// f, err := os.Open("/Users/ayman/Downloads/57376114.jpeg")
	// if err != nil {
	// 	log.Fatalf("could not load image: %v", err)
	// }
	//
	// defer f.Close() //nolint:errcheck

	// img, _, err := image.Decode(f)
	// if err != nil {
	// 	log.Fatalf("could not load image: %v", err)
	// }

	// im := bimage.New(img, area)

	im, err := bimage.NewLocal("/Users/ayman/Downloads/57376114.jpeg", area)
	if err != nil {
		log.Fatalf("could not load image: %v", err)
	}

	m := model{m: im}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("error running program: %v", err)
	}
}
