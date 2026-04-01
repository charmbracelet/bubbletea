//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos

package tea

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// handleSIGQUIT listens for SIGQUIT and writes a goroutine dump to the given
// writer. This restores the default Go runtime behavior of dumping all
// goroutines on SIGQUIT, which is normally suppressed when Bubble Tea captures
// signals.
func (p *Program) handleSIGQUIT(w *os.File) chan struct{} {
	ch := make(chan struct{})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGQUIT)
		defer func() {
			signal.Stop(sig)
			close(ch)
		}()

		for {
			select {
			case <-p.ctx.Done():
				return
			case <-sig:
				buf := make([]byte, 1<<20) // 1 MB buffer
				n := runtime.Stack(buf, true)
				_, _ = w.Write(buf[:n])
			}
		}
	}()

	return ch
}
