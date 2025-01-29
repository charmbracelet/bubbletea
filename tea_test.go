package tea

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

type incrementMsg struct{}

type testModel struct {
	executed atomic.Value
	counter  atomic.Value
}

func newTestModel() (*testModel, Cmd) {
	return &testModel{}, nil
}

func newTestProgram(in io.Reader, out io.Writer) *Program[*testModel] {
	p := Program[*testModel]{
		Init:   newTestModel,
		Update: (*testModel).Update,
		View:   (*testModel).View,
	}
	p.Input = in
	p.Output = out
	p.ForceInputTTY = true
	return &p
}

func (m *testModel) Update(msg Msg) (*testModel, Cmd) {
	switch msg.(type) {
	case incrementMsg:
		i := m.counter.Load()
		if i == nil {
			m.counter.Store(1)
		} else {
			m.counter.Store(i.(int) + 1)
		}

	case KeyPressMsg:
		return m, Quit
	}

	return m, nil
}

func (m *testModel) View() Frame {
	m.executed.Store(true)
	return NewFrame("success\n")
}

func TestTeaModel(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer
	in.Write([]byte("q"))

	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	p := newTestProgram(&in, &buf)
	go func() {
		<-ctx.Done()
		p.Quit()
	}()
	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if buf.Len() == 0 {
		t.Fatal("no output")
	}
}

func TestTeaQuit(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	p := newTestProgram(&in, &buf)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if p.Model.executed.Load() != nil {
				p.Quit()
				return
			}
		}
	}()

	if err := p.Run(); err != nil {
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

	shutdowns := uint32(0)
	p := newTestProgram(&in, &buf)
	p.Filter = func(_ *testModel, msg Msg) Msg {
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

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if shutdowns != preventCount {
		t.Errorf("Expected %d prevented shutdowns, got %d", preventCount, shutdowns)
	}
}

func TestTeaKill(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	p := newTestProgram(&in, &buf)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if p.Model.executed.Load() != nil {
				p.Kill()
				return
			}
		}
	}()

	if err := p.Run(); !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}
}

func TestTeaContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var buf bytes.Buffer
	var in bytes.Buffer

	p := newTestProgram(&in, &buf)
	go func() {
		<-ctx.Done()
		p.Kill()
	}()
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if p.Model.executed.Load() != nil {
				cancel()
				return
			}
		}
	}()

	if err := p.Run(); !errors.Is(err, ErrProgramKilled) {
		t.Fatalf("Expected %v, got %v", ErrProgramKilled, err)
	}
}

func TestTeaBatchMsg(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	inc := func() Msg {
		return incrementMsg{}
	}

	p := newTestProgram(&in, &buf)
	go func() {
		p.Send(BatchMsg{inc, inc})

		for {
			time.Sleep(time.Millisecond)
			i := p.Model.counter.Load()
			if i != nil && i.(int) >= 2 {
				p.Quit()
				return
			}
		}
	}()

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if p.Model.counter.Load() != 2 {
		t.Fatalf("counter should be 2, got %d", p.Model.counter.Load())
	}
}

func TestTeaSequenceMsg(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	inc := func() Msg {
		return incrementMsg{}
	}

	p := newTestProgram(&in, &buf)
	go p.Send(sequenceMsg{inc, inc, Quit})

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if p.Model.counter.Load() != 2 {
		t.Fatalf("counter should be 2, got %d", p.Model.counter.Load())
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

	p := newTestProgram(&in, &buf)
	go p.Send(sequenceMsg{batch, inc, Quit})

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	if p.Model.counter.Load() != 3 {
		t.Fatalf("counter should be 3, got %d", p.Model.counter.Load())
	}
}

func TestTeaSend(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	p := newTestProgram(&in, &buf)

	// sending before the program is started is a blocking operation
	go p.Send(Quit())

	if err := p.Run(); err != nil {
		t.Fatal(err)
	}

	// sending a message after program has quit is a no-op
	p.Send(Quit())
}

func TestTeaNoRun(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer

	_ = newTestProgram(&in, &buf)
}
