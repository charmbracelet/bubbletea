package tea

import (
	"bytes"
	"os/exec"
	"testing"
)

type execFinishedMsg struct{ err error }
type testExecModel struct{ err error }

func (m testExecModel) Init() Cmd {
	c := exec.Command("true") //nolint:gosec
	return ExecProcess(c, func(err error) Msg {
		return execFinishedMsg{err}
	})
}

func (m *testExecModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case execFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, Quit
	}

	return m, nil
}

func (m *testExecModel) View() string {
	return "\n"
}

func TestTeaExec(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testExecModel{}
	p := NewProgram(m, WithInput(&in), WithOutput(&buf))
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}

	if m.err != nil {
		t.Fatal(m.err)
	}
}
