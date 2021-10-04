//go:build !darwin && !windows && !linux && !solaris && !freebsd && !netbsd && !openbsd
// +build !darwin,!windows,!linux,!solaris,!freebsd,!netbsd,!openbsd

package tea

import (
	"io"
)

// newInputReader returns a allbackInputReader that satisfies the inputReader
// but does not actually support cancelation.
func newInputReader(reader io.Reader) (inputReader, error) {
	return newFallbackInputReader(reader)
}
