package tea

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("output", func(t *testing.T) {
		var b bytes.Buffer
		p := NewProgram[*testModel](nil).WithOutput(&b)
		assert.Nil(t, p.output.TTY())
	})

	t.Run("custom input", func(t *testing.T) {
		var b bytes.Buffer
		p := NewProgram[*testModel](nil).WithInput(&b)
		assert.Equal(t, &b, p.input)
		assert.Equal(t, customInput, p.inputType)
	})

	t.Run("renderer", func(t *testing.T) {
		p := NewProgram[*testModel](nil).WithoutRenderer()
		assert.IsType(t, (*nilRenderer)(nil), p.renderer)
	})

	t.Run("without signals", func(t *testing.T) {
		p := NewProgram[*testModel](nil).WithoutSignals()
		assert.True(t, p.ignoreSignals)
	})

	t.Run("filter", func(t *testing.T) {
		p := NewProgram[*testModel](nil).WithFilter(func(_ *testModel, msg Msg) Msg { return msg })
		assert.NotNil(t, p.filter)
	})

	t.Run("input options", func(t *testing.T) {
		t.Run("tty input", func(t *testing.T) {
			p := NewProgram[*testModel](nil).WithInputTTY()
			assert.Equal(t, ttyInput, p.inputType)
		})

		t.Run("custom input", func(t *testing.T) {
			var b bytes.Buffer
			p := NewProgram[*testModel](nil).WithInput(&b)
			assert.Equal(t, customInput, p.inputType)
		})
	})

	t.Run("startup options", func(t *testing.T) {
		t.Run("alt screen", func(t *testing.T) {
			p := NewProgram[*testModel](nil).WithAltScreen()
			assert.True(t, p.startupOptions.has(withAltScreen))
		})

		t.Run("ansi compression", func(t *testing.T) {
			p := NewProgram[*testModel](nil).WithANSICompressor()
			assert.True(t, p.startupOptions.has(withANSICompressor))
		})

		t.Run("without catch panics", func(t *testing.T) {
			p := NewProgram[*testModel](nil).WithoutCatchPanics()
			assert.True(t, p.startupOptions.has(withoutCatchPanics))
		})

		t.Run("without signal handler", func(t *testing.T) {
			p := NewProgram[*testModel](nil).WithoutSignalHandler()
			assert.True(t, p.startupOptions.has(withoutSignalHandler))
		})

		t.Run("mouse cell motion", func(t *testing.T) {
			p := NewProgram[*testModel](nil).WithMouseAllMotion().WithMouseCellMotion()
			assert.True(t, p.startupOptions.has(withMouseCellMotion))
			assert.False(t, p.startupOptions.has(withMouseAllMotion))
		})

		t.Run("mouse all motion", func(t *testing.T) {
			p := NewProgram[*testModel](nil).WithMouseCellMotion().WithMouseAllMotion()
			assert.True(t, p.startupOptions.has(withMouseAllMotion))
			assert.False(t, p.startupOptions.has(withMouseCellMotion))
		})
	})

	t.Run("multiple", func(t *testing.T) {
		p := NewProgram[*testModel](nil).WithMouseAllMotion().WithAltScreen().WithInputTTY()
		for _, opt := range []startupOptions{withMouseAllMotion, withAltScreen} {
			assert.True(t, p.startupOptions.has(opt))
			assert.Equal(t, ttyInput, p.inputType)
		}
	})
}
