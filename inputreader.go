package tea

import (
	"fmt"
	"io"
	"sync"
)

var errCanceled = fmt.Errorf("read cancelled")

// inputReader allows cancellable reads of input events. The inputReader has to
// be closed.
type inputReader interface {
	ReadInput() ([]Msg, error)
	Close() error

	// Cancel cancels ongoing and future reads an returns true if it succeeded.
	Cancel() bool
}

// fallbackInputReader implements inputReader but does not actually support
// cancelation during an ongoing ReadInput() call. Thus, Cancel() always returns
// false. However, after calling Cancel(), new ReadInput() calls immediately
// return errCanceled and don't consume any data anymore.
type fallbackInputReader struct {
	r         io.Reader
	cancelled bool
}

// newFallbackInputReader is a fallback for newInputReader that cannot actually
// cancel an ongoing read but will immediately return on future reads if it has
// been cancelled.
func newFallbackInputReader(reader io.Reader) (inputReader, error) {
	return &fallbackInputReader{r: reader}, nil
}

func (r *fallbackInputReader) ReadInput() ([]Msg, error) {
	if r.cancelled {
		return nil, errCanceled
	}

	msg, err := parseInputMsgFromReader(r.r)
	if err != nil {
		return nil, err
	}

	return []Msg{msg}, nil
}

func (r *fallbackInputReader) Cancel() bool {
	r.cancelled = true

	return false
}

func (r *fallbackInputReader) Close() error {
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
