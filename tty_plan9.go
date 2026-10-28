package tea

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/x/term"
)

func (p *Program) initInput() (err error) {
	// Check if input is a terminal
	if f, ok := p.input.(term.File); ok && term.IsTerminal(f.Fd()) {
		p.ttyInput = f
		p.previousTtyInputState, err = term.MakeRaw(p.ttyInput.Fd())
		if err != nil {
			return fmt.Errorf("error entering raw mode: %w", err)
		}
	}

	if f, ok := p.output.(term.File); ok && term.IsTerminal(f.Fd()) {
		p.ttyOutput = f
	}

	return nil
}

func openInputTTY() (*os.File, error) {
	f, err := os.Open("/dev/cons")
	if err != nil {
		return nil, fmt.Errorf("could not open a new TTY: %w", err)
	}
	return f, nil
}

const suspendSupported = false

// Send SIGTSTP to the entire process group.
func suspendProcess() {
	p := os.Getpid()
	ctl := filepath.Join("proc", fmt.Sprintf("%d", p), "ctl")
	if err := os.WriteFile(ctl, []byte("stop"), 0); err != nil {
		log.Printf("Write sto to %q: %v", ctl, err)
	}
}
