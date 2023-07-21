package simple

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/muesli/termenv"
	"github.com/rprtr258/bubbletea/lipgloss"
	"github.com/stretchr/testify/assert"

	tea "github.com/rprtr258/bubbletea"
	"github.com/rprtr258/bubbletea/x/exp/teatest"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

func TestApp(t *testing.T) {
	m := model(10)
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
	)
	t.Cleanup(func() {
		assert.NoError(t, tm.Quit())
	})

	time.Sleep(time.Second + time.Millisecond*200)
	tm.Type("I'm typing things, but it'll be ignored by my program")
	tm.Send("ignored msg")
	tm.Send(tea.MsgKey{
		Type: tea.KeyEnter,
	})

	assert.NoError(t, tm.Quit())

	out := readBts(t, tm.FinalOutput(t))
	assert.Regexp(t, `This program will exit in \d+ seconds`, string(out))
	teatest.RequireEqualOutput(t, out)

	assert.Equal(t, model(9), tm.FinalModel(t))
}

func TestAppInteractive(t *testing.T) {
	m := model(10)
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(70, 30),
	)

	time.Sleep(time.Second + time.Millisecond*200)
	tm.Send("ignored msg")

	assert.Contains(t, string(readBts(t, tm.Output())), "This program will exit in 9 seconds")

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("This program will exit in 7 seconds"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.MsgKey{
		Type: tea.KeyEnter,
	})

	assert.NoError(t, tm.Quit())
	assert.Equal(t, model(7), tm.FinalModel(t))
}

func readBts(t *testing.T, r io.Reader) []byte {
	t.Helper()
	bts, err := io.ReadAll(r)
	assert.NoError(t, err)
	return bts
}
