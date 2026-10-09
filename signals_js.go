//go:build js
// +build js

package tea

// listenForResize is not available in js runtime.
func (p *Program) listenForResize(done chan struct{}) {
	close(done)
}
