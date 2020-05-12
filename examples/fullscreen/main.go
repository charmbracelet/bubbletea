package main

// A simple program that counts down from 5 and then exits.

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/boba"
)

type model int

type tickMsg time.Time

func newTickMsg(t time.Time) boba.Msg {
	return tickMsg(t)
}

func main() {
	boba.AltScreen()
	defer boba.ExitAltScreen()
	err := boba.NewProgram(initialize, update, view, subscriptions).Start()
	if err != nil {
		log.Fatal(err)
	}
}

func initialize() (boba.Model, boba.Cmd) {
	return model(5), nil
}

func update(message boba.Msg, mdl boba.Model) (boba.Model, boba.Cmd) {
	m, _ := mdl.(model)

	switch msg := message.(type) {

	case boba.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			fallthrough
		case "esc":
			fallthrough
		case "q":
			return m, boba.Quit
		}

	case tickMsg:
		m -= 1
		if m <= 0 {
			return m, boba.Quit
		}

	}

	return m, nil
}

func subscriptions(_ boba.Model) boba.Subs {
	return boba.Subs{
		"tick": boba.Every(time.Second, newTickMsg),
	}
}

func view(mdl boba.Model) string {
	m, _ := mdl.(model)
	return fmt.Sprintf("\n\n     Hi. This program will exit in %d seconds...", m)
}
