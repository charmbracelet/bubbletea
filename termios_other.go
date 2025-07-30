//go:build !darwin && !dragonfly && !freebsd && !linux && !solaris && !aix
// +build !darwin,!dragonfly,!freebsd,!linux,!solaris,!aix

package tea

import "github.com/charmbracelet/x/term"

func (p *Program) checkOptimizedMovements(*term.State) {
	p.useHardTabs = true
	p.useBackspace = true
}
