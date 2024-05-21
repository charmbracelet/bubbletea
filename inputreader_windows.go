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

func newInputReader(r io.Reader) (cancelreader.CancelReader, error) {
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

	originalMode, err := prepareConsole(conin,
		windows.ENABLE_MOUSE_INPUT,
		windows.ENABLE_WINDOW_INPUT,
		windows.ENABLE_EXTENDED_FLAGS,
	)
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

	return windows.CancelIo(r.conin) == nil
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
