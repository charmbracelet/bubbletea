package tea

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// ErrClipboardUnavailable is returned by [ClipboardBackend] implementations
// when no system clipboard is available.
var ErrClipboardUnavailable = errors.New("bubbletea: clipboard unavailable")

// ClipboardSelection represents a clipboard selection. The two selections
// Bubble Tea knows about are [uv.SystemClipboard] ('c') and
// [uv.PrimaryClipboard] ('p').
type ClipboardSelection = uv.ClipboardSelection

// ClipboardBackend performs clipboard operations outside of the terminal.
//
// Bubble Tea uses a backend automatically on terminals that are known not to
// support OSC52 clipboard sequences, such as Apple's Terminal.app, by shelling
// out to the operating system's clipboard tools (pbcopy and pbpaste on macOS,
// wl-clipboard, xclip, or xsel on Linux).
//
// Provide your own backend with [WithClipboardBackend], or disable the
// automatic fallback entirely with [WithoutClipboardFallback].
type ClipboardBackend interface {
	// Set writes content to the given clipboard selection.
	Set(selection ClipboardSelection, content string) error

	// Get reads the given clipboard selection. Implementations should
	// return [ErrClipboardUnavailable] when the selection cannot be read.
	Get(selection ClipboardSelection) (string, error)
}

// Versions from which terminals are known to support OSC52 clipboard
// sequences. They are intentionally conservative: a terminal that supports
// OSC52 but is bridged anyway still works, because the system clipboard is the
// same place the sequence would have written to.
const (
	// vteOSC52Version is the VTE version that added OSC52 support.
	vteOSC52Version = 7600
	// konsoleOSC52Version is the Konsole version that added OSC52 support.
	konsoleOSC52Version = 230400
)

// supportsOSC52 reports whether the terminal described by environ is expected
// to support OSC52 clipboard sequences.
func supportsOSC52(environ uv.Environ) bool {
	// Sessions reached over SSH use OSC52 to set the clipboard of the
	// terminal on the other end. Clipboard tools on the remote host would
	// target the wrong machine.
	if _, ok := environ.LookupEnv("SSH_TTY"); ok {
		return true
	}
	if _, ok := environ.LookupEnv("SSH_CONNECTION"); ok {
		return true
	}

	// Apple's Terminal.app does not support OSC52.
	if strings.EqualFold(environ.Getenv("TERM_PROGRAM"), "Apple_Terminal") {
		return false
	}

	// The Linux console has no clipboard at all.
	if term := environ.Getenv("TERM"); term == "dumb" || strings.HasPrefix(term, "linux") {
		return false
	}

	// VTE-based terminals only learned OSC52 in VTE 0.76.
	if vte, ok := environ.LookupEnv("VTE_VERSION"); ok {
		if v, err := strconv.Atoi(vte); err != nil || v < vteOSC52Version {
			return false
		}
	}

	// Konsole only learned OSC52 in 23.04.
	if konsole, ok := environ.LookupEnv("KONSOLE_VERSION"); ok {
		if v, err := strconv.Atoi(konsole); err != nil || v < konsoleOSC52Version {
			return false
		}
	}

	return true
}

// clipboardCommand returns the executable and arguments used for a clipboard
// operation on the given selection. An empty name means the operation is not
// supported.
type clipboardCommand func(selection ClipboardSelection) (name string, args []string)

// commandClipboardBackend is a [ClipboardBackend] backed by command line
// clipboard tools.
type commandClipboardBackend struct {
	set clipboardCommand
	get clipboardCommand
}

// Set writes content to the clipboard with the backend's set command.
func (b commandClipboardBackend) Set(selection ClipboardSelection, content string) error {
	name, args := b.set(selection)
	if name == "" {
		return ErrClipboardUnavailable
	}

	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Stdin = strings.NewReader(content)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bubbletea: %s: %w", name, err)
	}
	return nil
}

// Get reads the clipboard with the backend's get command.
func (b commandClipboardBackend) Get(selection ClipboardSelection) (string, error) {
	name, args := b.get(selection)
	if name == "" {
		return "", ErrClipboardUnavailable
	}

	out, err := exec.CommandContext(context.Background(), name, args...).Output()
	if err != nil {
		return "", fmt.Errorf("bubbletea: %s: %w", name, err)
	}
	return string(out), nil
}
