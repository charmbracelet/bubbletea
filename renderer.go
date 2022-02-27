package tea

// Renderer is the interface for Bubble Tea renderers.
type Renderer interface {
	// Start the renderer.
	Start()

	// Stop the renderer, but render the final frame in the buffer, if any.
	Stop()

	// Stop the renderer without doing any final rendering.
	Kill()

	// Write a frame to the renderer. The renderer can write this data to
	// output at its discretion.
	Write(string)

	// Request a full re-render.
	Repaint()

	// Whether or not the alternate screen buffer is enabled.
	AltScreen() bool

	// Record internally that the alternate screen buffer is enabled. This
	// does not actually toggle the alternate screen buffer.
	SetAltScreen(bool)

	HandleMessages(Msg)
}
