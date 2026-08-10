package tea

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

type execFinishedMsg struct{ err error }

type testExecModel struct {
	cmd string
	err error
}

type testExecNoInputModel struct{ testExecModel }

func (m *testExecModel) Init() Cmd {
	c := exec.Command(m.cmd) //nolint:gosec
	return ExecProcess(c, func(err error) Msg {
		return execFinishedMsg{err}
	})
}

func (m *testExecNoInputModel) Init() Cmd {
	return ExecProcess(successExecCommand(), func(err error) Msg {
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

func (m *testExecModel) View() View {
	return NewView("\n")
}

type spyRenderer struct {
	renderer
	calledReset bool
}

func successExecCommand() *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/c", "exit 0")
	}
	return exec.Command("true")
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
			t.Parallel()
			var buf bytes.Buffer
			var in bytes.Buffer

			m := &testExecModel{cmd: test.cmd}
			p := NewProgram(m,
				WithInput(&in),
				WithOutput(&buf),
			)
			if _, err := p.Run(); err != nil {
				t.Error(err)
			}
			p.renderer = &spyRenderer{renderer: p.renderer}

			if m.err != nil && !test.expectErr {
				t.Errorf("expected no error, got %v", m.err)

				if !p.renderer.(*spyRenderer).calledReset {
					t.Error("expected renderer to be reset")
				}
			}
			if m.err == nil && test.expectErr {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestTeaExecWithNilInput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer

	m := &testExecNoInputModel{}
	p := NewProgram(m,
		WithInput(nil),
		WithOutput(&buf),
	)

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if m.err != nil {
		t.Fatalf("expected no error, got %v", m.err)
	}
}

// execLeakModel renders a distinctive marker from View() so a test can detect
// whether the renderer flushes the View to stdout when ExecProcess hands the
// terminal to a subprocess. Before the fix for #431, the renderer's final
// flush on releaseTerminal wrote the last View() output to stdout, where it
// persisted after the subprocess exited.
type execLeakModel struct {
	done      bool
	altScreen bool
}

func (m *execLeakModel) Init() Cmd {
	return ExecProcess(successExecCommand(), func(err error) Msg {
		return execFinishedMsg{err}
	})
}

func (m *execLeakModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(execFinishedMsg); ok {
		m.done = true
		return m, Quit
	}
	return m, nil
}

func (m *execLeakModel) View() View {
	v := NewView("EXEC_MARKER")
	if m.done {
		v = NewView("QUIT_MARKER")
	}
	v.AltScreen = m.altScreen
	return v
}

// TestExecProcessDoesNotLeakViewOutput is a regression test for #431: when
// ExecProcess runs a subprocess, the renderer must not flush the last View()
// to stdout. The program renders EXEC_MARKER until the subprocess finishes and
// QUIT_MARKER afterwards, so EXEC_MARKER may only ever reach stdout through
// the exec hand-off. WithFPS(1) makes the renderer's ticker fire once per
// second, so during this short test the only flushes are the ones performed
// when releasing the terminal for the subprocess and when shutting down, which
// keeps the assertion deterministic.
func TestExecProcessDoesNotLeakViewOutput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &execLeakModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
		WithWindowSize(80, 24),
		WithFPS(1),
	)
	if _, err := p.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "EXEC_MARKER") {
		t.Errorf("View output leaked to stdout before the subprocess ran; the renderer must not flush the last View() when handing the terminal to a subprocess (see #431)")
	}
	if !strings.Contains(out, "QUIT_MARKER") {
		t.Errorf("expected the final View() (QUIT_MARKER) to be rendered on shutdown, got output: %q", out)
	}
}

// TestExecProcessAltscreenNoLeak verifies the #431 fix also holds in alt screen
// mode and that the program leaves the terminal in a clean state. The fix
// preserves the AltScreen flag when blanking the view, so the screen mode is
// restored when the program resumes; every alt screen enter must therefore be
// paired with a leave.
func TestExecProcessAltscreenNoLeak(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &execLeakModel{altScreen: true}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
		WithWindowSize(80, 24),
		WithFPS(1),
	)
	if _, err := p.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "EXEC_MARKER") {
		t.Errorf("alt screen View output leaked to stdout before the subprocess ran (see #431)")
	}
	const (
		enterAlt = "\x1b[?1049h"
		leaveAlt = "\x1b[?1049l"
	)
	enter := strings.Count(out, enterAlt)
	leave := strings.Count(out, leaveAlt)
	if enter == 0 {
		t.Errorf("expected the program to enter the alt screen at least once, got output: %q", out)
	}
	if enter != leave {
		t.Errorf("alt screen enter/leave unbalanced: enters=%d leaves=%d (output: %q)", enter, leave, out)
	}
}
