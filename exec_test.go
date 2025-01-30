package tea

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"testing"
)

type execFinishedMsg struct{ err error }

type testExecModel struct {
	cmd string
	err error
}

func newTestExecModel(cmd string) (*testExecModel, Cmd) {
	m := &testExecModel{cmd: cmd}
	c := exec.Command(m.cmd) //nolint:gosec
	return m, ExecProcess(c, func(err error) Msg {
		return execFinishedMsg{err}
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

func (m *testExecModel) View() fmt.Stringer {
	return NewFrame("\n")
}

func TestTeaExec(t *testing.T) {
	type test struct {
		name      string
		cmd       string
		expectErr bool
	}

	// TODO: add more tests for windows
	tests := []test{
		{
			name:      "invalid command",
			cmd:       "invalid",
			expectErr: true,
		},
	}

	if runtime.GOOS != "windows" {
		tests = append(tests, []test{
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
		}...)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var in bytes.Buffer

			p := Program[*testExecModel]{
				Init:   func() (*testExecModel, Cmd) { return newTestExecModel(test.cmd) },
				Update: (*testExecModel).Update,
				View:   (*testExecModel).View,
			}
			p.Input = &in
			p.Output = &buf
			p.ForceInputTTY = true
			if err := p.Run(); err != nil {
				t.Error(err)
			}

			if p.Model.err != nil && !test.expectErr {
				t.Errorf("expected no error, got %v", p.Model.err)
			}
			if p.Model.err == nil && test.expectErr {
				t.Error("expected error, got nil")
			}
		})
	}
}
