package tea

import (
	"fmt"
	"io"
	"sync"
)

var errCanceled = fmt.Errorf("read cancelled")

// cancelReader is a io.Reader whose Read() calls can be cancelled without data
// being consumed. The cancelReader has to be closed.
type cancelReader interface {
	io.ReadCloser

	// Cancel cancels ongoing and future reads an returns true if it succeeded.
	Cancel() bool
}

// fallbackCancelReader implements cancelReader but does not actually support
// cancelation during an ongoing Read() call. Thus, Cancel() always returns
// false. However, after calling Cancel(), new Read() calls immediately return
// errCanceled and don't consume any data anymore.
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

// cancelMixin represents a goroutine-safe cancelation status.
type cancelMixin struct {
	unsafeCancelled bool
	lock            sync.Mutex
}

func (c *cancelMixin) isCancelled() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.unsafeCancelled
}

func (c *cancelMixin) setCancelled() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.unsafeCancelled = true
}
