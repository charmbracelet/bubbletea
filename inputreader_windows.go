//go:build windows
// +build windows

package tea

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/x/term"
	"github.com/erikgeiser/coninput"
	"github.com/muesli/cancelreader"
	"golang.org/x/sys/windows"
)

type conInputReader struct {
	cancelMixin

	conin windows.Handle

	originalMode uint32
}

var _ cancelreader.CancelReader = &conInputReader{}

func newInputReader(r io.Reader, enableMouse bool) (cancelreader.CancelReader, error) {
	fallback := func(io.Reader) (cancelreader.CancelReader, error) {
		return cancelreader.NewReader(r)
	}
	if f, ok := r.(term.File); !ok || f.Fd() != os.Stdin.Fd() {
		return fallback(r)
	}

	conin, err := coninput.NewStdinHandle()
	if err != nil {
		return fallback(r)
	}

	modes := []uint32{
		windows.ENABLE_WINDOW_INPUT,
		windows.ENABLE_EXTENDED_FLAGS,
	}

	// Since we have options to enable mouse events, [WithMouseCellMotion],
	// [WithMouseAllMotion], and [EnableMouseCellMotion],
	// [EnableMouseAllMotion], and [DisableMouse], we need to check if the user
	// has enabled mouse events and add the appropriate mode accordingly.
	// Otherwise, mouse events will be enabled all the time.
	if enableMouse {
		modes = append(modes, windows.ENABLE_MOUSE_INPUT)
	}

	originalMode, err := prepareConsole(conin, modes...)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare console input: %w", err)
	}

	return &conInputReader{
		conin:        conin,
		originalMode: originalMode,
	}, nil
}

// Cancel implements cancelreader.CancelReader.
func (r *conInputReader) Cancel() bool {
	r.setCanceled()

	return windows.CancelIoEx(r.conin, nil) == nil || windows.CancelIo(r.conin) == nil
}

// Close implements cancelreader.CancelReader.
func (r *conInputReader) Close() error {
	if r.originalMode != 0 {
		err := windows.SetConsoleMode(r.conin, r.originalMode)
		if err != nil {
			return fmt.Errorf("reset console mode: %w", err)
		}
	}

	return nil
}

// Read implements cancelreader.CancelReader.
func (r *conInputReader) Read(_ []byte) (n int, err error) {
	if r.isCanceled() {
		err = cancelreader.ErrCanceled
	}
	return
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
