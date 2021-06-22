package tea

import (
	"fmt"
	"testing"
	"time"
)

func TestEvery(t *testing.T) {
	expected := "every ms"
	msg := Every(time.Millisecond, func(t time.Time) Msg {
		return expected
	})()
	if expected != msg {
		t.Fatalf("expected a msg %v but got %v", expected, msg)
	}
}

func TestTick(t *testing.T) {
	expected := "tick"
	msg := Tick(time.Millisecond, func(t time.Time) Msg {
		return expected
	})()
	if expected != msg {
		t.Fatalf("expected a msg %v but got %v", expected, msg)
	}
}

func TestSequentially(t *testing.T) {
	expectedErrMsg := fmt.Errorf("some err")
	expectedStrMsg := "some msg"

	nilReturnCmd := func() Msg {
		return nil
	}

	tests := []struct {
		name     string
		cmds     []Cmd
		expected Msg
	}{
		{
			name:     "all nil",
			cmds:     []Cmd{nilReturnCmd, nilReturnCmd},
			expected: nil,
		},
		{
			name:     "null cmds",
			cmds:     []Cmd{nil, nil},
			expected: nil,
		},
		{
			name: "one error",
			cmds: []Cmd{
				nilReturnCmd,
				func() Msg {
					return expectedErrMsg
				},
				nilReturnCmd,
			},
			expected: expectedErrMsg,
		},
		{
			name: "some msg",
			cmds: []Cmd{
				nilReturnCmd,
				func() Msg {
					return expectedStrMsg
				},
				nilReturnCmd,
			},
			expected: expectedStrMsg,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if msg := Sequentially(test.cmds...)(); msg != test.expected {
				t.Fatalf("expected a msg %v but got %v", test.expected, msg)
			}
		})
	}
}

func TestBatch(t *testing.T) {
	t.Run("nil cmd", func(t *testing.T) {
		if b := Batch(nil); b != nil {
			t.Fatalf("expected nil, got %+v", b)
		}
	})
	t.Run("empty cmd", func(t *testing.T) {
		if b := Batch(); b != nil {
			t.Fatalf("expected nil, got %+v", b)
		}
	})
	t.Run("single cmd", func(t *testing.T) {
		b := Batch(Quit)()
		if l := len(b.(BatchMsg)); l != 1 {
			t.Fatalf("expected a []Cmd with len 1, got %d", l)
		}
	})
	t.Run("mixed nil cmds", func(t *testing.T) {
		b := Batch(nil, Quit, nil, Quit, nil, nil)()
		if l := len(b.(BatchMsg)); l != 2 {
			t.Fatalf("expected a []Cmd with len 2, got %d", l)
		}
	})
}
