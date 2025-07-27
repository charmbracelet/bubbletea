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

func TestWrap(t *testing.T) {
	t.Run("wrapped nil cmd", func(t *testing.T) {
		if w := Wrap(nil, 1); w != nil {
			t.Fatalf("expected nil, got %+v", w)
		}
	})
	t.Run("wrapped single cmd", func(t *testing.T) {
		if w := Wrap(Quit, 1); w != nil {
			if m, ok := w().(WrappedMsg); !ok {
				t.Fatalf("expected WrappedMsg, got %+v", m)
			} else if m.Id != 1 {
				t.Fatalf("expected WrappedMsg{Id:1}, got %+v", m.Id)
			} else if _, ok := m.Msg.(QuitMsg); !ok {
				t.Fatalf("expected WrappedMsg{Msg:QuitMsg}, got %+v", m.Msg)
			}
		} else {
			t.Fatal("expected non-nil")
		}
	})
	t.Run("wrapped batch cmd", func(t *testing.T) {
		if w := Wrap(Batch(Quit, Quit), 1); w != nil {
			switch b := w().(type) {
			case BatchMsg:
				if l := len(b); l != 2 {
					t.Fatalf("expected a []Cmd with length 2, got %d", l)
				} else {
					// check *each* of the inner commands...
					for i, c := range b {
						if m, ok := c().(WrappedMsg); !ok {
							t.Fatalf("expected WrappedMsg for %d, got %+v", i, m)
						} else if m.Id != 1 {
							t.Fatalf("expected WrappedMsg{Id:1} for %d, got %+v", i, m.Id)
						} else if _, ok := m.Msg.(QuitMsg); !ok {
							t.Fatalf("expected WrappedMsg{Msg:QuitMsg} for %d, got %+v", i, m.Msg)
						}
					}
				}
			default:
				t.Fatalf("expected BatchMsg, got %#v", b)
			}
		} else {
			t.Fatal("expected non-nil")
		}
	})
	t.Run("wrapped sequence cmd", func(t *testing.T) {
		if w := Wrap(Sequence(Quit, Quit), 1); w != nil {
			switch b := w().(type) {
			case sequenceMsg:
				if l := len(b); l != 2 {
					t.Fatalf("expected a []Cmd with length 2, got %d", l)
				} else {
					// check *each* of the inner commands...
					for i, c := range b {
						if m, ok := c().(WrappedMsg); !ok {
							t.Fatalf("expected WrappedMsg for %d, got %+v", i, m)
						} else if m.Id != 1 {
							t.Fatalf("expected WrappedMsg{Id:1} for %d, got %+v", i, m.Id)
						} else if _, ok := m.Msg.(QuitMsg); !ok {
							t.Fatalf("expected WrappedMsg{Msg:QuitMsg} for %d, got %+v", i, m.Msg)
						}
					}
				}
			default:
				t.Fatalf("expected sequenceMsg, got %#v", b)
			}
		} else {
			t.Fatal("expected non-nil")
		}
	})
}
