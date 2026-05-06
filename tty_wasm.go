//go:build js && wasm

package tea

// initInput sets up input handling for WASM.
// On WASM, input comes through custom readers (e.g., from booba's JavaScript bridge),
// so we don't need to interact with TTY directly.
func (p *Program) initInput() error {
	return nil
}

const suspendSupported = false

// suspendProcess is a no-op on WASM.
func suspendProcess() {}
