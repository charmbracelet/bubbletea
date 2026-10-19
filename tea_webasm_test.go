//go:build js || wasip1 || wasip2

package tea

import (
	"bytes"
	"os"
	"testing"
)

// On WASM targets there is no TTY to open: a program using the default input
// must fall back to the runtime-managed stdin instead of failing at startup.
func TestRunDefaultInputWASM(t *testing.T) {
	tests := []struct {
		name   string
		opts   []ProgramOption
		width  int
		height int
	}{
		{
			name:   "default dimensions",
			width:  defaultWidth,
			height: defaultHeight,
		},
		{
			name:   "environment dimensions",
			opts:   []ProgramOption{WithEnvironment([]string{"COLUMNS=132", "LINES=43"})},
			width:  132,
			height: 43,
		},
		{
			name:   "invalid environment dimensions",
			opts:   []ProgramOption{WithEnvironment([]string{"COLUMNS=invalid", "LINES=-1"})},
			width:  defaultWidth,
			height: defaultHeight,
		},
		{
			name: "explicit dimensions override environment",
			opts: []ProgramOption{
				WithEnvironment([]string{"COLUMNS=132", "LINES=43"}),
				WithWindowSize(120, 50),
			},
			width:  120,
			height: 50,
		},
		{
			name: "partial explicit dimensions",
			opts: []ProgramOption{
				WithEnvironment([]string{"COLUMNS=132", "LINES=43"}),
				WithWindowSize(120, 0),
			},
			width:  120,
			height: 43,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			m := &windowSizeModel{}
			opts := append([]ProgramOption{WithOutput(&buf), WithEnvironment([]string{})}, tt.opts...)
			p := NewProgram(m, opts...)
			if _, err := p.Run(); err != nil {
				t.Fatalf("Run() returned an error: %v", err)
			}
			if p.input != os.Stdin {
				t.Error("expected runtime-managed stdin")
			}
			if m.windowSize.Width != tt.width || m.windowSize.Height != tt.height {
				t.Errorf("expected WindowSizeMsg %dx%d, got %dx%d",
					tt.width, tt.height, m.windowSize.Width, m.windowSize.Height)
			}
		})
	}
}
