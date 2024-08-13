package tea

// ReportModeMsg is a message that represents a mode report event (DECRPM).
//
// See: https://vt100.net/docs/vt510-rm/DECRPM.html
type ReportModeMsg struct {
	// Mode is the mode number.
	Mode int

	// Value is the mode value.
	Value int
}
