package tea

import (
	"bytes"
	"context"
	"os"
	"sync/atomic"
	"testing"
)

func TestOptions(t *testing.T) {
	t.Run("output", func(t *testing.T) {
		var b bytes.Buffer
		p := NewProgram(nil, WithOutput(&b))
		if f, ok := p.output.(*os.File); ok {
			t.Errorf("expected output to custom, got %v", f.Fd())
		}
	})

	t.Run("custom input", func(t *testing.T) {
		var b bytes.Buffer
		p := NewProgram(nil, WithInput(&b))
		if p.input != &b {
			t.Errorf("expected input to custom, got %v", p.input)
		}
		if p.inputType != customInput {
			t.Errorf("expected startup options to have custom input set, got %v", p.input)
		}
	})

	t.Run("renderer", func(t *testing.T) {
		p := NewProgram(nil, WithoutRenderer())
		switch p.renderer.(type) {
		case *nilRenderer:
			return
		default:
			t.Errorf("expected renderer to be a nilRenderer, got %v", p.renderer)
		}
	})

	t.Run("without signals", func(t *testing.T) {
		p := NewProgram(nil, WithoutSignals())
		if atomic.LoadUint32(&p.ignoreSignals) == 0 {
			t.Errorf("ignore signals should have been set")
		}
	})

	t.Run("filter", func(t *testing.T) {
		p := NewProgram(nil, WithFilter(func(_ Model, msg Msg) Msg { return msg }))
		if p.filter == nil {
			t.Errorf("expected filter to be set")
		}
	})

	t.Run("external context", func(t *testing.T) {
		extCtx, extCancel := context.WithCancel(context.Background())
		defer extCancel()

		p := NewProgram(nil, WithContext(extCtx))
		if p.externalCtx != extCtx || p.externalCtx == context.Background() {
			t.Errorf("expected passed in external context, got default (nil)")
		}
	})

	t.Run("input options", func(t *testing.T) {
		exercise := func(t *testing.T, opt ProgramOption, expect inputType) {
			p := NewProgram(nil, opt)
			if p.inputType != expect {
				t.Errorf("expected input type %s, got %s", expect, p.inputType)
			}
		}

		t.Run("tty input", func(t *testing.T) {
			exercise(t, WithInputTTY(), ttyInput)
		})

		t.Run("custom input", func(t *testing.T) {
			var b bytes.Buffer
			exercise(t, WithInput(&b), customInput)
		})
	})

	t.Run("startup options", func(t *testing.T) {
		exercise := func(t *testing.T, opt ProgramOption, expect startupOptions) {
			p := NewProgram(nil, opt)
			if !p.startupOptions.has(expect) {
				t.Errorf("expected startup options have %v, got %v", expect, p.startupOptions)
			}
		}

		t.Run("alt screen", func(t *testing.T) {
			exercise(t, WithAltScreen(), withAltScreen)
		})

		t.Run("bracketed paste disabled", func(t *testing.T) {
			exercise(t, WithoutBracketedPaste(), withoutBracketedPaste)
		})

		t.Run("ansi compression", func(t *testing.T) {
			exercise(t, WithANSICompressor(), withANSICompressor)
		})

		t.Run("without catch panics", func(t *testing.T) {
			exercise(t, WithoutCatchPanics(), withoutCatchPanics)
		})

		t.Run("without signal handler", func(t *testing.T) {
			exercise(t, WithoutSignalHandler(), withoutSignalHandler)
		})

		t.Run("mouse cell motion", func(t *testing.T) {
			p := NewProgram(nil, WithMouseAllMotion(), WithMouseCellMotion())
			if !p.startupOptions.has(withMouseCellMotion) {
				t.Errorf("expected startup options have %v, got %v", withMouseCellMotion, p.startupOptions)
			}
			if p.startupOptions.has(withMouseAllMotion) {
				t.Errorf("expected startup options not have %v, got %v", withMouseAllMotion, p.startupOptions)
			}
		})

		t.Run("mouse all motion", func(t *testing.T) {
			p := NewProgram(nil, WithMouseCellMotion(), WithMouseAllMotion())
			if !p.startupOptions.has(withMouseAllMotion) {
				t.Errorf("expected startup options have %v, got %v", withMouseAllMotion, p.startupOptions)
			}
			if p.startupOptions.has(withMouseCellMotion) {
				t.Errorf("expected startup options not have %v, got %v", withMouseCellMotion, p.startupOptions)
			}
		})
	})

	t.Run("multiple", func(t *testing.T) {
		p := NewProgram(nil, WithMouseAllMotion(), WithoutBracketedPaste(), WithAltScreen(), WithInputTTY())
		for _, opt := range []startupOptions{withMouseAllMotion, withoutBracketedPaste, withAltScreen} {
			if !p.startupOptions.has(opt) {
				t.Errorf("expected startup options have %v, got %v", opt, p.startupOptions)
			}
			if p.inputType != ttyInput {
				t.Errorf("expected input to be %v, got %v", opt, p.startupOptions)
			}
		}
	})

	t.Run("multiple filters", func(t *testing.T) {
		type multiIncrement struct {
			num int
		}
		type eventuallyIncrementMsg incrementMsg

		// This filter converts multiIncrement to a sequence of eventuallyIncrementMsg.
		a := func(m Model, msg Msg) Msg {
			if mul, ok := msg.(multiIncrement); ok {
				var cmds []Cmd
				for range mul.num {
					cmds = append(cmds, func() Msg {
						return eventuallyIncrementMsg{}
					})
				}
				return sequenceMsg(cmds)
			}
			return msg
		}

		// This filter converts eventuallyIncrementMsg into incrementMsg.
		// If loaded out of order, the c filter breaks.
		b := func(_ Model, msg Msg) Msg {
			if msg, ok := msg.(eventuallyIncrementMsg); ok {
				return incrementMsg(msg)
			}
			return msg
		}

		// This filter quits after 10 incrementMsg.
		// Requires the b filter to work.
		c := func(m Model, msg Msg) Msg {
			p := m.(*testModel)
			// Stop after 10 increments.
			if _, ok := msg.(incrementMsg); ok {
				if v := p.counter.Load(); v != nil && v.(int) >= 10 {
					return QuitMsg{}
				}
			}

			return msg
		}

		var (
			buf bytes.Buffer
			in  bytes.Buffer
			m   = &testModel{}
		)
		p := NewProgram(m,
			// The combination of filters a, b, and c in this test causes the test
			// to correctly quit at 10 increments.

			// Convert into multiple eventuallyIncrementMsg.
			WithAddedFilter(a),
			// Convert into incrementMsg.
			WithAddedFilter(b),
			// Quit when the number of messages reaches 10.
			WithAddedFilter(c),

			WithInput(&in),
			WithOutput(&buf))
		go p.Send(multiIncrement{num: 20})

		if _, err := p.Run(); err != nil {
			t.Fatal(err)
		}

		if m.counter.Load().(int) != 10 {
			t.Fatalf("counter should be 10, got %d", m.counter.Load())
		}
	})
}
