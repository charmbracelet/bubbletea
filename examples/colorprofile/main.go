package main

import (
	"image/color"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

var myFancyColor color.Color

type model struct{}

var _ tea.Model = model{}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestCapability("RGB"),
		tea.RequestCapability("Tc"),
	)
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.ColorProfileMsg:
		return m, tea.Println("Color profile manually set to ", msg)
	}
	return m, nil
}

// View implements tea.Model.
func (m model) View() tea.View {
	return tea.NewView("This will produce the wrong colors on Apple Terminal :)\n\n" +
		ansi.Style{}.ForegroundColor(myFancyColor).Styled("Howdy!") +
		"\n\n" +
		"Press any key to exit.")
}

func main() {
	myFancyColor, _ = colorful.Hex("#6b50ff")

	p := tea.NewProgram(model{}, tea.WithColorProfile(colorprofile.TrueColor))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
