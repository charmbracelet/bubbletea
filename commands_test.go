package tea

import (
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

func TestBatch(t *testing.T) {
	testMultipleCommands[BatchMsg](t, Batch)
}

func TestSequence(t *testing.T) {
	testMultipleCommands[sequenceMsg](t, Sequence)
}

func TestChain(t *testing.T) {
	t.Run("empty chain returns nil", func(t *testing.T) {
		cmd := Chain()
		if cmd != nil {
			t.Fatalf("expected nil, got %+v", cmd)
		}
	})

	t.Run("single step receives nil and returns msg", func(t *testing.T) {
		cmd := Chain(func(msg Msg) Cmd {
			if msg != nil {
				t.Fatalf("expected nil input, got %+v", msg)
			}
			return func() Msg { return "hello" }
		})
		result := cmd()
		if result != "hello" {
			t.Fatalf("expected 'hello', got %+v", result)
		}
	})

	t.Run("chains multiple steps passing msgs through", func(t *testing.T) {
		cmd := Chain(
			func(msg Msg) Cmd {
				return func() Msg { return 1 }
			},
			func(msg Msg) Cmd {
				n := msg.(int)
				return func() Msg { return n + 10 }
			},
			func(msg Msg) Cmd {
				n := msg.(int)
				return func() Msg { return n * 2 }
			},
		)
		result := cmd()
		if result != 22 {
			t.Fatalf("expected 22, got %+v", result)
		}
	})

	t.Run("nil cmd in middle passes nil to next step", func(t *testing.T) {
		cmd := Chain(
			func(msg Msg) Cmd {
				return func() Msg { return "first" }
			},
			func(msg Msg) Cmd {
				return nil // nil command
			},
			func(msg Msg) Cmd {
				if msg != nil {
					t.Fatalf("expected nil after nil cmd, got %+v", msg)
				}
				return func() Msg { return "recovered" }
			},
		)
		result := cmd()
		if result != "recovered" {
			t.Fatalf("expected 'recovered', got %+v", result)
		}
	})
}

func testMultipleCommands[T ~[]Cmd](t *testing.T, createFn func(cmd ...Cmd) Cmd) {
	t.Run("nil cmd", func(t *testing.T) {
		if b := createFn(nil); b != nil {
			t.Fatalf("expected nil, got %+v", b)
		}
	})
	t.Run("empty cmd", func(t *testing.T) {
		if b := createFn(); b != nil {
			t.Fatalf("expected nil, got %+v", b)
		}
	})
	t.Run("single cmd", func(t *testing.T) {
		b := createFn(Quit)()
		if _, ok := b.(QuitMsg); !ok {
			t.Fatalf("expected a QuitMsg, got %T", b)
		}
	})
	t.Run("mixed nil cmds", func(t *testing.T) {
		b := createFn(nil, Quit, nil, Quit, nil, nil)()
		if l := len(b.(T)); l != 2 {
			t.Fatalf("expected a []Cmd with len 2, got %d", l)
		}
	})
}
