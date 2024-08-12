package tea

// ReportModeEvent represents a report mode event for sequence DECRPM.
//
// See: https://vt100.net/docs/vt510-rm/DECRPM.html
type ReportModeEvent struct {
	// Mode is the mode number.
	Mode int

	// Value is the mode value.
	Value int
}
