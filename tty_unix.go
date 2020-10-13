// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

// enableAnsiColors is only needed for Windows, so for other systems this is
// a no-op.
func enableAnsiColors() {}
