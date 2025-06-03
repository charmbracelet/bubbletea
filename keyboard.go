package tea

import (
	"github.com/charmbracelet/x/ansi"
)

// KeyboardEnhancements is a type that represents a set of keyboard
// enhancements.
type KeyboardEnhancements struct {
	// Kitty progressive keyboard enhancements protocol. This can be used to
	// enable different keyboard features.
	//
	//  - 0: disable all features
	//  - 1: [ansi.KittyDisambiguateEscapeCodes] Disambiguate escape codes such as
	//  ctrl+i and tab, ctrl+[ and escape, ctrl+space and ctrl+@, etc.
	//  - 2: [ansi.KittyReportEventTypes] Report event types such as key presses,
	//  releases, and repeat events.
	//  - 4: [ansi.KittyReportAlternateKeys] Report keypresses as though they were
	//  on a PC-101 ANSI US keyboard layout regardless of what they layout
	//  actually is. Also include information about whether or not is enabled,
	//  - 8: [ansi.KittyReportAllKeysAsEscapeCodes] Report all key events as escape
	//  codes. This includes simple printable keys like "a" and other Unicode
	//  characters.
	//  - 16: [ansi.KittyReportAssociatedKeys] Report associated text with key
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

	// keyReleases indicates whether we have key release events enabled. This is mainly
	// used in Windows to ignore key releases when they are not requested.
	keyReleases bool
}

// KeyboardEnhancementOption is a type that represents a keyboard enhancement.
type KeyboardEnhancementOption func(k *KeyboardEnhancements)

// withKeyReleases enables support for reporting release key events. This is
// useful for terminals that support the Kitty keyboard protocol "Report event
// types" progressive enhancement feature.
//
// Note that not all terminals support this feature.
func withKeyReleases(k *KeyboardEnhancements) {
	k.kittyFlags |= ansi.KittyReportEventTypes
	k.keyReleases = true
}

// withUniformKeyLayout enables support for reporting key events as though they
// were on a PC-101 layout. This is useful for uniform key event reporting
// across different keyboard layouts. This is equivalent to the Kitty keyboard
// protocol "Report alternate keys" and "Report all keys as escape codes"
// progressive enhancement features.
//
// Note that not all terminals support this feature.
func withUniformKeyLayout(k *KeyboardEnhancements) {
	k.kittyFlags |= ansi.KittyReportAlternateKeys | ansi.KittyReportAllKeysAsEscapeCodes
}

// withKeyDisambiguation enables support for disambiguating keyboard escape
// codes. This is useful for terminals that support the Kitty keyboard protocol
// "Disambiguate escape codes" progressive enhancement feature or the XTerm
// modifyOtherKeys mode 1 feature to report ambiguous keys as escape codes.
func withKeyDisambiguation(k *KeyboardEnhancements) {
	k.kittyFlags |= ansi.KittyDisambiguateEscapeCodes
	if k.modifyOtherKeys < 1 {
		k.modifyOtherKeys = 1
	}
}

type enableKeyboardEnhancementsMsg []KeyboardEnhancementOption

// RequestKeyDisambiguation is a command that enables support for reporting
// disambiguous keys as escape codes. This is enabled by default in Bubble Tea
// and there's no need to call this function unless you disabled keyboard
// enhancements through [DisableKeyboardEnhancements].
//
// If the terminal supports the requested enhancements, it will send a
// [KeyboardEnhancementsMsg] message with the supported enhancements.
//
// Note that not all terminals support this feature. If the terminal does not
// support this feature, the program will not receive disambiguated key
// events.
func RequestKeyDisambiguation() Msg {
	return enableKeyboardEnhancementsMsg{withKeyDisambiguation}
}

// RequestKeyReleases is a command that enables support for reporting key
// release events.
//
// If the terminal supports the requested enhancements, it will send a
// [KeyboardEnhancementsMsg] message with the supported enhancements.
//
// Note that not all terminals support all enhancements.
func RequestKeyReleases() Msg {
	return enableKeyboardEnhancementsMsg{withKeyReleases}
}

// RequestUniformKeyLayout is a command that enables support for reporting key
// events as though they were on a PC-101 layout.
//
// If the terminal supports the requested enhancements, it will send a
// [KeyboardEnhancementsMsg] message with the supported enhancements.
//
// Note that not all terminals support all enhancements.
func RequestUniformKeyLayout() Msg {
	return enableKeyboardEnhancementsMsg{withUniformKeyLayout}
}

type disableKeyboardEnhancementsMsg struct{}

// DisableKeyboardEnhancements is a command that disables keyboard enhancements
// in the terminal.
func DisableKeyboardEnhancements() Msg {
	return disableKeyboardEnhancementsMsg{}
}

// KeyboardEnhancementsMsg is a message that gets sent when the terminal
// supports keyboard enhancements.
type KeyboardEnhancementsMsg KeyboardEnhancements

// SupportsKeyDisambiguation returns whether the terminal supports reporting
// disambiguous keys as escape codes.
func (k KeyboardEnhancementsMsg) SupportsKeyDisambiguation() bool {
	if isWindows() {
		// We use Windows Console API which supports reporting disambiguous keys.
		return true
	}
	return k.kittyFlags&ansi.KittyDisambiguateEscapeCodes != 0 || k.modifyOtherKeys >= 1
}

// SupportsKeyReleases returns whether the terminal supports key release
// events.
func (k KeyboardEnhancementsMsg) SupportsKeyReleases() bool {
	if isWindows() {
		// We use Windows Console API which supports key release events.
		return k.keyReleases
	}
	return k.kittyFlags&ansi.KittyReportEventTypes != 0
}

// SupportsUniformKeyLayout returns whether the terminal supports reporting key
// events as though they were on a PC-101 layout.
func (k KeyboardEnhancementsMsg) SupportsUniformKeyLayout() bool {
	if isWindows() {
		// We use Windows Console API which supports reporting key events as
		// though they were on a PC-101 layout.
		return true
	}
	return k.SupportsKeyDisambiguation() &&
		k.kittyFlags&ansi.KittyReportAlternateKeys != 0 &&
		k.kittyFlags&ansi.KittyReportAllKeysAsEscapeCodes != 0
}
