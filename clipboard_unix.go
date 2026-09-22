//go:build !windows && !darwin
// +build !windows,!darwin

package tea

import (
	"os/exec"

	uv "github.com/charmbracelet/ultraviolet"
)

// newLocalClipboardBackend returns a clipboard backend backed by the system's
// clipboard tools. Wayland sessions use wl-clipboard, X11 sessions use xclip
// or xsel. It returns nil when no usable tool is found.
func newLocalClipboardBackend(environ uv.Environ) ClipboardBackend {
	if environ.Getenv("WAYLAND_DISPLAY") != "" {
		if backend := waylandClipboardBackend(); backend != nil {
			return backend
		}
	}
	if environ.Getenv("DISPLAY") != "" {
		return xClipboardBackend()
	}
	return nil
}

// waylandClipboardBackend returns a backend backed by wl-copy and wl-paste, or
// nil when they are not available.
func waylandClipboardBackend() ClipboardBackend {
	copyPath, err := exec.LookPath("wl-copy")
	if err != nil {
		return nil
	}
	pastePath, err := exec.LookPath("wl-paste")
	if err != nil {
		return nil
	}

	return commandClipboardBackend{
		set: func(selection ClipboardSelection) (string, []string) {
			if primarySelection(selection) {
				return copyPath, []string{"--primary"}
			}
			return copyPath, nil
		},
		get: func(selection ClipboardSelection) (string, []string) {
			// --no-newline avoids wl-paste appending a trailing newline,
			// which would corrupt the clipboard contents.
			args := []string{"--no-newline"}
			if primarySelection(selection) {
				args = append(args, "--primary")
			}
			return pastePath, args
		},
	}
}

// xClipboardBackend returns a backend backed by xclip or xsel, or nil when
// neither is available.
func xClipboardBackend() ClipboardBackend {
	if xclip, err := exec.LookPath("xclip"); err == nil {
		return commandClipboardBackend{
			set: func(selection ClipboardSelection) (string, []string) {
				return xclip, []string{"-selection", xSelection(selection)}
			},
			get: func(selection ClipboardSelection) (string, []string) {
				return xclip, []string{"-selection", xSelection(selection), "-o"}
			},
		}
	}

	if xsel, err := exec.LookPath("xsel"); err == nil {
		return commandClipboardBackend{
			set: func(selection ClipboardSelection) (string, []string) {
				return xsel, []string{"--" + xSelection(selection), "--input"}
			},
			get: func(selection ClipboardSelection) (string, []string) {
				return xsel, []string{"--" + xSelection(selection), "--output"}
			},
		}
	}

	return nil
}

// xSelection returns the X11 selection name for the given selection.
func xSelection(selection ClipboardSelection) string {
	if primarySelection(selection) {
		return "primary"
	}
	return "clipboard"
}

// primarySelection reports whether selection refers to the primary (X11 and
// Wayland only) clipboard.
func primarySelection(selection ClipboardSelection) bool {
	return selection == uv.PrimaryClipboard
}
