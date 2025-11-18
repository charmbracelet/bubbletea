//go:build windows
// +build windows

package tea

import "github.com/charmbracelet/x/term"

func (p *Program) checkOptimizedMovements(*term.State) {
	p.useHardTabs = true
	p.useBackspace = true
}
