package main

import (
	"fmt"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbletea/teatest"
)

func TestApp(t *testing.T) {
	m := model(10)
	teatest.TestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
		teatest.WithProgramInteractions(func(p teatest.Program, in io.Writer) {
			time.Sleep(time.Second + time.Millisecond*200)
			p.Send("ignored msg")
			p.Send(tea.KeyMsg{
				Type: tea.KeyEnter,
			})
		}),
		teatest.WithRequiredOutputChecker(func(out []byte) {
			teatest.RequireRegexpOutput(t, out, `This program will exit in \d+ seconds`)
			teatest.RequireEqualOutput(t, out)
		}),
		teatest.WithValidateFinalModel(func(mm tea.Model) error {
			m := mm.(model)
			if m != 10 {
				return fmt.Errorf("expected model to be 10, was %d", m)
			}
			return nil
		}),
	)
}
