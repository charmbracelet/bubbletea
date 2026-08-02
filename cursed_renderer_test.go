package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
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

func TestCursedRenderer_restoresKittyKeyboardStack(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	r := newCursedRenderer(&out, []string{"TERM=xterm-256color"}, 80, 24)

	view := NewView("hello")
	view.KeyboardEnhancements.ReportEventTypes = true
	r.render(view)
	if err := r.flush(false); err != nil {
		t.Fatal(err)
	}
	if err := r.close(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	wantPush := ansi.PushKittyKeyboard(keyboardEnhancementsFlags(view.KeyboardEnhancements))
	if !strings.Contains(got, wantPush) {
		t.Fatalf("expected kitty keyboard protocol to be pushed with %q, got %q", wantPush, got)
	}
	wantPop := ansi.PopKittyKeyboard(0)
	if !strings.Contains(got, wantPop) {
		t.Fatalf("expected kitty keyboard protocol to be popped with %q, got %q", wantPop, got)
	}
	if strings.Contains(got, ansi.KittyKeyboard(0, 1)) {
		t.Fatalf("expected kitty keyboard protocol not to be reset in-place with %q, got %q", ansi.KittyKeyboard(0, 1), got)
	}
}
