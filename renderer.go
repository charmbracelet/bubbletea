package tea

// Renderer is the interface for Bubble Tea renderers.
type Renderer interface {
	// Close closes the renderer and flushes any remaining data.
	Close() error

	// Write a frame to the renderer. The renderer can write this data to
	// output at its discretion.
	Write([]byte) (int, error)

	// WriteString a frame to the renderer. The renderer can WriteString this
	// data to output at its discretion.
	WriteString(string) (int, error)

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

	// ClearScreen clear the terminal screen.
	ClearScreen()

	// Whether or not the alternate screen buffer is enabled.
	AltScreen() bool
	// Enable the alternate screen buffer.
	EnterAltScreen()
	// Disable the alternate screen buffer.
	ExitAltScreen()

	// CursorVisibility returns whether the cursor is visible.
	CursorVisibility() bool
	// Show the cursor.
	ShowCursor()
	// Hide the cursor.
	HideCursor()

	// Execute writes a sequence to the underlying output.
	Execute(string)
}

// repaintMsg forces a full repaint.
type repaintMsg struct{}
