// +build !windows

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

// newCancelReader returns a reader and a cancel function. If the input reader
// is an *os.File, the cancel function can be used to interrupt a blocking call
// read call. In this case, the cancel function returns true if the call was
// cancelled successfully. If the input reader is not a *os.File, the cancel
// function does nothing and always returns false.
func newCancelReader(reader io.Reader) (*cancelReader, func() bool, error) {
	r := &cancelReader{
		file:           reader.(*os.File),
		fallbackReader: reader,
	}

	if r.file != nil {
		var err error

		r.cancelSignalReader, r.cancelSignalWriter, err = os.Pipe()
		if err != nil {
			return nil, nil, err
		}
	}

	return r, r.cancel, nil
}

type cancelReader struct {
	fallbackReader     io.Reader // only used when the reader is not a *os.File
	file               *os.File
	cancelSignalReader *os.File
	cancelSignalWriter *os.File
	cancelled          bool
	sync.Mutex
}

func (r *cancelReader) Read(data []byte) (int, error) {
	if r.cancelled {
		return 0, errCanceled
	}

	if r.file == nil {
		return r.fallbackReader.Read(data)
	}

	r.Lock()
	defer r.Unlock()
	for {
		err := waitForRead(r.file, r.cancelSignalReader)
		if err != nil {
			if errors.Is(err, unix.EINTR) && !r.cancelled {
				continue // try again if syscall was interrupted
			}

			if errors.Is(err, errCanceled) {
				// remove signal from pipe
				var b [1]byte
				_, _ = r.cancelSignalReader.Read(b[:])
				// close pipe
				_ = r.cancelSignalReader.Close()
				_ = r.cancelSignalWriter.Close()
			}

			return 0, err
		}

		return r.file.Read(data)
	}
}

func (r *cancelReader) cancel() bool {
	r.cancelled = true

	// if the underlying reader is not a *os.File, read calls cannot be
	// cancelled
	if r.file == nil {
		return false
	}

	// send cancel signal
	_, err := r.cancelSignalWriter.Write([]byte{'q'})
	if err != nil {
		return false
	}

	r.Lock()
	// don't return until Read call exited
	defer r.Unlock()

	return true
}

var errCanceled = fmt.Errorf("read cancelled")

func waitForRead(reader *os.File, abort *os.File) error {
	readerFd := int(reader.Fd())
	abortFd := int(abort.Fd())

	maxFd := readerFd
	if abortFd > maxFd {
		maxFd = abortFd
	}

	// this is a limitation of the select syscall
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
		return errCanceled
	}

	if fdSet.IsSet(readerFd) {
		return nil
	}

	return fmt.Errorf("select returned without setting a file descriptor")
}
