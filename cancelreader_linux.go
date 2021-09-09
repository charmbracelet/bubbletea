//go:build linux
// +build linux

// nolint:revive
package tea

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// newCancelReader returns a reader and a cancel function. If the input reader
// is an *os.File, the cancel function can be used to interrupt a blocking call
// read call. In this case, the cancel function returns true if the call was
// cancelled successfully. If the input reader is not a *os.File, the cancel
// function does nothing and always returns false. The linux implementation is
// based on the epoll mechanism.
func newCancelReader(reader io.Reader) (cancelReader, error) {
	file, ok := reader.(*os.File)
	if !ok {
		return newFallbackCancelReader(reader)
	}

	epoll, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("create epoll: %w", err)
	}

	r := &epollCancelReader{
		file:  file,
		epoll: epoll,
	}

	r.cancelSignalReader, r.cancelSignalWriter, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	err = unix.EpollCtl(epoll, unix.EPOLL_CTL_ADD, int(file.Fd()), &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(file.Fd()),
	})
	if err != nil {
		return nil, fmt.Errorf("add reader to epoll interrest list")
	}

	err = unix.EpollCtl(epoll, unix.EPOLL_CTL_ADD, int(r.cancelSignalReader.Fd()), &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(r.cancelSignalReader.Fd()),
	})
	if err != nil {
		return nil, fmt.Errorf("add reader to epoll interrest list")
	}

	return r, nil
}

type epollCancelReader struct {
	file               *os.File
	cancelSignalReader *os.File
	cancelSignalWriter *os.File
	cancelMixin
	epoll int
}

func (r *epollCancelReader) Read(data []byte) (int, error) {
	if r.isCancelled() {
		return 0, errCanceled
	}

	err := r.wait()
	if err != nil {
		if errors.Is(err, errCanceled) {
			// remove signal from pipe
			var b [1]byte
			_, readErr := r.cancelSignalReader.Read(b[:])
			if readErr != nil {
				return 0, fmt.Errorf("reading cancel signal: %w", readErr)
			}
		}

		return 0, err
	}

	return r.file.Read(data)
}

func (r *epollCancelReader) Cancel() bool {
	r.setCancelled()

	// send cancel signal
	_, err := r.cancelSignalWriter.Write([]byte{'c'})
	return err == nil
}

func (r *epollCancelReader) Close() error {
	var errMsgs []string

	// close kqueue
	err := unix.Close(r.epoll)
	if err != nil {
		errMsgs = append(errMsgs, fmt.Sprintf("closing epoll: %v", err))
	}

	// close pipe
	err = r.cancelSignalWriter.Close()
	if err != nil {
		errMsgs = append(errMsgs, fmt.Sprintf("closing cancel signal writer: %v", err))
	}

	err = r.cancelSignalReader.Close()
	if err != nil {
		errMsgs = append(errMsgs, fmt.Sprintf("closing cancel signal reader: %v", err))
	}

	if len(errMsgs) > 0 {
		return fmt.Errorf(strings.Join(errMsgs, ", "))
	}

	return nil
}

func (r *epollCancelReader) wait() error {
	events := make([]unix.EpollEvent, 1)

	for {
		_, err := unix.EpollWait(r.epoll, events, -1)
		if errors.Is(err, unix.EINTR) {
			continue // try again if the syscall was interrupted
		}

		if err != nil {
			return fmt.Errorf("kevent: %w", err)
		}

		break
	}

	switch events[0].Fd {
	case int32(r.file.Fd()):
		return nil
	case int32(r.cancelSignalReader.Fd()):
		return errCanceled
	}

	return fmt.Errorf("unknown error")
}
