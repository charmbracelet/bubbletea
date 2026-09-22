//go:build windows
// +build windows

package tea

import uv "github.com/charmbracelet/ultraviolet"

// newLocalClipboardBackend always returns nil on Windows: Windows Terminal
// supports OSC52 natively, and Windows has no clipboard tool that can read the
// clipboard without an additional dependency.
func newLocalClipboardBackend(uv.Environ) ClipboardBackend {
	return nil
}
