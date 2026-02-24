package tea

import (
	"github.com/charmbracelet/x/ansi"
)

// KeyboardEnhancementsMsg is a message that gets sent when the terminal
// supports keyboard enhancements.
type KeyboardEnhancementsMsg struct {
	// Flags is a bitmask of enabled keyboard enhancement features. A non-zero
	// value indicates that at least we have key disambiguation support.
	//
	// See [ansi.KittyReportEventTypes] and other constants for details.
	//
	// Example:
	//
	//  ```go
	//  // The hard way
	//  if msg.Flags&ansi.KittyReportEventTypes != 0 {
	//     // Terminal supports reporting different key event types
	//  }
	//
	//  // The easy way
	//  if msg.SupportsEventTypes() {
	//     // Terminal supports reporting different key event types
	//  }
	//  ```
	Flags int
}

// SupportsKeyDisambiguation returns whether the terminal supports key
// disambiguation (e.g., distinguishing between different modifier keys).
func (k KeyboardEnhancementsMsg) SupportsKeyDisambiguation() bool {
	return k.Flags > 0
}

// SupportsEventTypes returns whether the terminal supports reporting
// different types of key events (press, release, and repeat).
func (k KeyboardEnhancementsMsg) SupportsEventTypes() bool {
	return k.Flags&ansi.KittyReportEventTypes != 0
}
