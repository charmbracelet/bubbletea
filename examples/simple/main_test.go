package main

import (
	"io"
	"strconv"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbletea/teatest"
)

func TestApp(t *testing.T) {
	t.Parallel()
	for i := 0; i < 3; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			m := model(10)
			teatest.TestModel(
				t, m,
				func(p teatest.Program, in io.Writer) {
					time.Sleep(time.Second + time.Millisecond*200)
					p.Send("ignored msg")
					p.Send(tea.KeyMsg{
						Type: tea.KeyEnter,
					})
				},
				func(out []byte) {
					teatest.RequireEqualOutput(t, out)
				},
			)
		})
	}
}
