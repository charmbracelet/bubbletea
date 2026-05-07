//go:build js && wasm

package tea

// listenForResize is not available on WASM because window resize events
// come through JavaScript and are handled by the JavaScript embedding framework.
// This function is a no-op and simply closes the done channel immediately.
func (p *Program) listenForResize(done chan struct{}) {
	close(done)
}
