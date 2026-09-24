//go:build wasip1

package tea

import "github.com/charmbracelet/x/term"

// checkOptimizedMovements is a no-op on WASI.
// Terminal optimization flags are determined by the runtime, not the program.
func (*Program) checkOptimizedMovements(*term.State) {}
