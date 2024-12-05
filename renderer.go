package tea

import "io"

// renderer is the interface for Bubble Tea renderers.
type renderer interface {
	// close closes the renderer and flushes any remaining data.
	close() error

	// render renders a frame to the output.
	render(string)

	// flush flushes the renderer's buffer to the output.
	flush() error

	// reset resets the renderer's state to its initial state.
	reset()

	// update updates the renderer's state with the given message. It returns a
	// [tea.Cmd] that can be used to send messages back to the program.
	update(Msg)
}

// repaintMsg forces a full repaint.
type repaintMsg struct{}

// rendererWriter is an internal message used to set the output of the renderer.
type rendererWriter struct {
	io.Writer
}
