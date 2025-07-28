package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"examples/nav/nav"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	_          nav.PageLife = helpPage{}
	fullScreen              = false
)

func main() {
	p := tea.NewProgram(home{})
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type home struct{}

func (m home) Init() tea.Cmd {
	return nav.Push(splash{Name: "splash"})
}

func (m home) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			return m, nav.PushOrBack(helpPage{val: rand.Intn(100)})
		case "a":
			return m, nav.Push(apiPage{})
		case "b":
			return m, nav.Back()
		case "ctrl+c":
			return m, tea.Quit
		}

	}

	return m, nav.Update(msg)
}

func (m home) View() string {

	if fullScreen {
		return nav.View()
	}

	return fmt.Sprintf(
		"%s\n%s\n%s\nhistories=%d splash=%+v",
		"****************************************",
		nav.View(),
		"****************************************",
		len(nav.Histories()),
		nav.CurrentPage(),
	)
}

/****************************************
* more splash
****************************************/
type splash struct {
	Name string
}

func (m splash) Init() tea.Cmd {
	return nil
}
func (m splash) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m splash) View() string {
	return m.Name
}

type helpPage struct {
	val  int
	seek int
}

func (m helpPage) OnEntering() (tea.Model, tea.Cmd) {
	fullScreen = true
	m.seek = rand.Intn(100)
	return m, nil
}

func (m helpPage) OnLeaving() (tea.Model, tea.Cmd) {
	fullScreen = false
	return m, nil
}

func (m helpPage) Init() tea.Cmd {
	return nil
}
func (m helpPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
func (m helpPage) View() string {
	return fmt.Sprintf("v:%d s:%d back change: press a to api back", m.val, m.seek)
}

type apiPage struct {
	data string
}
type apiMsg string

func (m apiPage) Init() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Second * 3)
		return apiMsg("api done")
	}
}
func (m apiPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case apiMsg:
		m.data = string(msg)
	}

	return m, nil
}

func (m apiPage) View() string {

	if len(m.data) == 0 {
		return "api loading..."
	}

	return fmt.Sprintf("go to other page back api page don't need load api\n%s", m.data)
}
