//go:build !darwin && !windows && !linux && !solaris && !freebsd && !netbsd && !openbsd && !dragonfly
// +build !darwin,!windows,!linux,!solaris,!freebsd,!netbsd,!openbsd,!dragonfly

package tea

import (
	"io"
)

// newCancelReader returns a fallbackCancelReader that satisfies the
// cancelReader but does not actually support cancelation.
func newCancelReader(reader io.Reader) (cancelReader, error) {
	return newFallbackCancelReader(reader)
}
