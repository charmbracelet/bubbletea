// +build !darwin,!windows,!linux,!solaris,!freebsd,!netbsd,!openbsd

package tea

import (
	"io"
)

// newCancelReader returns a reader that can NOT be canceled on Windows. The
// cancel function will always return false.
func newCancelReader(reader io.Reader) (cancelReader, error) {
	return newFallbackCancelReader(reader)
}
