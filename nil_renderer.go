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
func (NilRenderer) Render(string) {}

// Repaint implements the Renderer interface.
func (NilRenderer) Repaint() {}

// ClearScreen implements the Renderer interface.
func (NilRenderer) ClearScreen() {}

// InsertAbove implements the Renderer interface.
func (NilRenderer) InsertAbove(string) error { return nil }

// Resize implements the Renderer interface.
func (NilRenderer) Resize(int, int) {}

// SetMode implements the Renderer interface.
func (NilRenderer) SetMode(int, bool) {}

// Mode implements the Renderer interface.
func (NilRenderer) Mode(int) bool { return false }
