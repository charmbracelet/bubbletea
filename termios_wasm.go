//go:build js && wasm

package tea

import "github.com/charmbracelet/x/term"

// checkOptimizedMovements is a no-op on WASM.
// Terminal optimization flags don't apply in the browser.
func (*Program) checkOptimizedMovements(*term.State) {}
