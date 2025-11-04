package tea

import (
	"github.com/charmbracelet/x/ansi"
)

// KeyboardEnhancementsMsg is a message that gets sent when the terminal
// supports keyboard enhancements.
type KeyboardEnhancementsMsg struct {
	Flags int
}

// SupportsKeyDisambiguation returns whether the terminal supports reporting
// disambiguous keys as escape codes.
func (k KeyboardEnhancementsMsg) SupportsKeyDisambiguation() bool {
	return k.Flags&ansi.KittyDisambiguateEscapeCodes != 0
}

// SupportsKeyReleases returns whether the terminal supports key release
// events.
func (k KeyboardEnhancementsMsg) SupportsKeyReleases() bool {
	return k.Flags&ansi.KittyReportEventTypes != 0
}

// SupportsUniformKeyLayout returns whether the terminal supports reporting key
// events as though they were on a PC-101 layout.
func (k KeyboardEnhancementsMsg) SupportsUniformKeyLayout() bool {
	return k.SupportsKeyDisambiguation() &&
		k.Flags&ansi.KittyReportAlternateKeys != 0 &&
		k.Flags&ansi.KittyReportAllKeysAsEscapeCodes != 0
}
