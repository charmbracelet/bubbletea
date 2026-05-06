//go:build !windows && !darwin && !dragonfly && !freebsd && !linux && !solaris && !aix && !js
// +build !windows,!darwin,!dragonfly,!freebsd,!linux,!solaris,!aix,!js

package tea

import "github.com/charmbracelet/x/term"

func (*Program) checkOptimizedMovements(*term.State) {}
