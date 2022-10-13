package tea

import (
	"bytes"
	"testing"
)

func TestOptions(t *testing.T) {
	t.Run("output", func(t *testing.T) {
		var b bytes.Buffer
		p := NewProgram(nil, WithOutput(&b))
		if p.output.TTY() != nil {
			t.Errorf("expected output to custom, got %v", p.output.TTY().Fd())
		}
	})

	t.Run("input", func(t *testing.T) {
		var b bytes.Buffer
		p := NewProgram(nil, WithInput(&b))
		if p.input != &b {
			t.Errorf("expected input to custom, got %v", p.input)
		}
		if p.startupOptions&withCustomInput == 0 {
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

	t.Run("filter", func(t *testing.T) {
		p := NewProgram(nil, WithFilter(func(_ Model, msg Msg) Msg { return msg }))
		if p.filter == nil {
			t.Errorf("expected filter to be set")
		}
	})

	t.Run("startup options", func(t *testing.T) {
		exercise := func(t *testing.T, opt ProgramOption, expect startupOptions) {
			p := NewProgram(nil, opt)
			if !p.startupOptions.has(expect) {
				t.Errorf("expected startup options have %v, got %v", expect, p.startupOptions)
			}
		}

		t.Run("input tty", func(t *testing.T) {
			exercise(t, WithInputTTY(), withInputTTY)
		})

		t.Run("alt screen", func(t *testing.T) {
			exercise(t, WithAltScreen(), withAltScreen)
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
		p := NewProgram(nil, WithMouseAllMotion(), WithAltScreen(), WithInputTTY())
		for _, opt := range []startupOptions{withMouseAllMotion, withAltScreen, withInputTTY} {
			if !p.startupOptions.has(opt) {
				t.Errorf("expected startup options have %v, got %v", opt, p.startupOptions)
			}
		}
	})
}
