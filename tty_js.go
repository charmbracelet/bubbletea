//go:build js
// +build js

package tea

func (p *Program) initInput() (err error) {
	// No termios to place in raw mode on js/wasm; whatever reader the caller
	// supplied via WithInput is used as-is.
	return nil
}

const suspendSupported = false

func suspendProcess() {}
