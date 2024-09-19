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
		if _, ok := b.(QuitMsg); !ok {
			t.Fatalf("expected a QuitMsg, got %T", b)
		}
	})
	t.Run("mixed nil cmds", func(t *testing.T) {
		b := Batch(nil, Quit, nil, Quit, nil, nil)()
		if l := len(b.(BatchMsg)); l != 2 {
			t.Fatalf("expected a []Cmd with len 2, got %d", l)
		}
	})
}
