//go:build wasip1

package tea

// initInput sets up input handling for WASI.
// On WASI, input comes through stdin/stdout managed by the runtime,
// so we don't need to interact with TTY directly.
func (p *Program) initInput() error {
	return nil
}

const suspendSupported = false

// suspendProcess is a no-op on WASI.
func suspendProcess() {}
