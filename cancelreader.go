// nolint:revive
package tea

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

func newCancelReader(reader io.Reader) (*ctxReader, func() bool, error) {
	r := &ctxReader{
		file:           reader.(*os.File),
		fallbackReader: reader,
	}

	if r.file != nil {
		var err error

		r.rStop, r.wStop, err = os.Pipe()
		if err != nil {
			return nil, nil, err
		}
	}

	return r, r.cancel, nil
}

type ctxReader struct {
	fallbackReader io.Reader
	file           *os.File
	rStop          *os.File
	wStop          *os.File
	cancelled      bool
	// A mutex that is held when Read is in process.
	mutex sync.Mutex
}

func (r *ctxReader) Read(data []byte) (int, error) {
	if r.cancelled {
		return 0, errAborted
	}

	if r.file == nil {
		return r.fallbackReader.Read(data)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	for {
		err := WaitForRead(r.file, r.rStop)
		if err != nil {
			if errors.Is(err, unix.EINTR) && !r.cancelled {
				continue
			}

			if errors.Is(err, errAborted) {
				var b []byte
				_, _ = r.rStop.Read(b[:])
				_ = r.rStop.Close()
				_ = r.wStop.Close()
			}

			return 0, err
		}

		return r.file.Read(data)
	}
}

func (r *ctxReader) cancel() bool {
	if r.file == nil {
		return false
	}

	r.cancelled = true
	_, _ = r.wStop.Write([]byte{'q'})

	r.mutex.Lock()
	// don't return until Read call exited
	defer r.mutex.Unlock()

	return true
}

var errAborted = fmt.Errorf("select aborted")

func WaitForRead(reader *os.File, abort *os.File) error {
	readerFd := int(reader.Fd())
	abortFd := int(abort.Fd())

	maxFd := readerFd
	if abortFd > maxFd {
		maxFd = abortFd
	}

	if maxFd >= 1024 {
		return fmt.Errorf("cannot select on file descriptor %d which is larger than 1024", maxFd)
	}

	fdSet := &unix.FdSet{}
	fdSet.Set(int(reader.Fd()))
	fdSet.Set(int(abort.Fd()))

	_, err := unix.Select(maxFd+1, fdSet, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("select: %w", err)
	}

	if fdSet.IsSet(abortFd) {
		return errAborted
	}

	if fdSet.IsSet(readerFd) {
		return nil
	}

	return fmt.Errorf("select returned without setting a file descriptor")
}
