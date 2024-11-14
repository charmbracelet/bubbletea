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
		return KeyPressMsg{Mod: mod, Code: KeyBackspace}
	case ansi.HT:
		return KeyPressMsg{Mod: mod, Code: KeyTab}
	case ansi.CR:
		return KeyPressMsg{Mod: mod, Code: KeyEnter}
	case ansi.ESC:
		return KeyPressMsg{Mod: mod, Code: KeyEscape}
	case ansi.DEL:
		return KeyPressMsg{Mod: mod, Code: KeyBackspace}
	}

	// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
	k := KeyPressMsg{Code: r, Mod: mod}
	if k.Mod <= ModShift {
		k.Text = string(r)
	}

	return k
}

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
