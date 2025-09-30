package tea

import (
	"github.com/charmbracelet/x/ansi"
)

// ProgramOption is used to set options when initializing a Program. Program can
// accept a variable number of options.
//
// Example usage:
//
//	p := NewProgram(model, WithInput(someInput), WithOutput(someOutput))
type ProgramOption func(*Program)

// WithKeyReleases enables support for reporting key release events. This is
// useful for terminals that support the Kitty keyboard protocol "Report event
// types" progressive enhancement feature.
//
// Note that not all terminals support this feature. If the terminal does not
// support this feature, the program will not receive key release events.
func WithKeyReleases() ProgramOption {
	return func(p *Program) {
		p.requestedEnhancements.kittyFlags |= ansi.KittyReportEventTypes
		p.requestedEnhancements.keyReleases = true
	}
}

// WithUniformKeyLayout enables support for reporting key events as though they
// were on a PC-101 layout. This is useful for uniform key event reporting
// across different keyboard layouts. This is equivalent to the Kitty keyboard
// protocol "Report alternate keys" and "Report all keys as escape codes"
// progressive enhancement features.
//
// Note that not all terminals support this feature. If the terminal does not
// support this feature, the program will not receive key events in
// uniform layout format.
func WithUniformKeyLayout() ProgramOption {
	return func(p *Program) {
		p.requestedEnhancements.kittyFlags |= ansi.KittyReportAlternateKeys | ansi.KittyReportAllKeysAsEscapeCodes
	}
}

// WithoutKeyEnhancements disables all key enhancements. This is useful if you
// want to disable all key enhancements for your program and keep your program
// legacy compatible with older terminals.
func WithoutKeyEnhancements() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withoutKeyEnhancements
	}
}
