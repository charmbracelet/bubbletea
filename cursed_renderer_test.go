package tea

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func renderScrolledFrame(t *testing.T, scrollOptim, resetRenderer bool) string {
	t.Helper()
	var output bytes.Buffer
	r := newCursedRenderer(&output, []string{"TERM=xterm-256color"}, 10, 5, scrollOptim)
	if resetRenderer {
		r.reset()
	}
	r.render(View{Content: "AAAAAAAAAA\nBBBBBBBBBB\nCCCCCCCCCC\nDDDDDDDDDD\nEEEEEEEEEE", AltScreen: true})
	if err := r.flush(false); err != nil {
		t.Fatalf("initial flush failed: %v", err)
	}
	output.Reset()
	r.render(View{Content: "CCCCCCCCCC\nDDDDDDDDDD\nEEEEEEEEEE\nFFFFFFFFFF\nGGGGGGGGGG", AltScreen: true})
	if err := r.flush(false); err != nil {
		t.Fatalf("scrolled flush failed: %v", err)
	}
	return output.String()
}

func TestCursedRenderer_scrollOptimization(t *testing.T) {
	for _, resetRenderer := range []bool{false, true} {
		resetRenderer := resetRenderer
		t.Run(fmt.Sprintf("reset=%v", resetRenderer), func(t *testing.T) {
			t.Parallel()
			enabled := renderScrolledFrame(t, true, resetRenderer)
			disabled := renderScrolledFrame(t, false, resetRenderer)

			if !strings.Contains(enabled, "\x1b[2S") {
				t.Fatalf("enabled renderer did not use scroll optimization: %q", enabled)
			}
			if strings.Contains(disabled, "\x1b[2J") {
				t.Fatalf("disabled renderer used a full-screen clear for a regular frame update: %q", disabled)
			}
			for _, sequence := range []string{"\x1b[S", "\x1b[2S", "\x1b[T", "\x1b[2T", "\x1b[L", "\x1b[M"} {
				if strings.Contains(disabled, sequence) {
					t.Fatalf("disabled renderer emitted scroll optimization sequence %q: %q", sequence, disabled)
				}
			}
		})
	}
}

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
