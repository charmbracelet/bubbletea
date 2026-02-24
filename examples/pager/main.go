package main

// An example program demonstrating the pager component from the Bubbles
// component library.

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type model struct {
	content  string
	ready    bool
	viewport viewport.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(viewport.WithWidth(msg.Width), viewport.WithHeight(msg.Height-verticalMarginHeight))
			m.viewport.YPosition = headerHeight
			m.viewport.LeftGutterFunc = func(info viewport.GutterContext) string {
				if info.Soft {
					return "     │ "
				}
				if info.Index >= info.TotalLines {
					return "   ~ │ "
				}
				return fmt.Sprintf("%4d │ ", info.Index+1)
			}
			m.viewport.HighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("34"))
			m.viewport.SelectedHighlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Background(lipgloss.Color("47"))
			m.viewport.SetContent(m.content)
			m.viewport.SetHighlights(regexp.MustCompile("artichoke").FindAllStringIndex(m.content, -1))
			m.viewport.HighlightNext()
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - verticalMarginHeight)
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true                    // use the full size of the terminal in its "alternate screen buffer"
	v.MouseMode = tea.MouseModeCellMotion // turn on mouse support so we can track the mouse wheel
	if !m.ready {
		v.SetContent("\n  Initializing...")
	} else {
		v.SetContent(fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView()))
	}
	return v
}

func (m model) headerView() string {
	title := titleStyle.Render("Mr. Pager")
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%:%3.f%%", m.viewport.ScrollPercent()*100, m.viewport.HorizontalScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func main() {
	// Load some text for our viewport
	content, err := os.ReadFile("artichoke.md")
	if err != nil {
		fmt.Println("could not load file:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		model{content: string(content)},
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
