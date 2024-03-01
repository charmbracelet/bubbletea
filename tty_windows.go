//go:build windows
// +build windows

package tea

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

func (p *Program) initInput() (err error) {
	// Save stdin state and enable VT input
	// We enable VT processing using Termenv, but we also need to enable VT
	// input here.
	if f, ok := p.input.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		p.tty = f
		p.previousTtyState, err = term.MakeRaw(int(p.tty.Fd()))
		if err != nil {
			return err
		}

		// Enable VT input
		var mode uint32
		if err := windows.GetConsoleMode(windows.Handle(p.tty.Fd()), &mode); err != nil {
			return fmt.Errorf("error getting console mode: %w", err)
		}

		if err := windows.SetConsoleMode(windows.Handle(p.tty.Fd()), mode|windows.ENABLE_VIRTUAL_TERMINAL_INPUT); err != nil {
			return fmt.Errorf("error setting console mode: %w", err)
		}
	}

	return
}

// Open the Windows equivalent of a TTY.
func openInputTTY() (*os.File, error) {
	f, err := os.OpenFile("CONIN$", os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}
