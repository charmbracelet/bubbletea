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

// WithAltScreen starts the program with the alternate screen buffer enabled
// (i.e. the program starts in full window mode). Note that the altscreen will
// be automatically exited when the program quits.
//
// Example:
//
//	p := tea.NewProgram(Model{}, tea.WithAltScreen())
//	if _, err := p.Run(); err != nil {
//	    fmt.Println("Error running program:", err)
//	    os.Exit(1)
//	}
//
// To enter the altscreen once the program has already started running use the
// EnterAltScreen command.
func WithAltScreen() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withAltScreen
	}
}

// WithoutBracketedPaste starts the program with bracketed paste disabled.
func WithoutBracketedPaste() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withoutBracketedPaste
	}
}

// WithMouseCellMotion starts the program with the mouse enabled in "cell
// motion" mode.
//
// Cell motion mode enables mouse click, release, and wheel events. Mouse
// movement events are also captured if a mouse button is pressed (i.e., drag
// events). Cell motion mode is better supported than all motion mode.
//
// This will try to enable the mouse in extended mode (SGR), if that is not
// supported by the terminal it will fall back to normal mode (X10).
//
// To enable mouse cell motion once the program has already started running use
// the EnableMouseCellMotion command. To disable the mouse when the program is
// running use the DisableMouse command.
//
// The mouse will be automatically disabled when the program exits.
func WithMouseCellMotion() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withMouseCellMotion // set
		p.startupOptions &^= withMouseAllMotion // clear
	}
}

// WithReportFocus enables reporting when the terminal gains and loses
// focus. When this is enabled [FocusMsg] and [BlurMsg] messages will be sent
// to your Update method.
//
// Note that while most terminals and multiplexers support focus reporting,
// some do not. Also note that tmux needs to be configured to report focus
// events.
func WithReportFocus() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withReportFocus
	}
}

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

// WithGraphemeClustering disables grapheme clustering. This is useful if you
// want to disable grapheme clustering for your program.
//
// Grapheme clustering is a character width calculation method that accurately
// calculates the width of wide characters in a terminal. This is useful for
// properly rendering double width characters such as emojis and CJK
// characters.
//
// See https://mitchellh.com/writing/grapheme-clusters-in-terminals
func WithGraphemeClustering() ProgramOption {
	return func(p *Program) {
		p.startupOptions |= withGraphemeClustering
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
