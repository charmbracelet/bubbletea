package tea

import (
	"github.com/charmbracelet/x/ansi"
)

func parseXTermModifyOtherKeys(csi *ansi.CsiSequence) Msg {
	// XTerm modify other keys starts with ESC [ 27 ; <modifier> ; <code> ~
	mod := KeyMod(csi.Param(1) - 1)
	r := rune(csi.Param(2))

	switch r {
	case ansi.BS:
		return KeyPressMsg{Mod: mod, Type: KeyBackspace}
	case ansi.HT:
		return KeyPressMsg{Mod: mod, Type: KeyTab}
	case ansi.CR:
		return KeyPressMsg{Mod: mod, Type: KeyEnter}
	case ansi.ESC:
		return KeyPressMsg{Mod: mod, Type: KeyEscape}
	case ansi.DEL:
		return KeyPressMsg{Mod: mod, Type: KeyBackspace}
	}

	// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
	return KeyPressMsg{
		Mod:   mod,
		Runes: []rune{r},
	}
}

// ModifyOtherKeysMsg is a message that represents XTerm modifyOtherKeys
// report. Querying the terminal for the modifyOtherKeys mode will return a
// ModifyOtherKeysMsg message with the current mode set.
//
//	0: disable
//	1: enable mode 1
//	2: enable mode 2
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Functions-using-CSI-_-ordered-by-the-final-character_s_
// See: https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyOtherKeys
type ModifyOtherKeysMsg uint8

// TerminalVersionMsg is a message that represents the terminal version.
type TerminalVersionMsg string

// terminalVersion is an internal message that queries the terminal for its
// version using XTVERSION.
type terminalVersion struct{}

// TerminalVersion is a command that queries the terminal for its version using
// XTVERSION. Note that some terminals may not support this command.
func TerminalVersion() Msg {
	return terminalVersion{}
}
