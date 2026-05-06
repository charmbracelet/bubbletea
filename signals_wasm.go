//go:build js && wasm

package tea

// listenForResize is not available on WASM because window resize events
// come through JavaScript and are handled by the embedding framework (e.g., booba).
func (p *Program) listenForResize(done chan struct{}) {
	close(done)
}
