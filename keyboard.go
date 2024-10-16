package tea

import (
	"runtime"

	"github.com/charmbracelet/x/ansi"
)

// keyboardEnhancements is a type that represents a set of keyboard
// enhancements.
type keyboardEnhancements struct {
	// Kitty progressive keyboard enhancements protocol. This can be used to
	// enable different keyboard features.
	//
	//  - 0: disable all features
	//  - 1: [ansi.DisambiguateEscapeCodes] Disambiguate escape codes such as
	//  ctrl+i and tab, ctrl+[ and escape, ctrl+space and ctrl+@, etc.
	//  - 2: [ansi.ReportEventTypes] Report event types such as key presses,
	//  releases, and repeat events.
	//  - 4: [ansi.ReportAlternateKeys] Report keypresses as though they were
	//  on a PC-101 ANSI US keyboard layout regardless of what they layout
	//  actually is. Also include information about whether or not is enabled,
	//  - 8: [ansi.ReportAllKeysAsEscapeCodes] Report all key events as escape
	//  codes. This includes simple printable keys like "a" and other Unicode
	//  characters.
	//  - 16: [ansi.ReportAssociatedText] Report associated text with key
	//  events. This encodes multi-rune key events as escape codes instead of
	//  individual runes.
	//
	kittyFlags int

	// Xterm modifyOtherKeys feature.
	//
	//  - Mode 0 disables modifyOtherKeys.
	//  - Mode 1 reports ambiguous keys as escape codes. This is similar to
	//  [ansi.KittyDisambiguateEscapeCodes] but uses XTerm escape codes.
	//  - Mode 2 reports all key as escape codes including printable keys like "a" and "shift+b".
	modifyOtherKeys int
}

// KeyboardEnhancement is a type that represents a keyboard enhancement.
type KeyboardEnhancement func(k *keyboardEnhancements)

// WithKeyReleases enables support for reporting release key events. This is
// useful for terminals that support the Kitty keyboard protocol "Report event
// types" progressive enhancement feature.
//
// Note that not all terminals support this feature.
func WithKeyReleases(k *keyboardEnhancements) {
	k.kittyFlags |= ansi.KittyReportEventTypes
}

// WithUniformKeyLayout enables support for reporting key events as though they
// were on a PC-101 layout. This is useful for uniform key event reporting
// across different keyboard layouts. This is equivalent to the Kitty keyboard
// protocol "Report alternate keys" and "Report all keys as escape codes"
// progressive enhancement features.
//
// Note that not all terminals support this feature.
func WithUniformKeyLayout(k *keyboardEnhancements) {
	k.kittyFlags |= ansi.KittyReportAlternateKeys | ansi.KittyReportAllKeysAsEscapeCodes
}

// withKeyDisambiguation enables support for disambiguating keyboard escape
// codes. This is useful for terminals that support the Kitty keyboard protocol
// "Disambiguate escape codes" progressive enhancement feature or the XTerm
// modifyOtherKeys mode 1 feature to report ambiguous keys as escape codes.
func withKeyDisambiguation(k *keyboardEnhancements) {
	k.kittyFlags |= ansi.KittyDisambiguateEscapeCodes
	if k.modifyOtherKeys < 1 {
		k.modifyOtherKeys = 1
	}
}

type enableKeyboardEnhancementsMsg []KeyboardEnhancement

// EnableKeyboardEnhancements is a command that enables keyboard enhancements
// in the terminal.
func EnableKeyboardEnhancements(enhancements ...KeyboardEnhancement) Cmd {
	return func() Msg {
		return enableKeyboardEnhancementsMsg(append(enhancements, withKeyDisambiguation))
	}
}

type disableKeyboardEnhancementsMsg struct{}

// DisableKeyboardEnhancements is a command that disables keyboard enhancements
// in the terminal.
func DisableKeyboardEnhancements() Msg {
	return disableKeyboardEnhancementsMsg{}
}

// KeyboardEnhancementsMsg is a message that gets sent when the terminal
// supports keyboard enhancements.
type KeyboardEnhancementsMsg keyboardEnhancements

// SupportsKeyDisambiguation returns whether the terminal supports reporting
// disambiguous keys as escape codes.
func (k KeyboardEnhancementsMsg) SupportsKeyDisambiguation() bool {
	if runtime.GOOS == "windows" {
		// We use Windows Console API which supports reporting disambiguous keys.
		return true
	}
	return k.kittyFlags&ansi.KittyDisambiguateEscapeCodes != 0 || k.modifyOtherKeys >= 1
}

// SupportsKeyReleases returns whether the terminal supports key release
// events.
func (k KeyboardEnhancementsMsg) SupportsKeyReleases() bool {
	if runtime.GOOS == "windows" {
		// We use Windows Console API which supports key release events.
		return true
	}
	return k.kittyFlags&ansi.KittyReportEventTypes != 0
}

// SupportsUniformKeyLayout returns whether the terminal supports reporting key
// events as though they were on a PC-101 layout.
func (k KeyboardEnhancementsMsg) SupportsUniformKeyLayout() bool {
	if runtime.GOOS == "windows" {
		// We use Windows Console API which supports reporting key events as
		// though they were on a PC-101 layout.
		return true
	}
	return k.SupportsKeyDisambiguation() &&
		k.kittyFlags&ansi.KittyReportAlternateKeys != 0 &&
		k.kittyFlags&ansi.KittyReportAllKeysAsEscapeCodes != 0
}
