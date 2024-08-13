package tea

// ClipboardMsg is a clipboard read message event.
// This message is emitted when a terminal receives an OSC52 clipboard read
// message event.
type ClipboardMsg string

// String returns the string representation of the clipboard message.
func (e ClipboardMsg) String() string {
	return string(e)
}
