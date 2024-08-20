package tea

import "io"

// NilRenderer is a no-op renderer. It implements the Renderer interface but
// doesn't render anything to the terminal.
type NilRenderer struct{}

var _ Renderer = NilRenderer{}

// SetOutput implements the Renderer interface.
func (NilRenderer) SetOutput(io.Writer) {}

// Flush implements the Renderer interface.
func (NilRenderer) Flush() error { return nil }

// Close implements the Renderer interface.
func (NilRenderer) Close() error { return nil }

// Render implements the Renderer interface.
func (NilRenderer) Render(string) error { return nil }

// Repaint implements the Renderer interface.
func (NilRenderer) Repaint() {}

// ClearScreen implements the Renderer interface.
func (NilRenderer) ClearScreen() {}

// AltScreen implements the Renderer interface.
func (NilRenderer) AltScreen() bool { return false }

// EnterAltScreen implements the Renderer interface.
func (NilRenderer) EnterAltScreen() {}

// ExitAltScreen implements the Renderer interface.
func (NilRenderer) ExitAltScreen() {}

// CursorVisibility implements the Renderer interface.
func (NilRenderer) CursorVisibility() bool { return false }

// ShowCursor implements the Renderer interface.
func (NilRenderer) ShowCursor() {}

// HideCursor implements the Renderer interface.
func (NilRenderer) HideCursor() {}

// InsertAbove implements the Renderer interface.
func (NilRenderer) InsertAbove(string) error { return nil }

// Resize implements the Renderer interface.
func (NilRenderer) Resize(int, int) {}
