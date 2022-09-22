package tea

import (
	"bytes"
	"testing"
)

type testModel struct{}

func (m testModel) Init() Cmd {
	return nil
}

func (m testModel) Update(msg Msg) (Model, Cmd) {
	switch msg.(type) {
	case KeyMsg:
		return m, Quit
	}
	return m, nil
}

func (m testModel) View() string {
	return "success\n"
}

func TestTeaModel(t *testing.T) {
	var buf bytes.Buffer
	var in bytes.Buffer
	in.Write([]byte("q"))

	p := NewProgram(testModel{}, WithInput(&in), WithOutput(&buf))
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}

	if buf.Len() == 0 {
		t.Fatal("no output")
	}
}
