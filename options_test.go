package tea

import (
	"bytes"
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
}
