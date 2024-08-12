package tea

// ClipboardEvent is a clipboard read event.
type ClipboardEvent string

// String returns the string representation of the clipboard event.
func (e ClipboardEvent) String() string {
	return string(e)
}
