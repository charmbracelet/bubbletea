//go:build windows
// +build windows

package tea

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	xwindows "github.com/charmbracelet/x/windows"
	"github.com/muesli/cancelreader"
	"golang.org/x/sys/windows"
)

type conInputReader struct {
	cancelMixin

	conin       windows.Handle
	cancelEvent windows.Handle

	originalMode uint32

	// blockingReadSignal is used to signal that a blocking read is in progress.
	blockingReadSignal chan struct{}
}

var _ cancelreader.CancelReader = &conInputReader{}

func newCancelreader(r io.Reader) (cancelreader.CancelReader, error) {
	fallback := func(io.Reader) (cancelreader.CancelReader, error) {
		return cancelreader.NewReader(r)
	}

	var dummy uint32
	if f, ok := r.(cancelreader.File); !ok || f.Fd() != os.Stdin.Fd() ||
		// If data was piped to the standard input, it does not emit events
		// anymore. We can detect this if the console mode cannot be set anymore,
		// in this case, we fallback to the default cancelreader implementation.
		windows.GetConsoleMode(windows.Handle(f.Fd()), &dummy) != nil {
		return fallback(r)
	}

	conin, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		return fallback(r)
	}

	// Discard any pending input events.
	if err := xwindows.FlushConsoleInputBuffer(conin); err != nil {
		return fallback(r)
	}

	originalMode, err := prepareConsole(conin,
		windows.ENABLE_MOUSE_INPUT,
		windows.ENABLE_WINDOW_INPUT,
		windows.ENABLE_EXTENDED_FLAGS,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare console input: %w", err)
	}

	cancelEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("create stop event: %w", err)
	}

	return &conInputReader{
		conin:              conin,
		cancelEvent:        cancelEvent,
		originalMode:       originalMode,
		blockingReadSignal: make(chan struct{}, 1),
	}, nil
}

// Cancel implements cancelreader.CancelReader.
func (r *conInputReader) Cancel() bool {
	r.setCanceled()

	select {
	case r.blockingReadSignal <- struct{}{}:
		err := windows.SetEvent(r.cancelEvent)
		if err != nil {
			return false
		}
		<-r.blockingReadSignal
	case <-time.After(100 * time.Millisecond):
		// Read() hangs in a GetOverlappedResult which is likely due to
		// WaitForMultipleObjects returning without input being available
		// so we cannot cancel this ongoing read.
		return false
	}

	return true
}

// Close implements cancelreader.CancelReader.
func (r *conInputReader) Close() error {
	err := windows.CloseHandle(r.cancelEvent)
	if err != nil {
		return fmt.Errorf("closing cancel event handle: %w", err)
	}

	if r.originalMode != 0 {
		err := windows.SetConsoleMode(r.conin, r.originalMode)
		if err != nil {
			return fmt.Errorf("reset console mode: %w", err)
		}
	}

	return nil
}

// Read implements cancelreader.CancelReader.
func (r *conInputReader) Read(data []byte) (n int, err error) {
	if r.isCanceled() {
		return 0, cancelreader.ErrCanceled
	}

	err = waitForInput(r.conin, r.cancelEvent)
	if err != nil {
		return 0, err
	}

	if r.isCanceled() {
		return 0, cancelreader.ErrCanceled
	}

	r.blockingReadSignal <- struct{}{}
	n, err = overlappedReader(r.conin).Read(data)
	<-r.blockingReadSignal

	return
}

func prepareConsole(input windows.Handle, modes ...uint32) (originalMode uint32, err error) {
	err = windows.GetConsoleMode(input, &originalMode)
	if err != nil {
		return 0, fmt.Errorf("get console mode: %w", err)
	}

	var newMode uint32
	for _, mode := range modes {
		newMode |= mode
	}

	err = windows.SetConsoleMode(input, newMode)
	if err != nil {
		return 0, fmt.Errorf("set console mode: %w", err)
	}

	return originalMode, nil
}

func waitForInput(conin, cancel windows.Handle) error {
	event, err := windows.WaitForMultipleObjects([]windows.Handle{conin, cancel}, false, windows.INFINITE)
	switch {
	case windows.WAIT_OBJECT_0 <= event && event < windows.WAIT_OBJECT_0+2:
		if event == windows.WAIT_OBJECT_0+1 {
			return cancelreader.ErrCanceled
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
		return fmt.Errorf("unexpected error: %w", err)
	}
}

// cancelMixin represents a goroutine-safe cancelation status.
type cancelMixin struct {
	unsafeCanceled bool
	lock           sync.Mutex
}

func (c *cancelMixin) setCanceled() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.unsafeCanceled = true
}

func (c *cancelMixin) isCanceled() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.unsafeCanceled
}

type overlappedReader windows.Handle

// Read performs an overlapping read fom a windows.Handle.
func (r overlappedReader) Read(data []byte) (int, error) {
	hevent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return 0, fmt.Errorf("create event: %w", err)
	}

	overlapped := windows.Overlapped{HEvent: hevent}

	var n uint32

	err = windows.ReadFile(windows.Handle(r), data, &n, &overlapped)
	if err != nil && err != windows.ERROR_IO_PENDING {
		return int(n), err
	}

	err = windows.GetOverlappedResult(windows.Handle(r), &overlapped, &n, true)
	if err != nil {
		return int(n), err
	}

	return int(n), nil
}
