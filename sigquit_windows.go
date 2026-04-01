//go:build windows

package tea

import "os"

// handleSIGQUIT is a no-op on Windows since SIGQUIT is not available.
func (p *Program) handleSIGQUIT(_ *os.File) chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
