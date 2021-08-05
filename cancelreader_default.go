// +build windows

package tea

import (
	"fmt"
	"io"
)

var errCanceled = fmt.Errorf("read cancelled")

// newCancelReader returns a reader that can NOT be canceled on Windows. The
// cancel function will always return false.
func newCancelReader(reader io.Reader) (*cancelReader, func() bool, error) {
	return &cancelReader{reader}, func() bool { return false }, nil
}

type cancelReader struct {
	io.Reader
}
