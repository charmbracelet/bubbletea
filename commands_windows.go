// +build windows

package boba

// OnResize is not supported on Windows at this time as Windows does not
// support the SIGWINCH signal.
func OnResize(newMsgFunc func() Msg) Cmd {
	return nil
}
