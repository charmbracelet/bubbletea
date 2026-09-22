//go:build windows
// +build windows

package tea

import (
	"os/exec"

	uv "github.com/charmbracelet/ultraviolet"
)

// newLocalClipboardBackend returns a clipboard backend backed by the Windows
// clip.exe tool, or nil when it is not available.
//
// Windows Terminal supports OSC52 natively, so this is only used on terminals
// without OSC52 support. clip.exe can only write, so reads report
// [ErrClipboardUnavailable].
func newLocalClipboardBackend(uv.Environ) ClipboardBackend {
	clip, err := exec.LookPath("clip")
	if err != nil {
		return nil
	}
	return commandClipboardBackend{
		set: func(ClipboardSelection) (string, []string) { return clip, nil },
		get: func(ClipboardSelection) (string, []string) { return "", nil },
	}
}
