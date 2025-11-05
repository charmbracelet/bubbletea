package main

// This is a simple example illustrating how to enable enhanced keyboard
// support.

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type styles struct {
	ui lipgloss.Style
}

type model struct {
	supportsDisambiguation bool
	supportsEventTypes     bool
	styles                 styles
}

func (m model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Bubble Tea will send a [tea.KeyboardEnhancementsMsg] on startup if the
	// terminal supports keyboard enhancements features.
	//
	// These features extend the capabilities of keyboard input beyond the basic legacy
	// support found in most terminals. This includes features like:
	//  - Key disambiguation: Improved ability to distinguish between certain key presses
	//     like "enter" and "shift+enter" or "tab" and "ctrl+i".
	//  - Key event types: The ability to report different types of key events such as
	//   key presses and key releases.
	//
	// This allows for more nuanced input handling in terminal applications.
	// You can ask Bubble Tea to request additional keyboard enhancements
	// features by setting fields on the [tea.View.KeyboardEnhancements] struct
	// in your [tea.View] method.
	case tea.KeyboardEnhancementsMsg:
		// Check which features were able to be enabled.
		m.supportsDisambiguation = true // This is always enabled when this msg is received.
		m.supportsEventTypes = msg.SupportsEventTypes()

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

func (m model) View() tea.View {
	var v tea.View
	var b strings.Builder
	fmt.Fprintf(&b, "Terminal supports key releases: %v\n", m.supportsEventTypes)
	fmt.Fprintf(&b, "Terminal supports key disambiguation: %v\n", m.supportsDisambiguation)
	fmt.Fprint(&b, "This demo logs key events. Press ctrl+c to quit.")
	v.SetContent(b.String() + "\n")

	// Attempt to enable reporting key event types (key presses and key
	// releases). By default, only key disambiguation is enabled which improves
	// the ability to distinguish between certain key presses like "enter" and
	// "shift+enter" or "tab" and "ctrl+i".
	v.KeyboardEnhancements.ReportEventTypes = true

	return v
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
