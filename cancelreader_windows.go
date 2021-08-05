// +build windows

package tea

import (
	"io"
)

// TODO
func newCancelReader(reader io.Reader) (io.Reader, func() bool, error) {
	return newFallbackCancelReader(reader)
}
