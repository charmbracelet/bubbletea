//go:build !darwin && !windows && !linux && !solaris && !freebsd && !netbsd && !openbsd
// +build !darwin,!windows,!linux,!solaris,!freebsd,!netbsd,!openbsd

package tea

import (
	"io"
)

// newCancelReader returns a fallbackCancelReader that satisfies the
// cancelReader but does not actually support cancelation.
func newCancelReader(reader io.Reader) (cancelReader, error) {
	return newFallbackCancelReader(reader)
}
