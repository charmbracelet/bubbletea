package main

// This is a simple example illustrating how to enable enhanced keyboard
// support.

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type styles struct {
	ui lipgloss.Style
}

type model struct {
	supportsRelease        bool
	supportsDisambiguation bool
	styles                 styles
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		// Attempt to enable keyboard enhancements. By default, this just
		// enables key disabiguation. For key releases, you'll need to opt-in
		// to that feature.
		tea.RequestKeyReleases,

		tea.RequestBackgroundColor,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// When tea.RequestKeyboardEnhancements is called, the program will receive
	// a tea.KeyboardEnhancementsMsg message. This means that an attempt to
	// enable keyboard enhancements was made, however it doesn't guarantee that
	// it was successful.
	case tea.KeyboardEnhancementsMsg:
		// Check which features were able to be enabled.
		m.supportsRelease = msg.SupportsKeyReleases()
		m.supportsDisambiguation = msg.SupportsKeyDisambiguation()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		default:
			return m, tea.Println("  press: " + msg.String())
		}

	case tea.KeyReleaseMsg:
		return m, tea.Printf("release: %s", msg.String())

	case tea.BackgroundColorMsg:
		// Initialize styles.
		m.updateStyles(msg.IsDark())
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Terminal supports key releases: %v\n", m.supportsRelease)
	fmt.Fprintf(&b, "Terminal supports key disambiguation: %v\n", m.supportsDisambiguation)
	fmt.Fprint(&b, "This demo logs key events. Press ctrl+c to quit.")
	return m.styles.ui.Render(b.String())
}

func (m *model) updateStyles(isDark bool) {
	// Initialize styles.
	lightDark := lipgloss.LightDark(isDark)
	grey := lightDark(lipgloss.Color("239"), lipgloss.Color("245"))
	darkGray := lightDark(lipgloss.Color("245"), lipgloss.Color("239"))

	m.styles.ui = lipgloss.NewStyle().
		Foreground(grey).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(darkGray)
}

func initialModel() model {
	m := model{}
	m.updateStyles(true) // default to dark styles.
	return m
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Urgh: %v\n", err)
		os.Exit(1)
	}
}
