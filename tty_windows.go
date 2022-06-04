//go:build windows
// +build windows

package tea

import (
	"os"

	"github.com/containerd/console"
)

func (p *Program) initInput() error {
	// If input's a file, use console to manage it
	if f, ok := p.input.(*os.File); ok {
		// Save a reference to the current stdin then replace stdin with our
		// input. We do this so we can hand input off to containerd/console to
		// set raw mode, and do it in this fashion because the method
		// console.ConsoleFromFile isn't supported on Windows.
		p.windowsStdin = os.Stdin
		os.Stdin = f

		// Note: this will panic if it fails.
		c := console.Current()
		p.console = c
	}

	return nil
}

// restoreInput restores stdout in the event that we placed it aside to handle
// input with CONIN$, above.
func (p *Program) restoreInput() error {
	if p.windowsStdin != nil {
		os.Stdin = p.windowsStdin
	}

	return nil
}

// Open the Windows equivalent of a TTY.
func openInputTTY() (*os.File, error) {
	f, err := os.OpenFile("CONIN$", os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}
