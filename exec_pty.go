package tea

import (
	"fmt"
	"io"
	"os/exec"
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/term"
	xpty "github.com/charmbracelet/x/xpty"
)

// execBridged runs c on a pseudo-terminal and bridges its clipboard traffic
// through the program's clipboard backend. This is what makes copy and paste
// work in external programs running on terminals without OSC52 support, in the
// same way osc52pty does for a wrapped process.
//
// It must only be called while the program has released the terminal.
func (p *Program) execBridged(c *exec.Cmd) error {
	width, height := p.width, p.height
	if width <= 0 || height <= 0 {
		// Let the pty keep its default size.
		width, height = -1, -1
	}
	if p.ttyOutput != nil {
		if w, h, err := term.GetSize(p.ttyOutput.Fd()); err == nil {
			width, height = w, h
		}
	}

	pt, err := xpty.NewPty(width, height)
	if err != nil {
		return fmt.Errorf("bubbletea: could not create pty: %w", err)
	}
	defer pt.Close() //nolint:errcheck

	preparePtyCommand(c)
	if err := pt.Start(c); err != nil {
		return fmt.Errorf("bubbletea: could not start command on pty: %w", err)
	}

	var wg sync.WaitGroup

	done := make(chan struct{})
	var stopOnce sync.Once
	stop := func() { stopOnce.Do(func() { close(done) }) }
	defer stop()

	var cancelInput func()
	if p.input != nil {
		input, err := uv.NewCancelReader(p.input)
		if err != nil {
			return fmt.Errorf("bubbletea: could not create pty input reader: %w", err)
		}
		defer input.Close() //nolint:errcheck
		cancelInput = func() { _ = input.Cancel() }

		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = io.Copy(pt, input)
		}()
	}

	bridge := newOSC52Bridge(p.clipboard, p.output, pt)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(bridge, pt)
		_ = bridge.Flush()
	}()

	resize, stopResize := ptyResizeSignals()
	defer stopResize()
	go func() {
		for {
			select {
			case <-done:
				return
			case <-resize:
				if p.ttyOutput == nil {
					continue
				}
				if w, h, err := term.GetSize(p.ttyOutput.Fd()); err == nil {
					_ = pt.Resize(w, h)
				}
			}
		}
	}()

	waitErr := xpty.WaitProcess(p.ctx, c)

	// Stop forwarding and wait for the goroutines, so they don't steal input
	// from the program once it resumes.
	stop()
	if cancelInput != nil {
		cancelInput()
	}
	_ = pt.Close()
	wg.Wait()

	if waitErr != nil {
		return fmt.Errorf("bubbletea: command failed: %w", waitErr)
	}
	return nil
}
