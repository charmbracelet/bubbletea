//go:build js
// +build js

package tea

// listenForResize sends messages (or errors) when the terminal resizes.
// Argument output should be the file descriptor for the terminal; usually
// os.Stdout.
func (p *Program) listenForResize(done chan struct{}) {
	// Do nothing on JS.
	close(done)
}
