package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type styles struct {
	app           lipgloss.Style
	title         lipgloss.Style
	statusMessage lipgloss.Style
}

func newStyles(darkBG bool) styles {
	lightDark := lipgloss.LightDark(darkBG)

	return styles{
		app: lipgloss.NewStyle().
			Padding(1, 2),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1),
		statusMessage: lipgloss.NewStyle().
			Foreground(lightDark("#04B575", "#04B575")),
	}
}

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	insertItem       key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		insertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

// query tracks which terminal properties have been resolved.
type query int

const (
	backgroundColor = 1 << iota
	windowSize
)

// Ready returns true if all properties necessary for our app to function have
// been resolved.
func (q query) Ready() bool {
	return q == backgroundColor|windowSize
}

type model struct {
	styles        styles
	queries       query
	darkBG        bool
	width, height int
	once          *sync.Once
	list          list.Model
	itemGenerator *randomItemGenerator
	keys          *listKeyMap
	delegateKeys  *delegateKeyMap
}

func (m model) Init() (tea.Model, tea.Cmd) {
	m.once = new(sync.Once)
	return m, tea.Batch(
		tea.BackgroundColor,
		tea.EnterAltScreen,
	)
}

func (m *model) updateListProperties() {
	// Wait until we've queried for the necessary terminal
	// properties. Specifically, we need to know the background color and the
	// window size before we can construct the list.
	if !m.queries.Ready() {
		return
	}

	// Initialize the list, but only once.
	m.once.Do(func() {
		m.styles = newStyles(m.darkBG)

		delegateKeys := newDelegateKeyMap()
		listKeys := newListKeyMap()

		// Make initial list of items.
		var itemGenerator randomItemGenerator
		const numItems = 24
		items := make([]list.Item, numItems)
		for i := 0; i < numItems; i++ {
			items[i] = itemGenerator.next()
		}

		// Setup list.
		delegate := newItemDelegate(delegateKeys, &m.styles)
		groceryList := list.New(items, delegate, 0, 0)
		groceryList.Title = "Groceries"
		groceryList.Styles.Title = m.styles.title
		groceryList.AdditionalFullHelpKeys = func() []key.Binding {
			return []key.Binding{
				listKeys.toggleSpinner,
				listKeys.insertItem,
				listKeys.toggleTitleBar,
				listKeys.toggleStatusBar,
				listKeys.togglePagination,
				listKeys.toggleHelpMenu,
			}
		}

		m.list = groceryList
		m.keys = listKeys
		m.delegateKeys = delegateKeys
		m.itemGenerator = &itemGenerator
	})

	// Update list size.
	h, v := m.styles.app.GetFrameSize()
	m.list.SetSize(m.width-h, m.height-v)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.darkBG = msg.IsDark()
		m.queries |= backgroundColor
		m.updateListProperties()
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.queries |= windowSize
		m.updateListProperties()
		return m, nil
	}

	// Don't proceed until we've queried for the necessary terminal properties
	// above.
	if !m.queries.Ready() {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.toggleSpinner):
			cmd := m.list.ToggleSpinner()
			return m, cmd

		case key.Matches(msg, m.keys.toggleTitleBar):
			v := !m.list.ShowTitle()
			m.list.SetShowTitle(v)
			m.list.SetShowFilter(v)
			m.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.keys.toggleStatusBar):
			m.list.SetShowStatusBar(!m.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil

		case key.Matches(msg, m.keys.insertItem):
			m.delegateKeys.remove.SetEnabled(true)
			newItem := m.itemGenerator.next()
			insCmd := m.list.InsertItem(0, newItem)
			statusCmd := m.list.NewStatusMessage(m.styles.statusMessage.Render("Added " + newItem.Title()))
			return m, tea.Batch(insCmd, statusCmd)
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Don't render until we have everything we queried for.
	if !m.queries.Ready() {
		return ""
	}

	return m.styles.app.Render(m.list.View())
}

func main() {
	if _, err := tea.NewProgram(model{}).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
