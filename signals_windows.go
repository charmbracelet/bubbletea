//go:build windows
// +build windows

package tea

// listenForHangup is not available on windows because windows does not
// implement syscall.SIGHUP.
func (p *Program) listenForHangup(done chan struct{}) {
	close(done)
}

// listenForResize is not available on windows because windows does not
// implement syscall.SIGWINCH.
func (p *Program) listenForResize(done chan struct{}) {
	close(done)
}
