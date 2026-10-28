//go:build js || wasip1

package tea

// listenForResize is a no-op on WebAssembly. In the browser, resize events
// come through JavaScript and are delivered by the embedding framework; on
// WASI, the runtime manages terminal size. Either way there is no signal to
// listen for, so the done channel is closed immediately.
func (p *Program) listenForResize(done chan struct{}) {
	close(done)
}
