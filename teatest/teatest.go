package teatest

import (
	"bytes"
	"io"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type Sender interface {
	Send(tea.Msg)
}

func TestModel(
	tb testing.TB,
	m tea.Model,
	interact func(p Sender, in io.Writer),
	assert func(tb testing.TB, out io.Reader),
) {
	var in bytes.Buffer
	var out bytes.Buffer

	p := tea.NewProgram(m, tea.WithInput(&in), tea.WithOutput(&out))
	done := make(chan bool, 1)

	go func() {
		if err := p.Start(); err != nil {
			tb.Fatalf("app failed: %s", err)
		}
		done <- true
	}()

	interact(p, &in)
	<-done
	assert(tb, &out)
}
