// +build windows

package tea

func initTerminal() error {
	hideCursor()
	return nil
}

func restoreTerminal() {
	showCursor()
}
