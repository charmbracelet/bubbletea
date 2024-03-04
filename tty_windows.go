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
	// We also need to enable VT
	// input here.
	if f, ok := p.input.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		p.ttyInput = f
		p.previousTtyInputState, err = term.MakeRaw(int(p.ttyInput.Fd()))
		if err != nil {
			return err
		}

		// Enable VT input
		var mode uint32
		if err := windows.GetConsoleMode(windows.Handle(p.ttyInput.Fd()), &mode); err != nil {
			return fmt.Errorf("error getting console mode: %w", err)
		}

		if err := windows.SetConsoleMode(windows.Handle(p.ttyInput.Fd()), mode|windows.ENABLE_VIRTUAL_TERMINAL_INPUT); err != nil {
			return fmt.Errorf("error setting console mode: %w", err)
		}
	}

	// Save output screen buffer state and enable VT processing.
	if f, ok := p.output.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		p.ttyOutput = f
		p.previousOutputState, err = term.GetState(int(f.Fd()))
		if err != nil {
			return err
		}

		var mode uint32
		if err := windows.GetConsoleMode(windows.Handle(p.ttyOutput.Fd()), &mode); err != nil {
			return fmt.Errorf("error getting console mode: %w", err)
		}

		if err := windows.SetConsoleMode(windows.Handle(p.ttyOutput.Fd()), mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
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
