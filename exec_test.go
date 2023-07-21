package tea

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

type execFinishedMsg struct {
	err error
}

type testExecModel struct {
	cmd string
	err error
}

func (m testExecModel) Init() Cmd {
	c := exec.Command(m.cmd) //nolint:gosec
	return ExecProcess(c, func(err error) Msg {
		return execFinishedMsg{err: err}
	})
}

func (m *testExecModel) Update(msg Msg) (*testExecModel, Cmd) {
	switch msg := msg.(type) {
	case execFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		return m, Quit
	}

	return m, nil
}

func (m *testExecModel) View(Renderer) {}

func TestTeaExec(t *testing.T) {
	for name, test := range map[string]struct {
		cmd       string
		expectErr error
	}{
		"true": {
			cmd:       "true",
			expectErr: nil,
		},
		"false": {
			cmd:       "false",
			expectErr: &exec.ExitError{},
		},
		"invalid command": {
			cmd:       "invalid",
			expectErr: &exec.Error{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testExecModel{cmd: test.cmd}
			_, err := NewProgram(m).WithInput(&in).WithOutput(&buf).Run()
			assert.NoError(t, err)
			assert.IsType(t, test.expectErr, m.err)
		})
	}
}
