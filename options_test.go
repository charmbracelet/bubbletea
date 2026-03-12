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

	t.Run("renderer", func(t *testing.T) {
		p := NewProgram(nil, WithoutRenderer())
		if !p.disableRenderer {
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
			t.Errorf("expected passed in external context, got default")
		}
	})

	t.Run("input options", func(t *testing.T) {
		exercise := func(t *testing.T, opt ProgramOption, fn func(*Program)) {
			p := NewProgram(nil, opt)
			fn(p)
		}

		t.Run("nil input", func(t *testing.T) {
			exercise(t, WithInput(nil), func(p *Program) {
				if !p.disableInput || p.input != nil {
					t.Errorf("expected input to be disabled, got %v", p.input)
				}
			})
		})

		t.Run("custom input", func(t *testing.T) {
			var b bytes.Buffer
			exercise(t, WithInput(&b), func(p *Program) {
				if p.input != &b {
					t.Errorf("expected input to be custom, got %v", p.input)
				}
			})
		})
	})

	t.Run("startup options", func(t *testing.T) {
		exercise := func(t *testing.T, opt ProgramOption, fn func(*Program)) {
			p := NewProgram(nil, opt)
			fn(p)
		}

		t.Run("without catch panics", func(t *testing.T) {
			exercise(t, WithoutCatchPanics(), func(p *Program) {
				if !p.disableCatchPanics {
					t.Errorf("expected catch panics to be disabled")
				}
			})
		})

		t.Run("without signal handler", func(t *testing.T) {
			exercise(t, WithoutSignalHandler(), func(p *Program) {
				if !p.disableSignalHandler {
					t.Errorf("expected signal handler to be disabled")
				}
			})
		})
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
