//go:build js
// +build js

package tea

// listenForResize does nothing on js/wasm: there is no SIGWINCH. The host
// environment can drive terminal size through WindowSizeMsg instead.
func (p *Program) listenForResize(done chan struct{}) {
	close(done)
}
