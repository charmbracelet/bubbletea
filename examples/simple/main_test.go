package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestApp(t *testing.T) {
	m := model(10)

	var in bytes.Buffer
	var out bytes.Buffer

	p := tea.NewProgram(m, tea.WithInput(&in), tea.WithOutput(&out))
	done := make(chan bool, 1)

	go func() {
		if err := p.Start(); err != nil {
			t.Error(err)
		}
		done <- true
	}()

	time.Sleep(time.Second)
	p.Send("ignored msg")
	p.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	<-done

	if !strings.Contains(out.String(), "9 seconds") {
		t.Errorf("expected to exit immediately, seems like it waited: %q", out.String())
	}
}
