//go:build wasip2 || (!windows && !darwin && !dragonfly && !freebsd && !linux && !solaris && !aix)

package tea

import "github.com/charmbracelet/x/term"

func (*Program) checkOptimizedMovements(*term.State) {}
