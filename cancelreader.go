package tea

import (
	"fmt"
	"io"
)

var errCanceled = fmt.Errorf("read cancelled")

type cancelReader interface {
	io.ReadCloser
	Cancel() bool
}

type fallbackCancelReader struct {
	r         io.Reader
	cancelled bool
}

// newFallbackCancelReader is a fallback for newCancelReader that cannot
// actually cancel an ongoing read but will immediately return on future reads
// if it has been cancelled.
func newFallbackCancelReader(reader io.Reader) (cancelReader, error) {
	return &fallbackCancelReader{r: reader}, nil
}

func (r *fallbackCancelReader) Read(data []byte) (int, error) {
	if r.cancelled {
		return 0, errCanceled
	}

	return r.r.Read(data)
}

func (r *fallbackCancelReader) Cancel() bool {
	r.cancelled = true

	return false
}

func (r *fallbackCancelReader) Close() error {
	return nil
}
