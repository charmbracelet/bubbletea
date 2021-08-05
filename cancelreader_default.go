// +build !linux,!darwin,!windows

package tea

import (
	"io"
)

// newCancelReader returns a reader that can NOT be canceled on Windows. The
// cancel function will always return false.
func newCancelReader(reader io.Reader) (io.Reader, func() bool, error) {
	return newFallbackCancelReader(reader)
}
