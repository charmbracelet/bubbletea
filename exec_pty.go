package tea

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/term"
	xpty "github.com/charmbracelet/x/xpty"
)

// ptyDrainTimeout bounds how long execBridged waits for a command's output to
// be drained before giving up and closing the pseudo-terminal. It is long
// enough to drain the kernel's pty buffer, and keeps the program from hanging
// when a command's descendants keep the pty open after it exits.
const ptyDrainTimeout = 500 * time.Millisecond

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

	// Close the slave end in the parent, so that the command is its only
	// holder. When the command exits, reads on the master report EOF after
	// draining its output, instead of blocking forever.
	if slave, ok := pt.(interface{ Slave() *os.File }); ok {
		_ = slave.Slave().Close()
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

	// Stop forwarding, then wait for the pumps to drain the command's output
	// before closing the pseudo-terminal. Output written just before the
	// command exited may still be buffered in the pty, and closing it too
	// early would discard it. Waiting also ensures the goroutines don't steal
	// input from the program once it resumes.
	stop()
	if cancelInput != nil {
		cancelInput()
	}
	drained := make(chan struct{})
	go func() {
		wg.Wait()
		close(drained)
	}()
	select {
	case <-drained:
	case <-time.After(ptyDrainTimeout):
	}
	_ = pt.Close()

	if waitErr != nil {
		return fmt.Errorf("bubbletea: command failed: %w", waitErr)
	}
	return nil
}
