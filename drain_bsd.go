//go:build darwin || dragonfly || freebsd || netbsd || openbsd
// +build darwin dragonfly freebsd netbsd openbsd

package tea

import "golang.org/x/sys/unix"

// drainInput discards any pending input on the TTY. It is called during
// shutdown to remove unsolicited terminal responses (e.g. DECRPM replies to
// mode 2026/2027 queries) that arrived after the input reader was cancelled.
// Without this, those bytes are read by the user's shell after exit and
// printed as garbage characters.
func (p *Program) drainInput() {
	if p.ttyInput == nil {
		return
	}
	fd := int(p.ttyInput.Fd())
	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}} //nolint:gosec // tty fd never overflows int32

	// Responses can arrive in bursts, so flush, then poll, then flush
	// again until nothing more arrives within the timeout window.
	// FREAD (1) tells TIOCFLUSH to discard the read queue only.
	for {
		_ = unix.IoctlSetPointerInt(fd, unix.TIOCFLUSH, 1)

		n, err := unix.Poll(fds, drainTimeoutMs)
		if err != nil || n <= 0 {
			return
		}
	}
}
