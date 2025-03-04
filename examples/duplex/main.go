package main

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"math"
	"math/rand/v2"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	msgChan := make(chan message)
	userChan := make(chan username)

	go messageHandlerService(msgChan, userChan, ctx)
	m := initialModel(msgChan, userChan, cancel)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}

}

type username string

type message string

var dummyUsers = [4]username{"meowgorithm", "muesli", "aymanbagabas", "bashbunni"}

// messageHandlerService is a dummy service, in a real world scenario, it could be
// - A background service that may read from a socket connection and write messages to messageChan
// - A background service tht may store messages to SQLite and then pass them to messageChan
func messageHandlerService(msgChan chan<- message, modelChan <-chan username, ctx context.Context) {
	var userSelectedInModel username // the user for which we'll send messages to TUI
	var i uint64
	for i = 1; i <= math.MaxUint64; i++ {
		select {
		case <-ctx.Done():
			close(msgChan)
			return

		// here the model tells which conversation(username) is selected in the TUI
		case userSelectedInModel = <-modelChan:

		case <-time.After(100 * time.Millisecond): // some delay, so our machine doesn't fly away
			randIdx := rand.Uint32N(uint32(len(dummyUsers)))
			randUsr := dummyUsers[randIdx]
			dummyMsg := message(fmt.Sprintf("Message #%d, from %s", i, randUsr))
			// save all to DB
			saveToSQLite(dummyMsg, randUsr)
			// send the filtered one to Model
			if randUsr == userSelectedInModel {
				msgChan <- dummyMsg
			}
		}
	}
}

func saveToSQLite(msg message, u username) {
	// save here
}
