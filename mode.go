package tea

import "github.com/charmbracelet/uv"

// modeReportMsg is a message that represents a mode report event (DECRPM).
//
// See: https://vt100.net/docs/vt510-rm/DECRPM.html
type modeReportMsg = uv.ModeReportEvent
