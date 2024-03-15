//go:build windows
// +build windows

package tea

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/erikgeiser/coninput"
	"github.com/muesli/cancelreader"
	"golang.org/x/sys/windows"
)

type conInputReader struct {
	cancelMixin

	conin       windows.Handle
	cancelEvent windows.Handle

	originalMode uint32

	// inputEvent holds the input event that was read in order to avoid
	// unnecessary allocations. This re-use is possible because
	// InputRecord.Unwarp which is called inparseInputMsgFromInputRecord
	// returns an data structure that is independent of the passed InputRecord.
	inputEvent []coninput.InputRecord
}

var _ cancelreader.CancelReader = &conInputReader{}

func newInputReader(r io.Reader) (cancelreader.CancelReader, error) {
	fallback := func(io.Reader) (cancelreader.CancelReader, error) {
		return cancelreader.NewReader(r)
	}
	if f, ok := r.(*os.File); !ok || f.Fd() != os.Stdin.Fd() {
		return fallback(r)
	}

	conin, err := coninput.NewStdinHandle()
	if err != nil {
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
		conin:        conin,
		cancelEvent:  cancelEvent,
		originalMode: originalMode,
	}, nil
}

// Cancel implements cancelreader.CancelReader.
func (r *conInputReader) Cancel() bool {
	r.setCanceled()

	err := windows.SetEvent(r.cancelEvent)
	if err != nil {
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
func (*conInputReader) Read(_ []byte) (n int, err error) {
	return 0, nil
}

func prepareConsole(input windows.Handle, modes ...uint32) (originalMode uint32, err error) {
	err = windows.GetConsoleMode(input, &originalMode)
	if err != nil {
		return 0, fmt.Errorf("get console mode: %w", err)
	}

	newMode := coninput.AddInputModes(0, modes...)

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
