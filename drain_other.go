//go:build !windows && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !aix
// +build !windows,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!aix

package tea

// drainInput is a no-op on platforms where we don't have a portable way to
// discard pending TTY input.
func (p *Program) drainInput() {}
