//go:build !windows && !darwin && !dragonfly && !freebsd && !linux && !solaris && !aix && !js && !wasip1
// +build !windows,!darwin,!dragonfly,!freebsd,!linux,!solaris,!aix,!js,!wasip1

package tea

import "github.com/charmbracelet/x/term"

func (*Program) checkOptimizedMovements(*term.State) {}
