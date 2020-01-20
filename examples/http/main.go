package main

// A simple program that counts down from 5 and then exits.

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/charmbracelet/tea"
)

const url = "https://charm.sh/"

type Model struct {
	Status int
	Error  error
}

type statusMsg int

func main() {
	p := tea.NewProgram(initialize, update, view, nil)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func initialize() (tea.Model, tea.Cmd) {
	return Model{0, nil}, checkServer
}

func update(msg tea.Msg, model tea.Model) (tea.Model, tea.Cmd) {
	m, _ := model.(Model)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			fallthrough
		case "break":
			fallthrough
		case "q":
			return m, tea.Quit
		default:
			return m, nil
		}

	case statusMsg:
		m.Status = int(msg)
		return m, tea.Quit

	case tea.ErrMsg:
		// TODO: get the error out of tea.ErrMsg less hackily
		m.Error = errors.New(msg.Error())
		return m, nil

	default:
		return m, nil
	}
}

func view(model tea.Model) string {
	m, _ := model.(Model)
	if m.Error != nil {
		return fmt.Sprintf("Something went wrong: %s", m.Error)
	} else if m.Status == 0 {
		return fmt.Sprintf("Checking %s...", url)
	}
	return fmt.Sprintf("The server said: %d %s", m.Status, http.StatusText(m.Status))
}

func checkServer(_ tea.Model) tea.Msg {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := c.Get(url)
	if err != nil {
		return tea.NewErrMsg(err.Error())
	}
	return statusMsg(res.StatusCode)
}
