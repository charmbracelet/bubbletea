//go:build !windows
// +build !windows

package tea

// ReadEvents reads input events from the terminal.
//
// It reads the events available in the input buffer and returns them.
func (d *driver) ReadEvents() ([]Msg, error) {
	return d.readEvents()
}

// parseWin32InputKeyEvent parses a Win32 input key events. This function is
// only available on Windows.
func (p *inputParser) parseWin32InputKeyEvent(*win32InputState, uint16, uint16, rune, bool, uint32, uint16) Msg {
	return nil
}
