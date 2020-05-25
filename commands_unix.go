// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package tea

import (
	"os"
	"os/signal"
	"syscall"
)

// OnResize is used to listen for window resizes. Use GetTerminalSize to get
// the windows dimensions. We don't fetch the window size with this command to
// avoid a potential performance hit making the necessary system calls, since
// this command could potentially run a lot. On that note, consider debouncing
// this function.
func OnResize(newMsgFunc func() Msg) Cmd {
	return func() Msg {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGWINCH)
		<-sig
		return newMsgFunc()
	}
}
