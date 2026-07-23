package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
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

func TestCursedRenderer_insertAboveFlushesQueuedView(t *testing.T) {
	var output bytes.Buffer
	r := newCursedRenderer(
		&output,
		[]string{"TERM=xterm-256color", "TERM_PROGRAM=Apple_Terminal"},
		80,
		24,
	)

	r.render(NewView("stale frame\nstale row"))
	if err := r.flush(false); err != nil {
		t.Fatalf("flush initial view: %v", err)
	}

	output.Reset()
	r.render(NewView("current frame"))
	if err := r.insertAbove("committed block"); err != nil {
		t.Fatalf("insert above: %v", err)
	}

	if r.lastView == nil || r.lastView.Content != "current frame" {
		t.Fatalf("insertAbove did not flush the queued view: %#v", r.lastView)
	}
	written := output.String()
	currentAt := strings.Index(written, "current frame")
	committedAt := strings.Index(written, "committed block")
	if currentAt < 0 || committedAt < 0 {
		t.Fatalf("output missing markers (current=%d, committed=%d): %q", currentAt, committedAt, written)
	}
	if currentAt > committedAt {
		t.Fatalf("inserted scrollback before flushing current frame: %q", written)
	}
}
