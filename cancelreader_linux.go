// +build linux

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
// function does nothing and always returns false. The linux implementation is
// based on the epoll mechanism.
func newCancelReader(reader io.Reader) (io.Reader, func() bool, error) {
	file, ok := reader.(*os.File)
	if !ok {
		return newFallbackCancelReader(reader)
	}

	epoll, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, nil, fmt.Errorf("create epoll: %w", err)
	}

	r := &cancelReader{
		file:  file,
		epoll: epoll,
	}

	r.cancelSignalReader, r.cancelSignalWriter, err = os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	err = unix.EpollCtl(epoll, unix.EPOLL_CTL_ADD, int(file.Fd()), &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(file.Fd()),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("add reader to epoll interrest list")
	}

	err = unix.EpollCtl(epoll, unix.EPOLL_CTL_ADD, int(r.cancelSignalReader.Fd()), &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(r.cancelSignalReader.Fd()),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("add reader to epoll interrest list")
	}

	return r, r.cancel, nil
}

type cancelReader struct {
	file               *os.File
	cancelSignalReader *os.File
	cancelSignalWriter *os.File
	cancelled          bool
	epoll              int
	sync.Mutex
}

func (r *cancelReader) Read(data []byte) (int, error) {
	if r.cancelled {
		return 0, errCanceled
	}

	r.Lock()
	defer r.Unlock()

	err := r.wait()
	if err != nil {
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

func (r *cancelReader) cancel() bool {
	r.cancelled = true

	// send cancel signal
	_, err := r.cancelSignalWriter.Write([]byte{'c'})
	if err != nil {
		return false
	}

	r.Lock()
	// don't return until Read call exited
	defer r.Unlock()

	return true
}

func (r *cancelReader) wait() error {
	events := make([]unix.EpollEvent, 1)
	n, err := unix.EpollWait(r.epoll, events, -1)
	if err != nil {
		return fmt.Errorf("kevent: %w", err)
	}

	for i := 0; i < n; i++ {
		switch events[i].Fd {
		case int32(r.file.Fd()):
			return nil
		case int32(r.cancelSignalReader.Fd()):
			return errCanceled
		}
	}

	return fmt.Errorf("unknown error")
}
