//go:build wasip1

package tea

// listenForResize is not available on WASI because the runtime manages
// terminal size events and provides them through stdin.
func (p *Program) listenForResize(done chan struct{}) {
	close(done)
}
