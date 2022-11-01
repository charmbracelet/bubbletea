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

			time.Sleep(time.Second)
			if bts := readBts(t, out); !bytes.Contains(bts, []byte("This program will exit in 8 seconds")) {
				t.Fatalf("output does not match: expected %q", string(bts))
			}

			p.Send(tea.KeyMsg{
				Type: tea.KeyEnter,
			})
		}),
		teatest.WithValidateFinalModel(func(mm tea.Model) error {
			m := mm.(model)
			if m != 8 {
				return fmt.Errorf("expected model to be 8, was %d", m)
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
