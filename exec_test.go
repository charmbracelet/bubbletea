package tea

import (
	"bytes"
	"os/exec"
	"testing"
)

type execFinishedMsg struct{ err error }

type testExecModel struct {
	cmd string
	err error
}

func (m testExecModel) Init() Cmd {
	c := exec.Command(m.cmd) //nolint:gosec
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
	tests := []struct {
		name      string
		cmd       string
		expectErr bool
	}{
		{
			name:      "true",
			cmd:       "true",
			expectErr: false,
		},
		{
			name:      "false",
			cmd:       "false",
			expectErr: true,
		},
		{
			name:      "invalid command",
			cmd:       "invalid",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testExecModel{cmd: test.cmd}
			p := NewProgram(m, WithInput(&in), WithOutput(&buf))
			if _, err := p.Run(); err != nil {
				t.Error(err)
			}

			if m.err != nil && !test.expectErr {
				t.Errorf("expected no error, got %v", m.err)
			}
			if m.err == nil && test.expectErr {
				t.Error("expected error, got nil")
			}
		})
	}
}
