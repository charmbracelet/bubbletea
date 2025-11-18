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
}
