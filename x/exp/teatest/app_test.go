package teatest_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	tea "github.com/rprtr258/bubbletea"
	"github.com/rprtr258/bubbletea/x/exp/teatest"
)

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

	out := readBts(t, tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second)))
	assert.Regexp(t, `This program will exit in \d+ seconds`, string(out))
	teatest.RequireEqualOutput(t, out)

	assert.Equal(t, model(9), tm.FinalModel(t))
}

func TestAppInteractive(t *testing.T) {
	m := model(10)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(70, 30))

	time.Sleep(time.Second + time.Millisecond*200)
	tm.Send("ignored msg")

	bts := readBts(t, tm.Output())
	assert.Contains(t, string(bts), "This program will exit in 9 seconds")

	teatest.WaitFor(t, tm.Output(), func(out []byte) bool {
		return bytes.Contains(out, []byte("This program will exit in 7 seconds"))
	}, teatest.WithDuration(5*time.Second), teatest.WithCheckInterval(time.Millisecond*10))

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

// A model can be more or less any type of data. It holds all the data for a
// program, so often it's a struct. For this simple example, however, all
// we'll need is a simple integer.
type model int

// Init optionally returns an initial command we should run. In this case we
// want to start the timer.
func (m model) Init() tea.Cmd {
	return tick
}

// Update is called when messages are received. The idea is that you inspect the
// message and send back an updated model accordingly. You can also return
// a command, which is a function that performs I/O and returns a message.
func (m model) Update(msg tea.Msg) (model, tea.Cmd) {
	switch msg.(type) {
	case tea.MsgKey:
		return m, tea.Quit
	case tickMsg:
		m--
		if m <= 0 {
			return m, tea.Quit
		}
		return m, tick
	}
	return m, nil
}

// View returns a string based on data in the model. That string which will be
// rendered to the terminal.
func (m model) View(r tea.Renderer) {
	r.Write(fmt.Sprintf("Hi. This program will exit in %d seconds. To quit sooner press any key.\n", m))
}

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
