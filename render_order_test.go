package tea

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// sizeModel shows the width it was last given.
type sizeModel struct {
	width int
	views atomic.Int64
}

func (m *sizeModel) Init() Cmd { return nil }

func (m *sizeModel) Update(msg Msg) (Model, Cmd) {
	if msg, ok := msg.(WindowSizeMsg); ok {
		m.width = msg.Width
	}
	return m, nil
}

func (m *sizeModel) View() View {
	m.views.Add(1)
	return NewView(fmt.Sprint(m.width))
}

// sizeRenderer notes the first frame it flushes after a resize.
type sizeRenderer struct {
	nilRenderer
	mu      sync.Mutex
	view    View
	resized bool
	flushed chan string // the first frame after a resize
}

func (r *sizeRenderer) render(v View) {
	r.mu.Lock()
	r.view = v
	r.mu.Unlock()
}

func (r *sizeRenderer) resize(int, int) {
	r.mu.Lock()
	r.resized = true
	r.mu.Unlock()
}

func (r *sizeRenderer) flush(bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.resized {
		r.resized = false
		select {
		case r.flushed <- r.view.Content:
		default:
		}
	}
	return nil
}

// TestRenderAfterResize checks that the frame drawn after a resize is the
// view for the new size, even when a view was rendered just before.
func TestRenderAfterResize(t *testing.T) {
	t.Parallel()
	const fps = 10
	sr := &sizeRenderer{flushed: make(chan string, 1)}
	m := &sizeModel{}
	p := NewProgram(m,
		WithInput(&bytes.Buffer{}),
		WithOutput(io.Discard),
		WithoutSignalHandler(),
		WithWindowSize(80, 24),
		WithFPS(fps),
	)
	p.renderer = sr
	done := make(chan error, 1)
	go func() {
		_, err := p.Run()
		done <- err
	}()
	defer func() {
		p.Quit()
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()

	// Let the startup messages be rendered and drawn.
	time.Sleep(3 * time.Second / fps)
	select {
	case <-sr.flushed:
	default:
	}

	// A view rendered right before the resize.
	views := m.views.Load()
	p.Send(incrementMsg{})
	for m.views.Load() == views {
		time.Sleep(time.Millisecond)
	}
	p.Send(WindowSizeMsg{Width: 120, Height: 24})
	select {
	case got := <-sr.flushed:
		if got != "120" {
			t.Errorf("the first frame drawn after resizing to 120 columns was the view for %s", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no frame was drawn after the resize")
	}
}

// printModel prints the count it reaches with each increment.
type printModel struct {
	count int
	views atomic.Int64
}

func (m *printModel) Init() Cmd { return nil }

func (m *printModel) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(incrementMsg); ok {
		m.count++
		return m, Println(m.count)
	}
	return m, nil
}

func (m *printModel) View() View {
	m.views.Add(1)
	return NewView(fmt.Sprint(m.count))
}

// printRenderer notes the view it has when lines are printed above it.
type printRenderer struct {
	nilRenderer
	mu      sync.Mutex
	view    View
	printed []string // each line printed, with the view below it
}

func (r *printRenderer) render(v View) {
	r.mu.Lock()
	r.view = v
	r.mu.Unlock()
}

func (r *printRenderer) insertAbove(s string) error {
	r.mu.Lock()
	r.printed = append(r.printed, s+" "+r.view.Content)
	r.mu.Unlock()
	return nil
}

// TestRenderBeforePrint checks that a line printed by an update is printed
// above the view of that update or a later one, never an older one, even
// when the view's render was held back.
func TestRenderBeforePrint(t *testing.T) {
	t.Parallel()
	const n = 3
	pr := &printRenderer{}
	m := &printModel{}
	p := NewProgram(m,
		WithInput(&bytes.Buffer{}),
		WithOutput(io.Discard),
		WithoutSignalHandler(),
		WithFPS(10),
	)
	p.renderer = pr
	go func() {
		for m.views.Load() < 2 {
			time.Sleep(time.Millisecond)
		}
		for range n {
			p.Send(incrementMsg{})
		}
		for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
			pr.mu.Lock()
			printed := len(pr.printed)
			pr.mu.Unlock()
			if printed == n {
				break
			}
			time.Sleep(time.Millisecond)
		}
		p.Quit()
	}()
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	pr.mu.Lock()
	defer pr.mu.Unlock()
	if len(pr.printed) != n {
		t.Fatalf("printed %q, want %d lines", pr.printed, n)
	}
	for _, s := range pr.printed {
		var line, shown int
		if _, err := fmt.Sscan(s, &line, &shown); err != nil {
			t.Fatal(err)
		}
		if shown < line {
			t.Errorf("line %d was printed above the view of %d, from before the update that printed it", line, shown)
		}
	}
}
