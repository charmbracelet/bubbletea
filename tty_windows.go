//go:build windows
// +build windows

package tea

import (
	"fmt"

	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/windows"
)

func (p *Program) initInput() (err error) {
	// Save stdin state and enable VT input
	// We also need to enable VT
	// input here.
	if f, ok := p.input.(term.File); ok && term.IsTerminal(f.Fd()) {
		p.ttyInput = f
		p.previousTtyInputState, err = term.MakeRaw(p.ttyInput.Fd())
		if err != nil {
			return fmt.Errorf("error making terminal raw: %w", err)
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
	if f, ok := p.output.(term.File); ok && term.IsTerminal(f.Fd()) {
		p.ttyOutput = f
		p.previousOutputState, err = term.GetState(f.Fd())
		if err != nil {
			return fmt.Errorf("error getting terminal state: %w", err)
		}

		var mode uint32
		if err := windows.GetConsoleMode(windows.Handle(p.ttyOutput.Fd()), &mode); err != nil {
			return fmt.Errorf("error getting console mode: %w", err)
		}

		if err := windows.SetConsoleMode(windows.Handle(p.ttyOutput.Fd()),
			mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING|
				windows.DISABLE_NEWLINE_AUTO_RETURN); err != nil {
			return fmt.Errorf("error setting console mode: %w", err)
		}

		//nolint:godox
		// TODO: check if we can optimize cursor movements on Windows.
		p.checkOptimizedMovements(p.previousOutputState)
	}

	return //nolint:nakedret
}

const suspendSupported = false

func suspendProcess() {}
