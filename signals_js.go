//go:build js && wasm

package tea

// listenForResize does nothing in a browser.
//
// On a terminal this waits for SIGWINCH and re-measures the tty. WebAssembly
// in a browser has neither: there is no signal to wait for and no file
// descriptor to measure. The size has to come from whatever is hosting the
// program — a canvas, a terminal emulator written in JavaScript, an element
// being resized — and only the host knows when that changed.
//
// Hosts supply it with [WithWindowSize] at start and [Program.Send] of a
// [WindowSizeMsg] afterwards, which is the same message a SIGWINCH would have
// produced. This goroutine therefore has nothing to do but exit cleanly.
func (p *Program) listenForResize(done chan struct{}) {
	defer close(done)
	<-p.ctx.Done()
}
