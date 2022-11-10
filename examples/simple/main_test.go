package main

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
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
		teatest.WithProgramInteractions(func(p teatest.Program, _ io.Reader) {
			time.Sleep(time.Second + time.Millisecond*200)
			p.Send("ignored msg")
			p.Send(tea.KeyMsg{
				Type: tea.KeyEnter,
			})
		}),
		teatest.WithRequiredOutputChecker(func(out []byte) {
			if !regexp.MustCompile(`This program will exit in \d+ seconds`).Match(out) {
				t.Fatalf("output does not match the given regular expression: %s", string(out))
			}
			teatest.RequireEqualOutput(t, out)
		}),
		teatest.WithValidateFinalModel(func(mm tea.Model) error {
			m := mm.(model)
			if m != 9 {
				return fmt.Errorf("expected model to be 10, was %d", m)
			}
			return nil
		}),
	)
}

func TestAppInteractive(t *testing.T) {
	m := model(10)
	teatest.TestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
		teatest.WithProgramInteractions(func(p teatest.Program, out io.Reader) {
			time.Sleep(time.Second + time.Millisecond*200)
			p.Send("ignored msg")

			if bts := readBts(t, out); !bytes.Contains(bts, []byte("This program will exit in 9 seconds")) {
				t.Fatalf("output does not match: expected %q", string(bts))
			}

			teatest.WaitFor(t, out, func(out []byte) bool {
				return bytes.Contains(out, []byte("This program will exit in 7 seconds"))
			}, 5*time.Second, time.Second)

			p.Send(tea.KeyMsg{
				Type: tea.KeyEnter,
			})
		}),
		teatest.WithValidateFinalModel(func(mm tea.Model) error {
			m := mm.(model)
			if m != 7 {
				return fmt.Errorf("expected model to be 7, was %d", m)
			}
			return nil
		}),
	)
}

func readBts(tb testing.TB, r io.Reader) []byte {
	tb.Helper()
	bts, err := io.ReadAll(r)
	if err != nil {
		tb.Fatal(err)
	}
	return bts
}
