package tea

import (
	"bytes"
	"testing"
)

type windowSizeModel struct {
	windowSize WindowSizeMsg
	quit       bool
}

func (m *windowSizeModel) Init() Cmd {
	return nil
}

func (m *windowSizeModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case WindowSizeMsg:
		m.windowSize = msg
		m.quit = true
		return m, Quit
	}
	return m, nil
}

func (m *windowSizeModel) View() View {
	return NewView("")
}

func TestFallbackDimensions(t *testing.T) {
	t.Run("default fallback to 80x24", func(t *testing.T) {
		t.Setenv("COLUMNS", "")
		t.Setenv("LINES", "")

		p := NewProgram(nil)
		w, h := p.fallbackDimensions(0, 0)
		if w != defaultWidth || h != defaultHeight {
			t.Errorf("expected %dx%d, got %dx%d", defaultWidth, defaultHeight, w, h)
		}

		w, h = p.fallbackDimensions(-10, -5)
		if w != defaultWidth || h != defaultHeight {
			t.Errorf("expected %dx%d, got %dx%d", defaultWidth, defaultHeight, w, h)
		}
	})

	t.Run("preserve positive dimensions", func(t *testing.T) {
		p := NewProgram(nil)
		w, h := p.fallbackDimensions(120, 40)
		if w != 120 || h != 40 {
			t.Errorf("expected 120x40, got %dx%d", w, h)
		}
	})

	t.Run("mixed non-positive dimensions", func(t *testing.T) {
		t.Setenv("COLUMNS", "")
		t.Setenv("LINES", "")

		p := NewProgram(nil)
		w, h := p.fallbackDimensions(0, 50)
		if w != defaultWidth || h != 50 {
			t.Errorf("expected %dx50, got %dx%d", defaultWidth, w, h)
		}

		w, h = p.fallbackDimensions(100, 0)
		if w != 100 || h != defaultHeight {
			t.Errorf("expected 100x%d, got %dx%d", defaultHeight, w, h)
		}
	})

	t.Run("fallback to COLUMNS and LINES from os env", func(t *testing.T) {
		t.Setenv("COLUMNS", "132")
		t.Setenv("LINES", "43")

		p := NewProgram(nil)
		w, h := p.fallbackDimensions(0, 0)
		if w != 132 || h != 43 {
			t.Errorf("expected 132x43, got %dx%d", w, h)
		}
	})

	t.Run("fallback to COLUMNS and LINES from program environment", func(t *testing.T) {
		p := NewProgram(nil, WithEnvironment([]string{"COLUMNS=150", "LINES=60"}))
		w, h := p.fallbackDimensions(0, 0)
		if w != 150 || h != 60 {
			t.Errorf("expected 150x60, got %dx%d", w, h)
		}
	})

	t.Run("invalid or non-positive COLUMNS and LINES fall back to defaults", func(t *testing.T) {
		t.Setenv("COLUMNS", "invalid")
		t.Setenv("LINES", "-10")

		p := NewProgram(nil)
		w, h := p.fallbackDimensions(0, 0)
		if w != defaultWidth || h != defaultHeight {
			t.Errorf("expected %dx%d, got %dx%d", defaultWidth, defaultHeight, w, h)
		}

		t.Setenv("COLUMNS", "0")
		t.Setenv("LINES", "0")

		w, h = p.fallbackDimensions(0, 0)
		if w != defaultWidth || h != defaultHeight {
			t.Errorf("expected %dx%d, got %dx%d", defaultWidth, defaultHeight, w, h)
		}
	})
}

func TestProgramWindowSizeFallback(t *testing.T) {
	t.Run("falls back to 80x24 when terminal size is not set", func(t *testing.T) {
		t.Setenv("COLUMNS", "")
		t.Setenv("LINES", "")

		var buf bytes.Buffer
		var in bytes.Buffer
		m := &windowSizeModel{}

		p := NewProgram(m,
			WithInput(&in),
			WithOutput(&buf),
		)

		if _, err := p.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if m.windowSize.Width != defaultWidth || m.windowSize.Height != defaultHeight {
			t.Errorf("expected WindowSizeMsg %dx%d, got %dx%d",
				defaultWidth, defaultHeight, m.windowSize.Width, m.windowSize.Height)
		}
	})

	t.Run("falls back to COLUMNS and LINES when terminal size is not set", func(t *testing.T) {
		t.Setenv("COLUMNS", "110")
		t.Setenv("LINES", "35")

		var buf bytes.Buffer
		var in bytes.Buffer
		m := &windowSizeModel{}

		p := NewProgram(m,
			WithInput(&in),
			WithOutput(&buf),
		)

		if _, err := p.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if m.windowSize.Width != 110 || m.windowSize.Height != 35 {
			t.Errorf("expected WindowSizeMsg 110x35, got %dx%d",
				m.windowSize.Width, m.windowSize.Height)
		}
	})

	t.Run("preserves WithWindowSize when specified", func(t *testing.T) {
		var buf bytes.Buffer
		var in bytes.Buffer
		m := &windowSizeModel{}

		p := NewProgram(m,
			WithInput(&in),
			WithOutput(&buf),
			WithWindowSize(120, 50),
		)

		if _, err := p.Run(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if m.windowSize.Width != 120 || m.windowSize.Height != 50 {
			t.Errorf("expected WindowSizeMsg 120x50, got %dx%d",
				m.windowSize.Width, m.windowSize.Height)
		}
	})
}
