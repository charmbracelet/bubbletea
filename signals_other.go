//go:build wasip1 || wasip2

package tea

// listenForResize on WASI: no SIGWINCH; report immediately done.
func (p *Program) listenForResize(done chan struct{}) { close(done) }
