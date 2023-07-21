package help

import (
	"strings"

	tea "github.com/rprtr258/bubbletea"
	"github.com/rprtr258/bubbletea/bubbles/key"
	"github.com/rprtr258/bubbletea/lipgloss"
)

// KeyMap is a map of keybindings used to generate help. Since it's an
// interface it can be any type, though struct or a map[string][]key.Binding
// are likely candidates.
//
// Note that if a key is disabled (via key.Binding.SetEnabled) it will not be
// rendered in the help view, so in theory generated help should self-manage.
type KeyMap interface {
	// ShortHelp returns a slice of bindings to be displayed in the short
	// version of the help. The help bubble will render help in the order in
	// which the help items are returned here.
	ShortHelp() []key.Binding

	// FullHelp returns an extended group of help items, grouped by columns.
	// The help bubble will render the help in the order in which the help
	// items are returned here.
	FullHelp() [][]key.Binding
}

// Styles is a set of available style definitions for the Help bubble.
type Styles struct {
	Ellipsis lipgloss.Style

	// Styling for the short help
	ShortKey       lipgloss.Style
	ShortDesc      lipgloss.Style
	ShortSeparator lipgloss.Style

	// Styling for the full help
	FullKey       lipgloss.Style
	FullDesc      lipgloss.Style
	FullSeparator lipgloss.Style
}

// Model contains the state of the help view.
type Model struct {
	Width   int
	ShowAll bool // if true, render the "full" help menu

	ShortSeparator string
	FullSeparator  string

	// The symbol we use in the short help when help items have been truncated
	// due to width. Periods of ellipsis by default.
	Ellipsis string

	Styles Styles
}

// New creates a new help view with some useful defaults.
func New() Model {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	})

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	})

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#DDDADA",
		Dark:  "#3C3C3C",
	})

	return Model{
		ShortSeparator: " • ",
		FullSeparator:  "    ",
		Ellipsis:       "…",
		Styles: Styles{
			ShortKey:       keyStyle,
			ShortDesc:      descStyle,
			ShortSeparator: sepStyle,
			Ellipsis:       sepStyle.Copy(),
			FullKey:        keyStyle.Copy(),
			FullDesc:       descStyle.Copy(),
			FullSeparator:  sepStyle.Copy(),
		},
	}
}

// NewModel creates a new help view with some useful defaults.
//
// Deprecated: use [New] instead.
var NewModel = New

// Update helps satisfy the Bubble Tea Model interface. It's a no-op.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders the help view's current state.
func (m Model) View(k KeyMap) string {
	if m.ShowAll {
		return m.FullHelpView(k.FullHelp())
	}
	return m.ShortHelpView(k.ShortHelp())
}

// ShortHelpView renders a single line help view from a slice of keybindings.
// If the line is longer than the maximum width it will be gracefully
// truncated, showing only as many help items as possible.
func (m Model) ShortHelpView(bindings []key.Binding) string {
	if len(bindings) == 0 {
		return ""
	}

	var b strings.Builder
	var totalWidth int
	separator := m.Styles.ShortSeparator.Inline(true).Render(m.ShortSeparator)

	for i, kb := range bindings {
		if !kb.Enabled() {
			continue
		}

		var sep string
		if totalWidth > 0 && i < len(bindings) {
			sep = separator
		}

		str := sep +
			m.Styles.ShortKey.Inline(true).Render(kb.Help().Key) + " " +
			m.Styles.ShortDesc.Inline(true).Render(kb.Help().Desc)

		w := lipgloss.Width(str)

		// If adding this help item would go over the available width, stop
		// drawing.
		if m.Width > 0 && totalWidth+w > m.Width {
			// Although if there's room for an ellipsis, print that.
			tail := " " + m.Styles.Ellipsis.Inline(true).Render(m.Ellipsis)
			tailWidth := lipgloss.Width(tail)

			if totalWidth+tailWidth < m.Width {
				b.WriteString(tail)
			}

			break
		}

		totalWidth += w
		b.WriteString(str)
	}

	return b.String()
}

// FullHelpView renders help columns from a slice of key binding slices. Each
// top level slice entry renders into a column.
func (m Model) FullHelpView(groups [][]key.Binding) string {
	if len(groups) == 0 {
		return ""
	}

	var totalWidth int
	out := make([]string, 0, len(groups))
	sep := m.Styles.FullSeparator.Render(m.FullSeparator)
	sepWidth := lipgloss.Width(sep)

	// Iterate over groups to build columns
	for i, group := range groups {
		if group == nil || !shouldRenderColumn(group) {
			continue
		}

		var (
			keys         []string
			descriptions []string
		)

		// Separate keys and descriptions into different slices
		for _, kb := range group {
			if !kb.Enabled() {
				continue
			}
			keys = append(keys, kb.Help().Key)
			descriptions = append(descriptions, kb.Help().Desc)
		}

		col := lipgloss.JoinHorizontal(lipgloss.Top,
			m.Styles.FullKey.Render(strings.Join(keys, "\n")),
			m.Styles.FullKey.Render(" "),
			m.Styles.FullDesc.Render(strings.Join(descriptions, "\n")),
		)

		// Column
		totalWidth += lipgloss.Width(col)
		if m.Width > 0 && totalWidth > m.Width {
			break
		}

		out = append(out, col)

		// Separator
		if i < len(group)-1 {
			totalWidth += sepWidth
			if m.Width > 0 && totalWidth > m.Width {
				break
			}
		}

		out = append(out, sep)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}

func shouldRenderColumn(b []key.Binding) bool {
	for _, v := range b {
		if v.Enabled() {
			return true
		}
	}
	return false
}
