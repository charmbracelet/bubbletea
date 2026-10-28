package tea

import "time"

// listenForResize sends messages (or errors) when the terminal resizes.
// Argument output should be the file descriptor for the terminal; usually
// os.Stdout.
func (p *Program) listenForResize(done chan struct{}) {
	for {
		time.Sleep(100 * time.Second)
	}
}
