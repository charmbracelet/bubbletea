package tea

import (
	"bytes"
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type incrementMsg struct{}

type testModel struct {
	executed atomic.Value
	counter  atomic.Value
}

func (m testModel) Init() Cmd {
	return nil
}

func (m *testModel) Update(msg Msg) (Model, Cmd) {
	switch msg.(type) {
	case incrementMsg:
		i := m.counter.Load()
		if i == nil {
			m.counter.Store(1)
		} else {
			m.counter.Store(i.(int) + 1)
		}

	case KeyMsg:
		return m, Quit
	}

	return m, nil
}

func (m *testModel) View() string {
	m.executed.Store(true)
	return "success\n"
}

func TestTeaModel(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer
	in.Write([]byte("q"))

	p := NewProgram(&testModel{}, WithInput(&in), WithOutput(&buf))
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
	p := NewProgram(m, WithInput(&in), WithOutput(&buf))
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
		WithFilter(func(_ Model, msg Msg) Msg {
			if _, ok := msg.(QuitMsg); !ok {
				return msg
			}
			if shutdowns < preventCount {
				atomic.AddUint32(&shutdowns, 1)
				return nil
			}
			return msg
		}))

	go func() {
		for atomic.LoadUint32(&shutdowns) <= preventCount {
			time.Sleep(time.Millisecond)
			p.Quit()
		}
	}()

	if err := p.Start(); err != nil {
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
	p := NewProgram(m, WithInput(&in), WithOutput(&buf))
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				p.Kill()
				return
			}
		}
	}()

	if _, err := p.Run(); err != ErrProgramKilled {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}
}

func TestTeaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m, WithContext(ctx), WithInput(&in), WithOutput(&buf))
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if m.executed.Load() != nil {
				cancel()
				return
			}
		}
	}()

	if _, err := p.Run(); err != ErrProgramKilled {
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
	p := NewProgram(m, WithInput(&in), WithOutput(&buf))
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
	p := NewProgram(m, WithInput(&in), WithOutput(&buf))
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
	p := NewProgram(m, WithInput(&in), WithOutput(&buf))
	go p.Send(sequenceMsg{batch, inc, Quit})

	if _, err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if m.counter.Load() != 3 {
		t.Fatalf("counter should be 3, got %d", m.counter.Load())
	}
}

func TestTeaSend(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	m := &testModel{}
	p := NewProgram(m, WithInput(&in), WithOutput(&buf))

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
	NewProgram(m, WithInput(&in), WithOutput(&buf))
}
