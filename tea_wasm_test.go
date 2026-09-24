//go:build wasip1 || (js && wasm)

package tea

import (
	"bytes"
	"testing"
)

// On WASM targets there is no TTY to open: a program using the default input
// must fall back to the runtime-managed stdin instead of failing at startup.
func TestRunDefaultInputWASM(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgram(&testModel{}, WithOutput(&buf), WithWindowSize(80, 24))
	go p.Send(Quit())
	if _, err := p.Run(); err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}
}
