package tea

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestMediaKeyConstantsMatchUltraviolet(t *testing.T) {
	tests := []struct {
		name string
		got  rune
		want rune
	}{
		{"play", KeyMediaPlay, uv.KeyMediaPlay},
		{"pause", KeyMediaPause, uv.KeyMediaPause},
		{"play pause", KeyMediaPlayPause, uv.KeyMediaPlayPause},
		{"reverse", KeyMediaReverse, uv.KeyMediaReverse},
		{"stop", KeyMediaStop, uv.KeyMediaStop},
		{"fast forward", KeyMediaFastForward, uv.KeyMediaFastForward},
		{"rewind", KeyMediaRewind, uv.KeyMediaRewind},
		{"next", KeyMediaNext, uv.KeyMediaNext},
		{"previous", KeyMediaPrev, uv.KeyMediaPrev},
		{"record", KeyMediaRecord, uv.KeyMediaRecord},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("Bubble Tea key code %d does not match Ultraviolet key code %d", tt.got, tt.want)
			}
		})
	}
}
