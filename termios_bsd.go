//go:build dragonfly || freebsd
// +build dragonfly freebsd

package tea

import (
	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

func (p *Program) checkOptimizedMovements(s *term.State) {
	p.useHardTabs = s.Oflag&unix.TABDLY == unix.TAB0
}
