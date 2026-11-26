//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix

package tea

// drainTimeoutMs is how long we wait for additional bytes to arrive after
// flushing the input buffer. Local terminals reply within microseconds; this
// budget exists so SSH round-trips don't slip through.
const drainTimeoutMs = 200
