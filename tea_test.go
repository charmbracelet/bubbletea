package tea

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	for _, preventCount := range []uint32{0, 1, 2} {
		t.Run(fmt.Sprintf("prevent_%d", preventCount), func(t *testing.T) {
			t.Parallel()
			testTeaWithFilter(t, preventCount)
		})
	}
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

// runSequence runs cmds as a Sequence against a running Program with the given
// filter, the way the event loop runs one, and reports whether the Sequence
// stopped. It returns only once the Sequence has, so the caller sees exactly
// which commands ran, then ends the program and waits for Run to return.
func runSequence(t *testing.T, m Model, filter func(Model, Msg) Msg, cmds ...Cmd) bool {
	t.Helper()
	p := NewProgram(m, WithInput(&bytes.Buffer{}), WithOutput(&bytes.Buffer{}), WithFilter(filter))
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = p.Run()
	}()

	stopped := p.execSequenceMsg(sequenceMsg(cmds))

	p.Quit()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("program did not exit")
	}
	return stopped
}

func TestTeaSequenceFollowsTheEventLoop(t *testing.T) {
	t.Parallel()

	noop := func() Msg { return nil }
	increment := func() Msg { return incrementMsg{} }
	// Sequence and Batch return a lone command as is, so nest them literally.
	seq := func(cmds ...Cmd) Cmd { return func() Msg { return sequenceMsg(cmds) } }
	batch := func(cmds ...Cmd) Cmd { return func() Msg { return BatchMsg(cmds) } }

	dropFirstQuit := func() func(Model, Msg) Msg {
		var dropped atomic.Bool
		return func(_ Model, msg Msg) Msg {
			if _, ok := msg.(QuitMsg); ok && dropped.CompareAndSwap(false, true) {
				return nil
			}
			return msg
		}
	}
	replaceFirstQuit := func() func(Model, Msg) Msg {
		var replaced atomic.Bool
		return func(_ Model, msg Msg) Msg {
			if _, ok := msg.(QuitMsg); ok && replaced.CompareAndSwap(false, true) {
				return incrementMsg{}
			}
			return msg
		}
	}
	incrementToQuit := func() func(Model, Msg) Msg {
		return func(_ Model, msg Msg) Msg {
			if _, ok := msg.(incrementMsg); ok {
				return QuitMsg{}
			}
			return msg
		}
	}

	tests := []struct {
		name       string
		cmd        Cmd
		filter     func() func(Model, Msg) Msg
		stop       bool
		increments int
	}{
		{"quit", Quit, nil, true, 0},
		{"interrupt", Interrupt, nil, true, 0},
		{"quit in nested sequence", seq(noop, Quit), nil, true, 0},
		{"quit in batch", batch(Quit), nil, true, 0},
		{"quit in batch in nested sequence", seq(batch(Quit)), nil, true, 0},
		{"quit in nested sequence in batch", batch(seq(noop, Quit)), nil, true, 0},
		{"quit dropped by the filter", Quit, dropFirstQuit, false, 0},
		{"quit in batch dropped by the filter", batch(Quit), dropFirstQuit, false, 0},
		{"quit in nested sequence dropped by the filter", seq(noop, Quit), dropFirstQuit, false, 0},
		{"quit replaced by the filter", Quit, replaceFirstQuit, false, 1},
		{"quit in batch replaced by the filter", batch(Quit), replaceFirstQuit, false, 1},
		{"message the filter turns into quit", increment, incrementToQuit, true, 0},
		{"message in batch the filter turns into quit", batch(increment), incrementToQuit, true, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var ranLater atomic.Bool
			laterRan := make(chan struct{})
			later := func() Msg {
				ranLater.Store(true)
				close(laterRan)
				return nil
			}

			// The filter holds a message that ends the program until the next
			// command has run, or for 100ms: a Sequence that does not wait for
			// the event loop's decision runs that command in the meantime.
			var caseFilter func(Model, Msg) Msg
			if tc.filter != nil {
				caseFilter = tc.filter()
			}
			filter := func(m Model, msg Msg) Msg {
				if caseFilter != nil {
					msg = caseFilter(m, msg)
				}
				if isTerminalMsg(msg) {
					select {
					case <-laterRan:
					case <-time.After(100 * time.Millisecond):
					}
				}
				return msg
			}

			m := &testModel{}
			if stopped := runSequence(t, m, filter, tc.cmd, later); stopped != tc.stop {
				t.Errorf("Sequence stopped: got %v, want %v", stopped, tc.stop)
			}
			if ranLater.Load() == tc.stop {
				t.Errorf("next command ran: got %v, want %v", ranLater.Load(), !tc.stop)
			}
			if got, _ := m.counter.Load().(int); got != tc.increments {
				t.Errorf("messages the model counted: got %d, want %d", got, tc.increments)
			}
		})
	}
}

func TestTeaSequenceContinuesWhenFilterRejectsQuit(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	var in bytes.Buffer

	// The first Quit is dropped by the filter, so the sequence must go on to
	// its next command, whose Quit the filter lets through.
	var rejected atomic.Bool
	var ranAfter atomic.Bool
	later := func() Msg {
		ranAfter.Store(true)
		return QuitMsg{}
	}

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
		WithFilter(func(_ Model, msg Msg) Msg {
			if _, ok := msg.(QuitMsg); ok && rejected.CompareAndSwap(false, true) {
				return nil
			}
			return msg
		}),
	)
	go p.Send(sequenceMsg{Quit, later})

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if !ranAfter.Load() {
		t.Fatal("Sequence stopped after a Quit the filter rejected")
	}
}

func TestTeaSequenceContinuesWhenFilterReplacesQuit(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	var in bytes.Buffer

	// The first Quit is replaced by an ordinary message, which the model
	// counts; the sequence must go on, and the program ends on its next Quit.
	var replaced atomic.Bool
	var ranAfter atomic.Bool
	later := func() Msg {
		ranAfter.Store(true)
		return QuitMsg{}
	}

	m := &testModel{}
	p := NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
		WithFilter(func(_ Model, msg Msg) Msg {
			if _, ok := msg.(QuitMsg); ok && replaced.CompareAndSwap(false, true) {
				return incrementMsg{}
			}
			return msg
		}),
	)
	go p.Send(sequenceMsg{Quit, later})

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if !ranAfter.Load() {
		t.Fatal("Sequence stopped after a Quit the filter replaced")
	}
	if m.counter.Load() != 1 {
		t.Fatalf("expected the replacement message to reach the model once, got %v", m.counter.Load())
	}
}

func TestTeaSequenceStopsOnContextDone(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var ranAfter atomic.Bool
	first := func() Msg {
		cancel()
		return incrementMsg{}
	}
	second := func() Msg {
		ranAfter.Store(true)
		return incrementMsg{}
	}

	p := NewProgram(&testModel{},
		WithContext(ctx),
		WithInput(&bytes.Buffer{}),
		WithOutput(&bytes.Buffer{}),
	)
	if !p.execSequenceMsg(sequenceMsg{first, second}) {
		t.Error("Sequence did not report that it stopped")
	}
	if ranAfter.Load() {
		t.Fatal("Sequence ran a command after context cancellation")
	}
}

func TestTeaSend(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	NewProgram(m,
		WithInput(&in),
		WithOutput(&buf),
	)
}

func TestTeaPanic(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

// TestProgressBarStateStringOutOfRange is a regression test for
// https://github.com/charmbracelet/bubbletea/issues/1711: String panicked
// with an index-out-of-range error for any ProgressBarState outside the
// [ProgressBarNone, ProgressBarWarning] range. Since State is an exported
// field on ProgressBar, callers can assign such a value without going
// through NewProgressBar.
func TestProgressBarStateStringOutOfRange(t *testing.T) {
	for _, s := range []ProgressBarState{-1, 5, 100} {
		if got, want := s.String(), "Unknown"; got != want {
			t.Errorf("ProgressBarState(%d).String() = %q, want %q", int(s), got, want)
		}
	}

	for s := ProgressBarNone; s <= ProgressBarWarning; s++ {
		if got := s.String(); got == "Unknown" {
			t.Errorf("ProgressBarState(%d).String() = %q, want a known name", int(s), got)
		}
	}
}
