package tea

// ClipboardMsg is a clipboard read message event. This message is emitted when
// a terminal receives an OSC52 clipboard read message event.
type ClipboardMsg string

// String returns the string representation of the clipboard message.
func (e ClipboardMsg) String() string {
	return string(e)
}

// PrimaryClipboardMsg is a primary clipboard read message event. This message
// is emitted when a terminal receives an OSC52 primary clipboard read message
// event. Primary clipboard selection is a feature present in X11 and Wayland
// only.
type PrimaryClipboardMsg string

// String returns the string representation of the primary clipboard message.
func (e PrimaryClipboardMsg) String() string {
	return string(e)
}

// setClipboardMsg is an internal message used to set the system clipboard
// using OSC52.
type setClipboardMsg string

// SetClipboard produces a command that sets the system clipboard using OSC52.
// Note that OSC52 is not supported in all terminals.
func SetClipboard(s string) Cmd {
	return func() Msg {
		return setClipboardMsg(s)
	}
}

// readClipboardMsg is an internal message used to read the system clipboard
// using OSC52.
type readClipboardMsg struct{}

// ReadClipboard produces a command that reads the system clipboard using OSC52.
// Note that OSC52 is not supported in all terminals.
func ReadClipboard() Msg {
	return readClipboardMsg{}
}

// setPrimaryClipboardMsg is an internal message used to set the primary
// clipboard using OSC52.
type setPrimaryClipboardMsg string

// SetPrimaryClipboard produces a command that sets the primary clipboard using
// OSC52. Primary clipboard selection is a feature present in X11 and Wayland
// only.
// Note that OSC52 is not supported in all terminals.
func SetPrimaryClipboard(s string) Cmd {
	return func() Msg {
		return setPrimaryClipboardMsg(s)
	}
}

// readPrimaryClipboardMsg is an internal message used to read the primary
// clipboard using OSC52.
type readPrimaryClipboardMsg struct{}

// ReadPrimaryClipboard produces a command that reads the primary clipboard
// using OSC52. Primary clipboard selection is a feature present in X11 and
// Wayland only.
// Note that OSC52 is not supported in all terminals.
func ReadPrimaryClipboard() Msg {
	return readPrimaryClipboardMsg{}
}
