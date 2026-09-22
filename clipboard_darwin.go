//go:build darwin
// +build darwin

package tea

import (
	"os/exec"

	uv "github.com/charmbracelet/ultraviolet"
)

// newLocalClipboardBackend returns a clipboard backend backed by the macOS
// pbcopy and pbpaste tools, or nil when they are not available.
//
// macOS only has a single general pasteboard, so the primary selection maps to
// it as well.
func newLocalClipboardBackend(uv.Environ) ClipboardBackend {
	pbcopy, err := exec.LookPath("pbcopy")
	if err != nil {
		return nil
	}
	pbpaste, err := exec.LookPath("pbpaste")
	if err != nil {
		return nil
	}

	return commandClipboardBackend{
		set: func(ClipboardSelection) (string, []string) { return pbcopy, nil },
		get: func(ClipboardSelection) (string, []string) { return pbpaste, nil },
	}
}
