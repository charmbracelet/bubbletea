package main

// A simple program that counts down from 5 and then exits.

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const url = "https://charm.sh/"

type Model struct {
	status int
	err    error
}

type statusMsg int
type errMsg error

func main() {
	p := tea.NewProgram(initialize, update, view)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func initialize() (tea.Model, tea.Cmd) {
	return Model{0, nil}, checkServer
}

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, ok := model.(Model)
	if !ok {
		return Model{err: errors.New("could not perform assertion on model during update")}, nil
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			fallthrough
		case "ctrl+c":
			fallthrough
		case "q":
			return m, tea.Quit
		default:
			return m, nil
		}

	case statusMsg:
		m.status = int(msg)
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, nil

	default:
		return m, nil
	}
}

func view(model tea.Model) string {
	m, _ := model.(Model)
	s := fmt.Sprintf("Checking %s...", url)
	if m.err != nil {
		s += fmt.Sprintf("something went wrong: %s", m.err)
	} else if m.status != 0 {
		s += fmt.Sprintf("%d %s", m.status, http.StatusText(m.status))
	}
	return s + "\n"
}

func checkServer() tea.Msg {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := c.Get(url)
	if err != nil {
		return errMsg(err)
	}
	return statusMsg(res.StatusCode)
}
