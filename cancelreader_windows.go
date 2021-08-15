// +build windows

package tea

import (
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

var fileShareValidFlags uint32 = 0x00000007

// newCancelReader returns a reader and a cancel function. If the input reader
// is an *os.File with the same file descriptor as os.Stdin, the cancel function
// can be used to interrupt a blocking call read call. In this case, the cancel
// function returns true if the call was cancelled successfully. If the input
// reader is not a *os.File with the same file descriptor as os.Stdin, the
// cancel function does nothing and always returns false. The Windows
// implementation is based on WaitForMultipleObject with overlapping reads from
// CONIN$.
func newCancelReader(reader io.Reader) (cancelReader, error) {
	if f, ok := reader.(*os.File); !ok || f.Fd() != os.Stdin.Fd() {
		return newFallbackCancelReader(reader)
	}

	// it is neccessary to open CONIN$ (NOT windows.STD_INPUT_HANDLE) in
	// overlapped mode to be able to use it with WaitForMultipleObjects.
	conin, err := windows.CreateFile(
		&(utf16.Encode([]rune("CONIN$\x00"))[0]), windows.GENERIC_READ, fileShareValidFlags,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
	if err != nil {
		return nil, fmt.Errorf("open CONIN$ in overlapping mode: %w", err)
	}

	cancelEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("create stop event: %w", err)
	}

	return &winCancelReader{
		conin:        conin,
		cancelEvent:  cancelEvent,
		blockingRead: make(chan struct{}, 1),
	}, nil
}

type winCancelReader struct {
	conin       windows.Handle
	cancelEvent windows.Handle
	cancelMixin

	blockingRead chan struct{}
}

func (r *winCancelReader) Read(data []byte) (int, error) {
	if r.isCancelled() {
		return 0, errCanceled
	}

	err := r.wait()
	if err != nil {
		return 0, err
	}

	if r.isCancelled() {
		return 0, errCanceled
	}

	// windows.Read does not work on overlapping windows.Handles
	return r.readAsync(data)
}

// Cancel cancels ongoing and future Read() calls and returns true if the
// cancelation of the ongoing Read() was successful. On Windows Terminal,
// WaitForMultipleObjects sometimes immediately returns without input being
// available. In this case, graceful cancelation is not possible and Cancel()
// returns false.
func (r *winCancelReader) Cancel() bool {
	r.setCancelled()

	select {
	case r.blockingRead <- struct{}{}:
		err := windows.SetEvent(r.cancelEvent)
		if err != nil {
			return false
		}
		<-r.blockingRead
	case <-time.After(100 * time.Millisecond):
		// Read() hangs in a GetOverlappedResult which is likely due to
		// WaitForMultipleObjects returning without input being available
		// so we cannot cancel this ongoing read.
		return false
	}

	return true
}

func (r *winCancelReader) Close() error {
	err := windows.CloseHandle(r.cancelEvent)
	if err != nil {
		return fmt.Errorf("closing cancel event handle: %w", err)
	}

	err = windows.Close(r.conin)
	if err != nil {
		return fmt.Errorf("closing CONIN$")
	}

	return nil
}

func (r *winCancelReader) wait() error {
	event, err := windows.WaitForMultipleObjects([]windows.Handle{r.conin, r.cancelEvent}, false, windows.INFINITE)
	switch {
	case windows.WAIT_OBJECT_0 <= event && event < windows.WAIT_OBJECT_0+2:
		if event == windows.WAIT_OBJECT_0+1 {
			return errCanceled
		}

		if event == windows.WAIT_OBJECT_0 {
			return nil
		}

		return fmt.Errorf("unexpected wait object is ready: %d", event-windows.WAIT_OBJECT_0)
	case windows.WAIT_ABANDONED <= event && event < windows.WAIT_ABANDONED+2:
		return fmt.Errorf("abandoned")
	case event == uint32(windows.WAIT_TIMEOUT):
		return fmt.Errorf("timeout")
	case event == windows.WAIT_FAILED:
		return fmt.Errorf("failed")
	default:
		return fmt.Errorf("unexpected error: %w", error(err))
	}
}

// readAsync is neccessary to read from a windows.Handle in overlapping mode.
func (r *winCancelReader) readAsync(data []byte) (int, error) {
	hevent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return 0, fmt.Errorf("create event: %w", err)
	}

	overlapped := windows.Overlapped{
		HEvent: hevent,
	}

	var n uint32

	err = windows.ReadFile(r.conin, data, &n, &overlapped)
	if err != nil && err != windows.ERROR_IO_PENDING {
		return int(n), err
	}

	r.blockingRead <- struct{}{}
	err = windows.GetOverlappedResult(r.conin, &overlapped, &n, true)
	if err != nil {
		return int(n), nil
	}
	<-r.blockingRead

	return int(n), nil
}
