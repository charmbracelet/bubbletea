//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tea

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// preparePtyCommand makes the command a session leader with the pseudo-terminal
// as its controlling terminal. Without it the command would not receive
// SIGWINCH on resize, and the terminal's line discipline would not generate
// signals such as SIGINT for it.
func preparePtyCommand(c *exec.Cmd) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setsid = true
	c.SysProcAttr.Setctty = true
}

// ptyResizeSignals returns a channel that receives terminal resize signals,
// along with a function that stops listening for them.
func ptyResizeSignals() (<-chan os.Signal, func()) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)
	return sig, func() { signal.Stop(sig) }
}
