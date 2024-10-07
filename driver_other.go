//go:build !windows
// +build !windows

package tea

// ReadEvents reads input events from the terminal.
//
// It reads the events available in the input buffer and returns them.
func (d *driver) ReadEvents() ([]Msg, error) {
	return d.readEvents()
}
