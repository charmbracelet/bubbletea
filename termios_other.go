//go:build !darwin && !dragonfly && !freebsd && !linux && !solaris && !aix
// +build !darwin,!dragonfly,!freebsd,!linux,!solaris,!aix

package tea

import "github.com/charmbracelet/x/term"

func (*Program) checkOptimizedMovements(*term.State) {}
