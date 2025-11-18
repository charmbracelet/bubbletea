package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	gotReposSuccessMsg []repo
	gotReposErrMsg     error
)

type repo struct {
	Name string `json:"name"`
}

const reposURL = "https://api.github.com/orgs/charmbracelet/repos"

func getRepos() tea.Msg {
	req, err := http.NewRequest(http.MethodGet, reposURL, nil)
	if err != nil {
		return gotReposErrMsg(err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return gotReposErrMsg(err)
	}
	defer resp.Body.Close() // nolint: errcheck

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return gotReposErrMsg(err)
	}

	var repos []repo

	err = json.Unmarshal(data, &repos)
	if err != nil {
		return gotReposErrMsg(err)
	}

	return gotReposSuccessMsg(repos)
}

type model struct {
	textInput textinput.Model
	help      help.Model
	keymap    keymap
}

type keymap struct {
	complete, next, prev, quit key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.complete,
		k.next,
		k.prev,
		k.quit,
	}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func initialModel() model {
	ti := textinput.New()
	ti.Prompt = "charmbracelet/"

	s := ti.Styles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).MarginLeft(2)
	s.Cursor.Color = lipgloss.Color("63")
	ti.SetStyles(s)

	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 50
	ti.SetWidth(20)
	ti.ShowSuggestions = true

	km := keymap{
		// XXX: we should be using the keybindings on the textinput model.
		complete: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete"), key.WithDisabled()),
		next:     key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "next"), key.WithDisabled()),
		prev:     key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "prev"), key.WithDisabled()),

		quit: key.NewBinding(key.WithKeys("enter", "ctrl+c", "esc"), key.WithHelp("esc", "quit")),
	}

	return model{
		textInput: ti,
		keymap:    km,
		help:      help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(getRepos, textinput.Blink)
}

func (m model) Cursor() *tea.Cursor {
	c := m.textInput.Cursor()
	if c != nil {
		c.Y += lipgloss.Height(m.headerView())
	}
	return c
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case gotReposSuccessMsg:
		var suggestions []string
		for _, r := range msg {
			suggestions = append(suggestions, r.Name)
		}
		m.textInput.SetSuggestions(suggestions)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	// Determine whether to show completion keybindings.
	//
	// XXX: we should be using the keybindings on the textinput model.
	hasChoices := len(m.textInput.MatchedSuggestions()) > 1
	m.keymap.complete.SetEnabled(hasChoices)
	m.keymap.next.SetEnabled(hasChoices)
	m.keymap.prev.SetEnabled(hasChoices)

	return m, cmd
}

func (m model) View() tea.View {
	if len(m.textInput.AvailableSuggestions()) < 1 {
		return tea.NewView("One sec, we're fetching completions...")
	}

	v := tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		m.headerView(),
		m.textInput.View(),
		m.footerView(),
	))

	c := m.textInput.Cursor()
	if c != nil {
		c.Y += lipgloss.Height(m.headerView())
	}
	v.Cursor = c
	return v
}

func (m model) headerView() string { return "Enter a Charm™ repo:\n" }
func (m model) footerView() string { return "\n" + m.help.View(m.keymap) }
