package tea

// PasteMsg is an message that is emitted when a terminal receives pasted text
// using bracketed-paste.
type PasteMsg string

// PasteStartMsg is an message that is emitted when a terminal enters
// bracketed-paste mode.
type PasteStartMsg struct{}

// PasteEvent is an message that is emitted when a terminal receives pasted
// text.
type PasteEndMsg struct{}
