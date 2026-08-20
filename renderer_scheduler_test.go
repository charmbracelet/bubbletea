package tea

import (
	"io"
	"sync/atomic"
	"testing"
	"time"
)

type schedulerTestModel struct {
	content string
}

func (schedulerTestModel) Init() Cmd { return nil }

func (m schedulerTestModel) Update(Msg) (Model, Cmd) { return m, nil }

func (m schedulerTestModel) View() View { return NewView(m.content) }

type schedulerTestRenderer struct {
	nilRenderer
	flushes atomic.Int64
	flushed chan struct{}
}

func newSchedulerTestRenderer() *schedulerTestRenderer {
	return &schedulerTestRenderer{flushed: make(chan struct{}, 16)}
}

func (r *schedulerTestRenderer) flush(bool) error {
	r.flushes.Add(1)
	select {
	case r.flushed <- struct{}{}:
	default:
	}
	return nil
}

func waitForSchedulerFlush(t *testing.T, r *schedulerTestRenderer) {
	t.Helper()
	select {
	case <-r.flushed:
	case <-time.After(time.Second):
		t.Fatal("renderer did not flush")
	}
}

func TestRendererSleepsUntilViewChanges(t *testing.T) {
	const fps = 120
	frameInterval := time.Second / fps
	renderer := newSchedulerTestRenderer()
	program := NewProgram(schedulerTestModel{content: "initial"}, WithFPS(fps), WithOutput(io.Discard))
	program.renderer = renderer
	program.startRenderer()
	t.Cleanup(func() { program.stopRenderer(true) })

	program.render(program.initialModel)
	waitForSchedulerFlush(t, renderer)
	baseline := renderer.flushes.Load()

	time.Sleep(5 * frameInterval)
	if got := renderer.flushes.Load(); got != baseline {
		t.Fatalf("idle renderer flushed %d additional times", got-baseline)
	}

	program.render(schedulerTestModel{content: "changed"})
	waitForSchedulerFlush(t, renderer)
	if got := renderer.flushes.Load(); got != baseline+1 {
		t.Fatalf("view change produced %d flushes, want 1", got-baseline)
	}
}

func TestRendererCoalescesBurstWithinFrameInterval(t *testing.T) {
	const fps = 20
	frameInterval := time.Second / fps
	renderer := newSchedulerTestRenderer()
	program := NewProgram(schedulerTestModel{content: "initial"}, WithFPS(fps), WithOutput(io.Discard))
	program.renderer = renderer
	program.startRenderer()
	t.Cleanup(func() { program.stopRenderer(true) })

	program.render(program.initialModel)
	waitForSchedulerFlush(t, renderer)
	baseline := renderer.flushes.Load()

	for index := range 20 {
		program.render(schedulerTestModel{content: string(rune('a' + index))})
	}
	waitForSchedulerFlush(t, renderer)
	time.Sleep(2 * frameInterval)
	if got := renderer.flushes.Load(); got != baseline+1 {
		t.Fatalf("burst produced %d flushes, want 1", got-baseline)
	}
}

func TestRendererWakesForQueuedTerminalOutput(t *testing.T) {
	renderer := newSchedulerTestRenderer()
	program := NewProgram(schedulerTestModel{content: "initial"}, WithFPS(60), WithOutput(io.Discard))
	program.renderer = renderer
	program.startRenderer()
	t.Cleanup(func() { program.stopRenderer(true) })

	program.execute("terminal query")
	program.render(program.initialModel)
	waitForSchedulerFlush(t, renderer)
}
