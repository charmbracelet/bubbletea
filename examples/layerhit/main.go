package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type model struct {
	width, height        int
	selectedItem         int
	hitID                string
	hitStartX, hitStartY int
	hitEndX, hitEndY     int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.LayerHitMsg:
		mo := msg.Mouse
		switch mouse := mo.(type) {
		case tea.MouseMotionMsg:
			if mouse.Button != tea.MouseLeft {
				m.hitID = ""
				m.hitStartX, m.hitStartY = -1, -1
			} else {
				m.hitID = msg.ID
				m.hitEndX, m.hitEndY = mouse.X, mouse.Y
			}
		case tea.MouseReleaseMsg:
			if mouse.Button != tea.MouseLeft {
				m.hitID = ""
				m.hitStartX, m.hitStartY = -1, -1
			} else {
				m.hitID = msg.ID
				m.hitEndX, m.hitEndY = mouse.X, mouse.Y
			}
		case tea.MouseClickMsg:
			if mouse.Button != tea.MouseLeft {
				m.hitID = ""
			} else {
				m.hitID = msg.ID
				m.hitStartX, m.hitStartY = mouse.X, mouse.Y
				m.hitEndX, m.hitEndY = mouse.X, mouse.Y
			}
		default:
			m.hitID = ""
			m.hitStartX, m.hitStartY = -1, -1
		}
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.selectedItem > 0 {
				m.selectedItem--
			}
		case "down", "j":
			if m.selectedItem < len(items)-1 {
				m.selectedItem++
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var v tea.View

	bg := lipgloss.NewLayer("").
		Width(m.width).
		Height(m.height).
		ID("bg")

	layers := []*lipgloss.Layer{bg}

	items := items

	for i, item := range items {
		const gap = 1

		itemContent := item
		itemID := fmt.Sprintf("item-%d", i)
		if m.hitID == itemID {
			if m.hitStartX != -1 && m.hitEndX != -1 && m.hitStartY != -1 && m.hitEndY != -1 {
				// Normalize hit coordinates to ensure they are within bounds
				// of the item.
				x1, x2 := m.hitStartX-2, m.hitEndX-2 // 2 for padding
				if x1 > x2 {
					x1, x2 = x2, x1
				}
				if x1 < 0 {
					x1 = 0
				}
				if x2 > len(itemContent) {
					x2 = len(itemContent)
				}

				// Highlight the selected portion of the content.
				itemContent = itemContent[:x1] +
					lipgloss.NewStyle().Reverse(true).Render(itemContent[x1:x2]) +
					itemContent[x2:]
			}
		}

		itemHeight := lipgloss.Height(item)
		itemWidth := m.width - 2 // 2 for padding
		y := (gap + itemHeight) * i
		if y >= m.height {
			continue
		}

		itemLayer := lipgloss.NewLayer(itemContent).
			Width(itemWidth).
			Height(itemHeight).
			X(2).
			Y(y).
			ID(itemID)

		list := "  "
		if i == m.selectedItem {
			list = "> "
		}

		listLayer := lipgloss.NewLayer(list).
			AddLayers(itemLayer).
			ID(fmt.Sprintf("list-item-%d", i)).
			Width(m.width).
			Height(itemHeight + gap).
			Y(y)
		layers = append(layers, listLayer)
	}

	canvas := lipgloss.NewCanvas(layers...)
	v.Layer = canvas

	return v
}

func main() {
	m := model{}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

// items is a list of lorem ipsum text items to demonstrate the layer hit detection.
var items = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
	"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
	"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
	"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
	"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
	"Curabitur pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius, turpis et commodo pharetra, est eros bibendum elit, nec luctus magna felis sollicitudin mauris.",
	"Integer in mauris eu nibh. Nullam mollis. Etiam vel erat. Sed nunc est, mollis non, cursus non, egestas a, neque.",
	"Phasellus ornare. Fusce mollis. Donec sed odio eros. Donec viverra mi quis quam. Integer ut neque.",
	"Vivamus nisi metus, molestie vel, gravida in, condimentum sit amet, nunc. Nam a nibh. Donec suscipit eros. Nam mi. Proin viverra leo ut odio.",
	"Curabitur malesuada. Vestibulum a velit eu ante scelerisque vulputate. Donec in velit vel ipsum auctor pulvinar. Proin ut ligula vel nunc egestas porttitor.",
	"Morbi lectus risus, iaculis vel, suscipit quis, luctus non, massa. Fusce ac turpis quis ligula lacinia aliquet.",
	"Maecenas leo odio, condimentum id, luctus nec, molestie sed, justo. Praesent venenatis metus at tortor pulvinar varius.",
	"Nullam nulla eros, ultricies sit amet, nonummy id, imperdiet feugiat, pede. Sed lectus. Integer euismod lacus luctus magna.",
	"Quisque cursus, metus vitae pharetra auctor, sem massa mattis sem, at interdum magna augue eget diam.",
	"Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim.",
	"Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet.",
	"Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi.",
	"Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum.",
	"Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus.",
	"Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt.",
	"Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue.",
	"Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum.",
	"Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus.",
	"Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt.",
	"Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue.",
	"Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum.",
	"Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus.",
	"Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt.",
	"Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue.",
	"Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum.",
	"Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus.",
	"Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt.",
	"Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue.",
	"Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum.",
	"Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus.",
	"Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt.",
	"Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue.",
	"Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum.",
	"Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus.",
	"Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt.",
	"Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue.",
	"Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum.",
	"Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus.",
	"Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt.",
	"Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue.",
}
