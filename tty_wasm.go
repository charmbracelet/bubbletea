//go:build js || wasip1

package tea

// initInput is a no-op on WebAssembly. Input arrives through custom readers
// (such as a JavaScript bridge) or runtime-managed stdin, so there is no TTY
// to configure.
func (p *Program) initInput() error {
	return nil
}

const suspendSupported = false

// suspendProcess is a no-op on WebAssembly.
func suspendProcess() {}
