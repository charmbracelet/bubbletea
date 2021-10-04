//go:build windows
// +build windows

package tea

import (
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf16"

	"github.com/erikgeiser/coninput"
	"golang.org/x/sys/windows"
)

// newInputReader returns a cancelable input reader. If the input reader is an
// *os.File, the cancel method can be used to interrupt a blocking call read
// call. In this case, the cancel method returns true if the call was cancelled
// successfully. If the input reader is not a *os.File with the same file
// descriptor as os.Stdin, the cancel function does nothing and always returns
// false. The Windows implementation is based on WaitForMultipleObject. If
// os.Stdin is not a pipe, the events are read as input records, otherwise they
// are parsed from bytes using overlapping reads from CONIN$.
func newInputReader(reader io.Reader) (inputReader, error) {
	if f, ok := reader.(*os.File); !ok || f.Fd() != os.Stdin.Fd() {
		return newFallbackInputReader(reader)
	}

	conin, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		return nil, fmt.Errorf("get std input handle: %w", err)
	}

	// If data was piped to the standard input, it does not emit events anymore.
	// We can detect this if the console mode cannot be set anymore, in this
	// case, we use the compatibility reader.
	var dummy uint32
	err = windows.GetConsoleMode(conin, &dummy)
	if err != nil {
		return newCompatibilityInputReader()
	}

	return newInputRecordReader(conin)
}

func newInputRecordReader(conin windows.Handle) (*winInputRecordReader, error) {
	originalConsoleMode, err := prepareConsole(conin,
		windows.ENABLE_MOUSE_INPUT,
		windows.ENABLE_WINDOW_INPUT,
		windows.ENABLE_EXTENDED_FLAGS,
	)
	if err != nil {
		return nil, fmt.Errorf("prepare console: %w", err)
	}

	cancelEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("create stop event: %w", err)
	}

	return &winInputRecordReader{
		conin:               conin,
		cancelEvent:         cancelEvent,
		originalConsoleMode: originalConsoleMode,
		inputEvent:          make([]coninput.InputRecord, 4),
	}, nil

}

type winInputRecordReader struct {
	conin       windows.Handle
	cancelEvent windows.Handle
	cancelMixin

	originalConsoleMode uint32

	// inputEvent holds the input event that was read in order to avoid
	// unneccessary allocations. This re-use is possible because
	// InputRecord.Unwarp which is called inparseInputMsgFromInputRecord returns
	// an data structure that is independent of the passed InputRecord.
	inputEvent []coninput.InputRecord
}

func (r *winInputRecordReader) ReadInput() ([]Msg, error) {
	if r.isCancelled() {
		return nil, errCanceled
	}

	err := waitForInput(r.conin, r.cancelEvent)
	if err != nil {
		return nil, err
	}

	if r.isCancelled() {
		return nil, errCanceled
	}

	n, err := coninput.ReadConsoleInput(r.conin, r.inputEvent)
	if err != nil {
		return nil, fmt.Errorf("ReadConsoleInput: %w", err)
	}

	return parseInputMsgsFromInputRecords(r.inputEvent[:n])
}

// Cancel cancels ongoing and future Read() calls and returns true if the
// cancelation of the ongoing Read() was successful.
func (r *winInputRecordReader) Cancel() bool {
	r.setCancelled()

	err := windows.SetEvent(r.cancelEvent)
	if err != nil {
		return false
	}

	return true
}

func (r *winInputRecordReader) Close() error {
	err := windows.CloseHandle(r.cancelEvent)
	if err != nil {
		return fmt.Errorf("closing cancel event handle: %w", err)
	}

	if r.originalConsoleMode != 0 {
		err := windows.SetConsoleMode(r.conin, r.originalConsoleMode)
		if err != nil {
			return fmt.Errorf("reset console mode: %w", err)
		}
	}

	return nil
}

func newCompatibilityInputReader() (*winCompatibilityInputReader, error) {
	conin, err := windows.CreateFile(
		&(utf16.Encode([]rune("CONIN$\x00"))[0]), windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_WRITE|windows.FILE_SHARE_READ, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
	if err != nil {
		return nil, fmt.Errorf("open CONIN$ in overlapped mode: %w", err)
	}

	// set the *preferred* console mode, if data was piped to stdin this is not
	// possible anymore, so we ignore errors
	originalConsoleMode, _ := prepareConsole(conin,
		windows.ENABLE_EXTENDED_FLAGS,
		windows.ENABLE_INSERT_MODE,
		windows.ENABLE_QUICK_EDIT_MODE,
		// ENABLE_VIRTUAL_TERMINAL_INPUT causes unreadable inputs that trigger
		// WaitForMultipleObjects but it's necessary to receive special keys and
		// mouse events.
		windows.ENABLE_VIRTUAL_TERMINAL_INPUT,
	)

	cancelEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("create stop event: %w", err)
	}

	// flush input, otherwise it can contain events which trigger
	// WaitForMultipleObjects but which ReadFile cannot read, resulting in an
	// un-cancelable read
	err = coninput.FlushConsoleInputBuffer(conin)
	if err != nil {
		return nil, fmt.Errorf("flush console input buffer: %w", err)
	}

	return &winCompatibilityInputReader{
		conin:               conin,
		cancelEvent:         cancelEvent,
		originalConsoleMode: originalConsoleMode,
		blockingReadSignal:  make(chan struct{}, 1),
	}, nil
}

type winCompatibilityInputReader struct {
	conin       windows.Handle
	cancelEvent windows.Handle
	cancelMixin

	originalConsoleMode uint32
	blockingReadSignal  chan struct{}
}

func (r *winCompatibilityInputReader) ReadInput() ([]Msg, error) {
	if r.isCancelled() {
		return nil, errCanceled
	}

	err := waitForInput(r.conin, r.cancelEvent)
	if err != nil {
		return nil, err
	}

	if r.isCancelled() {
		return nil, errCanceled
	}

	r.blockingReadSignal <- struct{}{}
	msg, err := parseInputMsgFromReader(overlappedReader(r.conin))
	<-r.blockingReadSignal
	if err != nil {
		return nil, fmt.Errorf("parse input message from overlapped reader: %w", err)
	}

	return []Msg{msg}, nil
}

// Cancel cancels ongoing and future ReadInput() calls and returns true if the
// cancelation of the ongoing ReadInput() was successful. On Windows Terminal,
// WaitForMultipleObjects sometimes immediately returns without input being
// available. In this case, graceful cancelation is not possible and Cancel()
// returns false.
func (r *winCompatibilityInputReader) Cancel() bool {
	r.setCancelled()

	select {
	case r.blockingReadSignal <- struct{}{}:
		err := windows.SetEvent(r.cancelEvent)
		if err != nil {
			return false
		}
		<-r.blockingReadSignal
	case <-time.After(50 * time.Millisecond):
		// Read() hangs in a GetOverlappedResult which is likely due to
		// WaitForMultipleObjects returning without input being available
		// so we cannot cancel this ongoing read.
		return false
	}

	return true
}

func (r *winCompatibilityInputReader) Close() error {
	err := windows.CloseHandle(r.cancelEvent)
	if err != nil {
		return fmt.Errorf("closing cancel event handle: %w", err)
	}

	if r.originalConsoleMode != 0 {
		err := windows.SetConsoleMode(r.conin, r.originalConsoleMode)
		if err != nil {
			return fmt.Errorf("reset console mode: %w", err)
		}
	}

	// this does not close os.Stdin, just the handle
	err = windows.Close(r.conin)
	if err != nil {
		return fmt.Errorf("closing overlapped CONIN$ handle")
	}

	return nil
}

func waitForInput(conin, cancel windows.Handle) error {
	event, err := windows.WaitForMultipleObjects([]windows.Handle{conin, cancel}, false, windows.INFINITE)
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
		return int(n), nil
	}

	return int(n), nil
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
