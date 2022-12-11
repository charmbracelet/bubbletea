package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

var (
	docStyle      = lipgloss.NewStyle().Padding(1, 2)
	quitTextStyle = lipgloss.NewStyle().Padding(1, 2)
	titleStyle    = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0caf5")).
			Background(lipgloss.Color("#536c9e")).
			Padding(0, 1)
	itemStyle = lipgloss.NewStyle().PaddingLeft(2)
)

type item struct {
	title       string
	path        string
	description string
}

type model struct {
	list   list.Model
	choice string
	path   string
}

type editorFinishedMsg struct{ err error }

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) Path() string        { return i.path }
func (i item) FilterValue() string { return i.title }

func openEditor(path string) tea.Cmd {
	home := os.Getenv("HOME")
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	c := exec.Command("bash", "-c", "clear && cd "+home+"/"+path+" && "+editor)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "escape":
			return m, tea.Quit
		case " ", "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i.title)
				m.path = string(i.path)
			}
			return m, openEditor(m.path)
		}

	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return "\n" + m.list.View()
}

func main() {
	vp := viper.New()

	vp.SetConfigName("config")
	vp.SetConfigType("yaml")
	vp.AddConfigPath("$HOME/.config/p")

	err := vp.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}

	title := vp.GetString("title")
	statusbar := vp.GetBool("status-bar")
	filtering := vp.GetBool("filtering")
	// prjcts := vp.Get("projects")

	projects := []list.Item{
		item{title: "nvim", description: "~/.config/nvim", path: ".config/nvim"},
		item{title: "dwm", description: "~/.config/arco-dwm", path: ".config/arco-dwm"},
		item{title: "zsh", description: "~/.config/zsh", path: ".config/zsh"},
		item{title: "dmenu", description: "~/.config/dmenu", path: ".config/dmenu"},
		item{title: "btop", description: "~/.config/btop", path: ".config/btop"},
		item{title: "tmux", description: "~/.tmux", path: ".tmux"},
		item{
			title:       "st Simple Terminal",
			description: "~/.config/arco-st",
			path:        ".config/arco-st",
		},
		item{
			title:       "lazygit",
			description: "~/.config/lazygit",
			path:        ".config/lazygit",
		},
		item{
			title:       "ranger",
			description: "~/.config/ranger",
			path:        ".config/ranger",
		},
		item{
			title:       "fm file manager",
			description: "~/.config/fm",
			path:        ".config/fm",
		},
		item{title: "moc", description: "~/.moc", path: ".moc"},
		item{
			title:       "p app",
			description: "~/Documents/go/src/github.com/Pheon-Dev/p",
			path:        "Documents/go/src/github.com/Pheon-Dev/p",
		},
		item{
			title:       "go",
			description: "~/Documents/go/src/github.com/Pheon-Dev",
			path:        "Documents/go/src/github.com/Pheon-Dev",
		},
		item{
			title:       "bubbletea",
			description: "~/Documents/go/git/bubbletea/examples",
			path:        "Documents/go/git/bubbletea/examples",
		},
		item{
			title:       "go apps",
			description: "~/Documents/go/git",
			path:        "Documents/go/git",
		},
		item{
			title:       "typescript",
			description: "~/Documents/NextJS/App",
			path:        "Documents/NextJS/App",
		},
	}

	// vp.Set("title", "Configs")
	// vp.Set("status-bar", true)
	// vp.Set("filtering", true)
	// vp.Set("projects", projects)
	// vp.WriteConfig()

	l := list.New(projects, list.NewDefaultDelegate(), 0, 0)
	l.Title = title
	l.SetShowStatusBar(statusbar)
	l.SetFilteringEnabled(filtering)
	l.Styles.Title = titleStyle
	m := model{list: l}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error Running Program : ", err)
		os.Exit(1)
	}
}
