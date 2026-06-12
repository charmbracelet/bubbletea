package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
)

type mouseRaceModel struct {
	i int
}

func (m *mouseRaceModel) Init() Cmd { return nil }

func (m *mouseRaceModel) Update(msg Msg) (Model, Cmd) {
	switch msg.(type) {
	case MouseClickMsg, MouseMotionMsg, MouseWheelMsg:
		m.i++
	}
	return m, nil
}

func (m *mouseRaceModel) View() View {
	return View{
		Content:   fmt.Sprintf("tick-%d\n", m.i),
		MouseMode: MouseModeCellMotion,
	}
}

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1690
func TestCursedRenderer_mouseVsFlush(t *testing.T) {
	t.Parallel()

	pr, pw := io.Pipe()
	defer func() { _ = pw.Close() }()

	m := &mouseRaceModel{}
	p := NewProgram(
		m,
		WithContext(t.Context()),
		WithInput(pr),
		WithOutput(io.Discard),
		WithEnvironment([]string{
			"TERM=xterm-256color",
			"TERM_PROGRAM=Apple_Terminal",
		}),
		WithoutSignals(),
		WithWindowSize(80, 24),
	)

	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		_, _ = p.Run()
	}()

	time.Sleep(150 * time.Millisecond)

	const iterations = 100
	for i := range iterations {
		switch i % 4 {
		case 0:
			p.Send(MouseClickMsg{X: i % 80, Y: i % 24, Button: MouseLeft})
		case 1:
			p.Send(MouseMotionMsg{X: i % 80, Y: i % 24})
		case 2:
			p.Send(MouseWheelMsg{X: 0, Y: 0, Button: MouseWheelUp})
		default:
			p.Send(MouseReleaseMsg{X: i % 80, Y: i % 24, Button: MouseLeft})
		}
	}

	p.Quit()
	select {
	case <-runDone:
	case <-time.After(5 * time.Second):
		t.Fatal("program did not exit after Quit")
	}
}

// Fixes: https://github.com/charmbracelet/bubbletea/issues/1709
func TestCursedRenderer_insertAboveDownsamples(t *testing.T) {
	t.Parallel()

	const colored = "println: \x1b[38;2;255;0;0mred\x1b[m"
	env := []string{"TERM=xterm-256color"}

	t.Run("ascii", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		s := newCursedRenderer(&buf, env, 80, 24)
		s.setColorProfile(colorprofile.Ascii)

		if err := s.insertAbove(colored); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		if strings.Contains(out, "\x1b[38;2;255;0;0m") {
			t.Errorf("insertAbove kept a color the profile cannot display: %q", out)
		}
		if !strings.Contains(out, "red") {
			t.Errorf("insertAbove dropped message content: %q", out)
		}
	})

	t.Run("truecolor", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		s := newCursedRenderer(&buf, env, 80, 24)
		s.setColorProfile(colorprofile.TrueColor)

		if err := s.insertAbove(colored); err != nil {
			t.Fatal(err)
		}
		if out := buf.String(); !strings.Contains(out, "\x1b[38;2;255;0;0mred") {
			t.Errorf("insertAbove altered content the profile can display: %q", out)
		}
	})
}
