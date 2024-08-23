package tea

import "io"

// Renderer is the interface for Bubble Tea renderers.
type Renderer interface {
	// Close closes the renderer and flushes any remaining data.
	Close() error

	// Render renders a frame to the output.
	Render(string)

	// SetOutput sets the output for the renderer.
	SetOutput(io.Writer)

	// Flush flushes the renderer's buffer to the output.
	Flush() error

	// InsertAbove inserts lines above the current frame. This only works in
	// inline mode.
	InsertAbove(string) error

	// Resize sets the size of the terminal.
	Resize(w int, h int)

	// Request a full re-render. Note that this will not trigger a render
	// immediately. Rather, this method causes the next render to be a full
	// Repaint. Because of this, it's safe to call this method multiple times
	// in succession.
	Repaint()

	// ClearScreen clear the terminal screen. This should always have the same
	// behavior as the "clear" command which is equivalent to `CSI 2 J` and
	// `CSI H`.
	ClearScreen()

	// SetMode toggles a terminal mode such as bracketed paste, the altscreen,
	// and so on.
	//
	// The mode argument is an int consisting of the mode identifier. For
	// example, to set alt-screen mode, you would call SetMode(1049, true).
	SetMode(mode int, on bool)

	// Mode returns whether the render has a mode enabled. For example, to
	// check if alt-screen mode is enabled, you would call Mode(1049).
	Mode(mode int) bool
}

// repaintMsg forces a full repaint.
type repaintMsg struct{}

// Terminal modes used by SetMode and Mode in Bubble Tea.
const (
	graphemeClustering = 2027
	altScreenMode      = 1049
	hideCursor         = 25
)
