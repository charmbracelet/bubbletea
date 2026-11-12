//go:build js && wasm

package tea

// initInput does nothing in a browser.
//
// There is no terminal to put into raw mode: input arrives as whatever the
// host writes into the reader given to [WithInput], already decoded from
// keyboard events. A host that wants the equivalent of raw mode arranges it on
// its own side, because it is the only thing that can — it owns the keyboard
// handler.
func (p *Program) initInput() error { return nil }

// suspendSupported is false because there is no process to suspend. A browser
// tab cannot be sent to the background of a shell it does not have.
const suspendSupported = false

// suspendProcess is unreachable while suspendSupported is false, and is here so
// that the platform-independent code compiles.
func suspendProcess() {}
