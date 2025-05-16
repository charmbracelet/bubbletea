package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/segmentio/ksuid"
)

const maxDialogs = 6

// Styles
var (
	bgTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("239")).
			Padding(1, 2)

	bgWhitespace = []lipgloss.WhitespaceOption{
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))),
	}

	dialogWordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E7E1CC"))

	dialogStyle = dialogWordStyle.
			Width(36).
			Height(8).
			Padding(1, 3).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD"))

	hoveredDialogStyle = dialogStyle.
				BorderForeground(lipgloss.Color("#F25D94"))

	specialWordLightColor = lipgloss.Color("#43BF6D")
	specialWordDarkColor  = lipgloss.Color("#73F59F")

	buttonStyle = lipgloss.NewStyle().
			Padding(0, 3).
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#6124DF"))

	hoveredButtonStyle = buttonStyle.
				Background(lipgloss.Color("#FF5F87"))
)

// Model

type model struct {
	specialWordStyle lipgloss.Style
	width, height    int
	dialogs          []dialog
	canvas           *lipgloss.Canvas
	mouseDown        bool
	pressID          string
	dragID           string
	dragOffsetX      int
	dragOffsetY      int
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnableMouseAllMotion,
		tea.RequestBackgroundColor,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.BackgroundColorMsg:
		if msg.IsDark() {
			m.specialWordStyle = m.specialWordStyle.Foreground(specialWordDarkColor)
		} else {
			m.specialWordStyle = m.specialWordStyle.Foreground(specialWordLightColor)
		}

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	case tea.MouseMsg:
		mouse := msg.Mouse()
		hit := m.canvas.Hit(mouse.X, mouse.Y)

		switch msg.(type) {
		case tea.MouseClickMsg:
			if mouse.Button != tea.MouseLeft {
				break
			}
			// hit := m.canvas.HitTest(msg.X, msg.Y)

			// Initial press
			if !m.mouseDown {
				m.mouseDown = true
				m.pressID = ""
				if hit != nil {
					m.pressID = hit.GetID()
				}

				// Did we press on a dialog box?
				for i, d := range m.dialogs {
					if hit != nil && d.id != hit.GetID() {
						continue
					}

					// Init drag
					m.dragID = hit.GetID()
					m.dragOffsetX = mouse.X - d.x
					m.dragOffsetY = mouse.Y - d.y

					if len(m.dialogs) < 2 {
						break
					}

					// Move the one we're going to drag to the end of the slice
					// so that it gets the highest z-index when we do
					// compositing later. There are, of course, lots of other
					// ways you could manage the z-index, too.
					m.dialogs = m.removeDialog(i)
					m.dialogs = append(m.dialogs, d)
					break
				}

				break
			}

		// MouseMotion events are send when the mouse has moved and a mouse
		// button is not pressed.
		case tea.MouseMotionMsg:
			// Dragging
			if m.mouseDown && m.dragID != "" {
				// Find the dialog box we're dragging
				for i := range m.dialogs {
					d := &m.dialogs[i]
					if d.id != m.dragID {
						continue
					}

					// Move the dialog box with the cursor
					if m.dragID == d.id {
						d.x = clamp(mouse.X-(m.dragOffsetX), 0, m.width-lipgloss.Width(d.windowView()))
						d.y = clamp(mouse.Y-(m.dragOffsetY), 0, m.height-lipgloss.Height(d.windowView()))
					}

					break
				}
			}

			// Are we hoving over a dialog box?
			for i := range m.dialogs {

				d := &m.dialogs[i]
				d.hovering = false
				d.hoveringButton = false

				if d.id == hit.GetID() {
					d.hovering = true
					continue
				}
				if d.buttonID == hit.GetID() {
					d.hovering = true
					d.hoveringButton = true
					continue
				}
			}

		case tea.MouseReleaseMsg:

			// Make sure we're releasing on something with an ID. A successful
			// click is a press and release.
			if m.pressID == "" {
				break
			}

			// Did we click a button?
			for i, d := range m.dialogs {
				if hit.GetID() == d.buttonID && m.pressID == d.buttonID {
					// "Close" the window
					m.dialogs = m.removeDialog(i)
					break
				}
			}

			// Clicking the background spawns a new dialog
			if hit.GetID() == "bg" && m.pressID == "bg" {
				if len(m.dialogs) < maxDialogs {
					m.dialogs = append(m.dialogs, m.newDialog(mouse.X, mouse.Y))
				}
			}

			m.mouseDown = false
			m.dragID = ""
			m.pressID = ""
		}
	}

	m.canvas = m.Composite()

	return m, nil
}

func (m model) Composite() *lipgloss.Canvas {
	var body string

	n := len(m.dialogs)
	if n > 0 {
		body += "Drag to move. "
	}
	if n == 0 && n < maxDialogs {
		body += "Click to spawn."
	} else if n >= 1 && n < maxDialogs {
		body += fmt.Sprintf("Click to spawn up to %d more.", maxDialogs-len(m.dialogs))
	}
	body += "\n\nPress q to quit."

	bg := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Top,
		lipgloss.Left,
		bgTextStyle.Render(body),
		bgWhitespace...,
	)

	layers := make([]*lipgloss.Layer, len(m.dialogs)+1)
	layers[0] = lipgloss.NewLayer(bg).
		ID("bg").
		Width(m.width).
		Height(m.height)

	for i, d := range m.dialogs {
		layers[i+1] = d.view().Z(i + 1)
	}

	return lipgloss.NewCanvas(layers...)
}

func (m model) View() string {
	return m.Composite().Render()
}

func (m *model) newDialog(x, y int) (d dialog) {
	d.specialWordStyle = &m.specialWordStyle
	dummyView := d.windowView()
	w := lipgloss.Width(dummyView)
	h := lipgloss.Height(dummyView)
	d.x = clamp(x-w/2, 0, m.width-w)
	d.y = clamp(y-h/2, 0, m.height-h)
	d.text = nextRandomWord()
	d.id = ksuid.New().String()
	d.buttonID = ksuid.New().String()
	return d
}

func (m model) removeDialog(index int) []dialog {
	d := m.dialogs

	if len(d) <= index {
		return m.dialogs
	}

	copy(d[index:], d[index+1:]) // shift
	d[len(d)-1] = dialog{}       // nullify
	return d[:len(d)-1]          // truncate
}

// Dialog Windows

type dialog struct {
	specialWordStyle *lipgloss.Style
	id               string
	buttonID         string
	x, y             int
	text             string
	hovering         bool
	hoveringButton   bool
}

func (d dialog) buttonView() string {
	const label = "Run Away"

	if d.hoveringButton {
		return hoveredButtonStyle.Render(label)
	}
	return buttonStyle.Render(label)
}

func (d dialog) windowView() string {
	var style lipgloss.Style
	if d.hovering {
		style = hoveredDialogStyle
	} else {
		style = dialogStyle
	}

	s := d.specialWordStyle.Render(d.text) + dialogWordStyle.Render(" draws near. Command?")
	return style.Render(s)
}

func (d dialog) view() *lipgloss.Layer {
	const hGap, vGap = 3, 1

	window := d.windowView()
	button := d.buttonView()

	buttonX := lipgloss.Width(window) - lipgloss.Width(button) - 1 - hGap
	buttonY := lipgloss.Height(window) - lipgloss.Height(button) - 1 - vGap

	buttonLayer := lipgloss.NewLayer(button).
		ID(d.buttonID).
		X(buttonX).
		Y(buttonY)

	return lipgloss.NewLayer(window).
		ID(d.id).
		X(d.x).
		Y(d.y).
		AddLayers(buttonLayer)
}

// Main

func main() {
	ksuid.SetRand(ksuid.FastRander)

	path := os.Getenv("TEA_LOGFILE")
	if path != "" {
		f, err := tea.LogToFile(path, "layers")
		if err != nil {
			fmt.Println("could not open logfile:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	if _, err := tea.NewProgram(model{}, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}

func clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
