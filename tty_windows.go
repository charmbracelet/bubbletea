// +build windows

package boba

import "github.com/muesli/termenv"

func initTerminal() error {
	termenv.HideCursor()
	return nil
}

func restoreTerminal() {
	termenv.ShowCursor()
}
