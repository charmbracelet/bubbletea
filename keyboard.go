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

// EnableKeyboardEnhancements is a [Cmd] for enables keyboard enhancements, in
// supporting terminals. By default, keyboard enhancements do a couple things:
//
//   - It enables you to match all modifier keys such as super.
//   - It enables you to match all keys that are ambiguous such as ctrl+i,
//     which normally would be indistinguishable from tab.
//
// You can also enable specific enhancements, such as key releases, by passing
// them as arguments to the command:
//
//	cmd := EnableKeyboardEnhancements(WithKeyReleases)
//
// For available enhancements options see [KeyboardEnhancement].
//
// Note that not all terminals support these features. You can check if the
// terminal supports these features by matching on the
// [KeyboardEnhancementsMsg] message.
//
// This feature is enabled by default on Windows.
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
// supports keyboard enhancements. For some background on what this does, see
// the [EnableKeyboardEnhancements] command.
//
// Keyboard enhancements can be enabled when constructing a [Program] by using
// the [WithKeyboardEnhancements] option:
//
//	p := tea.NewProgram(initialModel, tea.WithKeyboardEnhancements())
//
// Or with the [EnableKeyboardEnhancements] command:
//
//	cmd := EnableKeyboardEnhancements()
//
// You can also enable specific enhancements by passing them as arguments to
// the [EnableKeyboardEnhancements] command:
//
//	cmd := EnableKeyboardEnhancements(WithKeyReleases, WithKeyDisambiguation)
//
// For available enhancements options see [KeyboardEnhancement].
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
