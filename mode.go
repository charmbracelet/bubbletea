package tea

import "github.com/charmbracelet/x/input"

// modeReportMsg is a message that represents a mode report event (DECRPM).
//
// See: https://vt100.net/docs/vt510-rm/DECRPM.html
type modeReportMsg = input.ModeReportEvent
