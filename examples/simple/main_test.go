package main

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbletea/teatest"
)

func TestApp(t *testing.T) {
	m := model(10)

	teatest.TestModel(
		t, m,
		func(p teatest.Sender, in io.Writer) error {
			time.Sleep(time.Second)
			p.Send("ignored msg")
			p.Send(tea.KeyMsg{
				Type: tea.KeyEnter,
			})
			return nil
		},
		func(tb testing.TB, out io.Reader) {
			bts, err := io.ReadAll(out)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(bts), "9 seconds") {
				t.Errorf("expected to exit immediately, seems like it waited: %q", string(bts))
			}
		},
	)
}
