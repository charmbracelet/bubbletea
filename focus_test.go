package tea

import (
	"testing"
)

func TestFocus(t *testing.T) {
	var p inputParser
	_, e := p.parseSequence([]byte("\x1b[I"))
	switch e.(type) {
	case FocusMsg:
		// ok
	default:
		t.Error("invalid sequence")
	}
}

func TestBlur(t *testing.T) {
	var p inputParser
	_, e := p.parseSequence([]byte("\x1b[O"))
	switch e.(type) {
	case BlurMsg:
		// ok
	default:
		t.Error("invalid sequence")
	}
}
