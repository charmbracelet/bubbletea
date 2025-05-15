package tea

import "github.com/charmbracelet/tv"

// modeReportMsg is a message that represents a mode report event (DECRPM).
//
// See: https://vt100.net/docs/vt510-rm/DECRPM.html
type modeReportMsg = tv.ModeReportEvent
