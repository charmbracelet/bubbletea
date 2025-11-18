package tea

// PasteMsg is an message that is emitted when a terminal receives pasted text
// using bracketed-paste.
type PasteMsg struct {
	Content string
}

// String returns the pasted content as a string.
func (p PasteMsg) String() string {
	return p.Content
}

// PasteStartMsg is an message that is emitted when the terminal starts the
// bracketed-paste text.
type PasteStartMsg struct{}

// PasteEndMsg is an message that is emitted when the terminal ends the
// bracketed-paste text.
type PasteEndMsg struct{}
