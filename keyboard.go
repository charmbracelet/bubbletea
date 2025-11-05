package tea

import (
	"github.com/charmbracelet/x/ansi"
)

// KeyboardEnhancementsMsg is a message that gets sent when the terminal
// supports keyboard enhancements.
type KeyboardEnhancementsMsg struct {
	// Flags is a bitmask of supported keyboard enhancement features.
	// See [ansi.KittyReportEventTypes] and other constants for details.
	Flags int
}

// SupportsEventTypes returns whether the terminal supports reporting
// different types of key events (press, release, repeat).
func (k KeyboardEnhancementsMsg) SupportsEventTypes() bool {
	return k.Flags&ansi.KittyReportEventTypes != 0
}
