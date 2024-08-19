package tea

// PasteMsg is an message that is emitted when a terminal receives pasted text
// using bracketed-paste.
type PasteMsg string

// PasteStartMsg is an message that is emitted when the terminal starts the
// bracketed-paste text
type PasteStartMsg struct{}

// PasteEndMsg is an message that is emitted when the terminal ends the
// bracketed-paste text.
type PasteEndMsg struct{}
