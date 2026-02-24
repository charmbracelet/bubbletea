package tea

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type ctxImplodeMsg struct {
	cancel context.CancelFunc
}

type incrementMsg struct{}

type panicMsg struct{}

func panicCmd() Msg {
	panic("testing goroutine panic behavior")
}

type testModel struct {
	executed atomic.Value
	counter  atomic.Value
}

func (m *testModel) Init() Cmd {
	return nil
}

func (m *testModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case ctxImplodeMsg:
		msg.cancel()
		time.Sleep(100 * time.Millisecond)

	case incrementMsg:
		i := m.counter.Load()
		if i == nil {
			m.counter.Store(1)
		} else {
			m.counter.Store(i.(int) + 1)
		}

	case KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, Quit
		}

	case panicMsg:
		panic("testing panic behavior")
	}

	return m, nil
}

func (m *testModel) View() View {
	m.executed.Store(true)
	return NewView("success")
}

func TestTeaModel(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer
	in.Write([]byte("q"))

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()

	p := NewProgram(&testModel{},
		WithContext(ctx),
		WithInput(&in),
		WithOutput(&buf),
	)
	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if buf.Len() == 0 {
		t.Fatal("no output")
	}
}

func TestTeaQuit(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				p.Quit()
				return
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestTeaWaitQuit(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	progStarted := make(chan struct{})
	waitStarted := make(chan struct{})
	errChan := make(chan error, 1)

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)

	go func() {
		_, err := p.Run()
		errChan <- err
	}()

	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				close(progStarted)

				<-waitStarted
				time.Sleep(50 * time.Millisecond)
				p.Quit()

				return
			}
		}
	}()

	<-progStarted

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			p.Wait()
			wg.Done()
		}()
	}
	close(waitStarted)
	wg.Wait()

	err := <-errChan
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}
}

func TestTeaWaitKill(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	progStarted := make(chan struct{})
	waitStarted := make(chan struct{})
	errChan := make(chan error, 1)

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)

	go func() {
		_, err := p.Run()
		errChan <- err
	}()

	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				close(progStarted)

				<-waitStarted
				time.Sleep(50 * time.Millisecond)
				p.Kill()

				return
			}
		}
	}()

	<-progStarted

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			p.Wait()
			wg.Done()
		}()
	}
	close(waitStarted)
	wg.Wait()

	err := <-errChan
	if !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}
}

func TestTeaWithFilter(t *testing.T) {
	testTeaWithFilter(t, 0)
	testTeaWithFilter(t, 1)
	testTeaWithFilter(t, 2)
}

func testTeaWithFilter(t *testing.T, preventCount uint32) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	shutdowns := uint32(0)
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	p.filter = func(_ Model, msg Msg) Msg {
		if _, ok := msg.(QuitMsg); !ok {
			return msg
		}
		if shutdowns < preventCount {
			atomic.AddUint32(&shutdowns, 1)
			return nil
		}
		return msg
	}

	go func() {
		for atomic.LoadUint32(&shutdowns) <= preventCount {
			time.Sleep(time.Millisecond)
			p.Quit()
		}
	}()

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if shutdowns != preventCount {
		t.Errorf("Expected %d prevented shutdowns, got %d", preventCount, shutdowns)
	}
}

func TestTeaKill(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				p.Kill()
				return
			}
		}
	}()

	_, err := p.Run()

	if !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}

	if errors.Is(err, context.Canceled) {
		// The end user should not know about the program's internal context state.
		// The program should only report external context cancellation as a context error.
		t.Fatalf("Internal context cancellation was reported as context error!")
	}
}

func TestTeaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m,
		WithContext(ctx),
		WithInput(&in),
		WithOutput(&buf),
	)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				cancel()
				return
			}
		}
	}()

	_, err := p.Run()

	if !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}

	if !errors.Is(err, context.Canceled) {
		// The end user should know that their passed in context caused the kill.
		t.Fatalf("Expected %v, got %v", context.Canceled, err)
	}
}

func TestTeaContextImplodeDeadlock(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m,
		WithContext(ctx),
		WithInput(&in),
		WithOutput(&buf),
	)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				p.Send(ctxImplodeMsg{cancel: cancel})
				return
			}
		}
	}()

	if _, err := p.Run(); !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}
}

func TestTeaContextBatchDeadlock(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	var buf bytes.Buffer
	var in bytes.Buffer

	inc := func() Msg {
		cancel()
		return incrementMsg{}
	}

	m := &testModel{}
	p := NewProgram(m,
		WithContext(ctx),
		WithInput(&in),
		WithOutput(&buf),
	)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				batch := make(BatchMsg, 100)
				for i := range batch {
					batch[i] = inc
				}
				p.Send(batch)
				return
			}
		}
	}()

	if _, err := p.Run(); !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}
}

func TestTeaBatchMsg(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	inc := func() Msg {
		return incrementMsg{}
	}

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	go func() {
		p.Send(BatchMsg{inc, inc})

		for {
			time.Sleep(time.Millisecond)
			i := m.counter.Load()
			if i != nil && i.(int) >= 2 {
				p.Quit()
				return
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if m.counter.Load() != 2 {
		t.Fatalf("counter should be 2, got %d", m.counter.Load())
	}
}

func TestTeaSequenceMsg(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	inc := func() Msg {
		return incrementMsg{}
	}

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	go p.Send(sequenceMsg{inc, inc, Quit})

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if m.counter.Load() != 2 {
		t.Fatalf("counter should be 2, got %d", m.counter.Load())
	}
}

func TestTeaSequenceMsgWithBatchMsg(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	inc := func() Msg {
		return incrementMsg{}
	}
	batch := func() Msg {
		return BatchMsg{inc, inc}
	}

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	go p.Send(sequenceMsg{batch, inc, Quit})

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if m.counter.Load() != 3 {
		t.Fatalf("counter should be 3, got %d", m.counter.Load())
	}
}

func TestTeaNestedSequenceMsg(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	inc := func() Msg {
		return incrementMsg{}
	}

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	go p.Send(sequenceMsg{inc, Sequence(inc, inc, Batch(inc, inc)), Quit})

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if m.counter.Load() != 5 {
		t.Fatalf("counter should be 5, got %d", m.counter.Load())
	}
}

func TestTeaSend(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)

	// sending before the program is started is a blocking operation
	go p.Send(Quit())

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	// sending a message after program has quit is a no-op
	p.Send(Quit())
}

func TestTeaNoRun(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
}

func TestTeaPanic(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				p.Send(panicMsg{})
				return
			}
		}
	}()

	_, err := p.Run()

	if !errors.Is(err, ErrProgramPanic) {
		t.Fatalf("Expected %v, got %v", ErrProgramPanic, err)
	}

	if !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}
}

func TestTeaGoroutinePanic(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				batch := make(BatchMsg, 10)
				for i := 0; i < len(batch); i += 2 {
					batch[i] = Sequence(panicCmd)
					batch[i+1] = Batch(panicCmd)
				}
				p.Send(batch)
				return
			}
		}
	}()

	_, err := p.Run()

	if !errors.Is(err, ErrProgramPanic) {
		t.Fatalf("Expected %v, got %v", ErrProgramPanic, err)
	}

	if !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}
}

type benchModel struct {
	t testing.TB
}

func (m benchModel) Init() Cmd {
	return nil
}

func (m benchModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, Quit
		}
	}

	return m, nil
}

func (m benchModel) View() View {
	view := strings.Join([]string{
		" \x1b[38;5;63m╭─────────────────────────╮\x1b[m",
		" \x1b[38;5;63m│\x1b[m\x1b[25X\x1b[28G\x1b[38;5;63m│\x1b[m",
		" \x1b[38;5;63m│\x1b[m    \x1b[38;5;231mHello There!\x1b[m    \x1b[38;5;63m│\x1b[m",
		" \x1b[38;5;63m│\x1b[m\x1b[25X\x1b[28G\x1b[38;5;63m│\x1b[m",
		" \x1b[38;5;63m╰─────────────────────────╯\x1b[m",
	}, "\n")

	return NewView(view)
}

func BenchmarkTeaRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer

		m := benchModel{b}
		r, w := io.Pipe()
		p := NewProgram(m,
			WithInput(r),
			WithOutput(&buf),
		)

		go func() {
			for _, input := range "abcdefghijklmnopq" {
				time.Sleep(10 * time.Millisecond)
				w.Write([]byte(string(input)))
			}
		}()

		if _, err := p.Run(); err != nil {
			b.Fatalf("Run failed: %v", err)
		}

		_ = r.CloseWithError(io.EOF)
	}
}
