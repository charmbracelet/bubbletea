package tea

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"
)

type testModel struct {
	executed atomic.Value
}

func (m testModel) Init() Cmd {
	return nil
}

func (m *testModel) Update(msg Msg) (Model, Cmd) {
	switch msg.(type) {
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
	if err := p.Start(); err != nil {
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

	if err := p.Start(); err != nil {
		t.Fatal(err)
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

	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
}
