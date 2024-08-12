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
		return KeyPressMsg{Mod: mod, Sym: KeyBackspace}
	case ansi.HT:
		return KeyPressMsg{Mod: mod, Sym: KeyTab}
	case ansi.CR:
		return KeyPressMsg{Mod: mod, Sym: KeyEnter}
	case ansi.ESC:
		return KeyPressMsg{Mod: mod, Sym: KeyEscape}
	case ansi.DEL:
		return KeyPressMsg{Mod: mod, Sym: KeyBackspace}
	}

	// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
	return KeyPressMsg{
		Mod:   mod,
		Runes: []rune{r},
	}
}

// ModifyOtherKeysEvent represents a modifyOtherKeys event.
//
//	0: disable
//	1: enable mode 1
//	2: enable mode 2
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Functions-using-CSI-_-ordered-by-the-final-character_s_
// See: https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyOtherKeys
type ModifyOtherKeysEvent uint8
