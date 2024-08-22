package tea

import (
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// setKittyKeyboardFlagsMsg is a message to set Kitty keyboard progressive
// enhancement protocol flags.
type setKittyKeyboardFlagsMsg int

// enableKittyKeyboard is a command to enable Kitty keyboard progressive
// enhancements.
//
// The flags parameter is a bitmask of the following
//
//	1:  Disambiguate escape codes
//	2:  Report event types
//	4:  Report alternate keys
//	8:  Report all keys as escape codes
//	16: Report associated text
//
// See https://sw.kovidgoyal.net/kitty/keyboard-protocol/ for more information.
func enableKittyKeyboard(flags int) Cmd { //nolint:unused
	return func() Msg {
		return setKittyKeyboardFlagsMsg(flags)
	}
}

// disableKittyKeyboard is a command to disable Kitty keyboard progressive
// enhancements.
func disableKittyKeyboard() Msg { //nolint:unused
	return setKittyKeyboardFlagsMsg(0)
}

// kittyKeyboardMsg is a message that queries the current Kitty keyboard
// progressive enhancement flags.
type kittyKeyboardMsg struct{}

// kittyKeyboard is a command that queries the current Kitty keyboard
// progressive enhancement flags from the terminal.
func kittyKeyboard() Msg { //nolint:unused
	return kittyKeyboardMsg{}
}

// _KittyKeyboardMsg represents Kitty keyboard progressive enhancement flags message.
type _KittyKeyboardMsg int

// Kitty Clipboard Control Sequences
var kittyKeyMap = map[int]KeyType{
	ansi.BS:  KeyBackspace,
	ansi.HT:  KeyTab,
	ansi.CR:  KeyEnter,
	ansi.ESC: KeyEscape,
	ansi.DEL: KeyBackspace,

	57344: KeyEscape,
	57345: KeyEnter,
	57346: KeyTab,
	57347: KeyBackspace,
	57348: KeyInsert,
	57349: KeyDelete,
	57350: KeyLeft,
	57351: KeyRight,
	57352: KeyUp,
	57353: KeyDown,
	57354: KeyPgUp,
	57355: KeyPgDown,
	57356: KeyHome,
	57357: KeyEnd,
	57358: KeyCapsLock,
	57359: KeyScrollLock,
	57360: KeyNumLock,
	57361: KeyPrintScreen,
	57362: KeyPause,
	57363: KeyMenu,
	57364: KeyF1,
	57365: KeyF2,
	57366: KeyF3,
	57367: KeyF4,
	57368: KeyF5,
	57369: KeyF6,
	57370: KeyF7,
	57371: KeyF8,
	57372: KeyF9,
	57373: KeyF10,
	57374: KeyF11,
	57375: KeyF12,
	57376: KeyF13,
	57377: KeyF14,
	57378: KeyF15,
	57379: KeyF16,
	57380: KeyF17,
	57381: KeyF18,
	57382: KeyF19,
	57383: KeyF20,
	57384: KeyF21,
	57385: KeyF22,
	57386: KeyF23,
	57387: KeyF24,
	57388: KeyF25,
	57389: KeyF26,
	57390: KeyF27,
	57391: KeyF28,
	57392: KeyF29,
	57393: KeyF30,
	57394: KeyF31,
	57395: KeyF32,
	57396: KeyF33,
	57397: KeyF34,
	57398: KeyF35,
	57399: KeyKp0,
	57400: KeyKp1,
	57401: KeyKp2,
	57402: KeyKp3,
	57403: KeyKp4,
	57404: KeyKp5,
	57405: KeyKp6,
	57406: KeyKp7,
	57407: KeyKp8,
	57408: KeyKp9,
	57409: KeyKpDecimal,
	57410: KeyKpDivide,
	57411: KeyKpMultiply,
	57412: KeyKpMinus,
	57413: KeyKpPlus,
	57414: KeyKpEnter,
	57415: KeyKpEqual,
	57416: KeyKpSep,
	57417: KeyKpLeft,
	57418: KeyKpRight,
	57419: KeyKpUp,
	57420: KeyKpDown,
	57421: KeyKpPgUp,
	57422: KeyKpPgDown,
	57423: KeyKpHome,
	57424: KeyKpEnd,
	57425: KeyKpInsert,
	57426: KeyKpDelete,
	57427: KeyKpBegin,
	57428: KeyMediaPlay,
	57429: KeyMediaPause,
	57430: KeyMediaPlayPause,
	57431: KeyMediaReverse,
	57432: KeyMediaStop,
	57433: KeyMediaFastForward,
	57434: KeyMediaRewind,
	57435: KeyMediaNext,
	57436: KeyMediaPrev,
	57437: KeyMediaRecord,
	57438: KeyLowerVol,
	57439: KeyRaiseVol,
	57440: KeyMute,
	57441: KeyLeftShift,
	57442: KeyLeftCtrl,
	57443: KeyLeftAlt,
	57444: KeyLeftSuper,
	57445: KeyLeftHyper,
	57446: KeyLeftMeta,
	57447: KeyRightShift,
	57448: KeyRightCtrl,
	57449: KeyRightAlt,
	57450: KeyRightSuper,
	57451: KeyRightHyper,
	57452: KeyRightMeta,
	57453: KeyIsoLevel3Shift,
	57454: KeyIsoLevel5Shift,
}

const (
	kittyShift = 1 << iota
	kittyAlt
	kittyCtrl
	kittySuper
	kittyHyper
	kittyMeta
	kittyCapsLock
	kittyNumLock
)

func fromKittyMod(mod int) KeyMod {
	var m KeyMod
	if mod&kittyShift != 0 {
		m |= ModShift
	}
	if mod&kittyAlt != 0 {
		m |= ModAlt
	}
	if mod&kittyCtrl != 0 {
		m |= ModCtrl
	}
	if mod&kittySuper != 0 {
		m |= ModSuper
	}
	if mod&kittyHyper != 0 {
		m |= ModHyper
	}
	if mod&kittyMeta != 0 {
		m |= ModMeta
	}
	if mod&kittyCapsLock != 0 {
		m |= ModCapsLock
	}
	if mod&kittyNumLock != 0 {
		m |= ModNumLock
	}
	return m
}

// parseKittyKeyboard parses a Kitty Keyboard Protocol sequence.
//
// In `CSI u`, this is parsed as:
//
//	CSI codepoint ; modifiers u
//	codepoint: ASCII Dec value
//
// The Kitty Keyboard Protocol extends this with optional components that can be
// enabled progressively. The full sequence is parsed as:
//
//	CSI unicode-key-code:alternate-key-codes ; modifiers:event-type ; text-as-codepoints u
//
// See https://sw.kovidgoyal.net/kitty/keyboard-protocol/
func parseKittyKeyboard(csi *ansi.CsiSequence) Msg {
	var isRelease bool
	key := Key{}

	if params := csi.Subparams(0); len(params) > 0 {
		code := params[0]
		if sym, ok := kittyKeyMap[code]; ok {
			key.Type = sym
		} else {
			r := rune(code)
			if !utf8.ValidRune(r) {
				r = utf8.RuneError
			}

			key.Type = KeyRunes
			key.Runes = []rune{r}

			// alternate key reporting
			switch len(params) {
			case 3:
				// shifted key + base key
				if b := rune(params[2]); unicode.IsPrint(b) {
					// XXX: When alternate key reporting is enabled, the protocol
					// can return 3 things, the unicode codepoint of the key,
					// the shifted codepoint of the key, and the standard
					// PC-101 key layout codepoint.
					// This is useful to create an unambiguous mapping of keys
					// when using a different language layout.
					key.baseRune = b
				}
				fallthrough
			case 2:
				// shifted key
				if s := rune(params[1]); unicode.IsPrint(s) {
					// XXX: We swap keys here because we want the shifted key
					// to be the Rune that is returned by the event.
					// For example, shift+a should produce "A" not "a".
					// In such a case, we set AltRune to the original key "a"
					// and Rune to "A".
					key.altRune = key.Rune()
					key.Runes = []rune{s}
				}
			}
		}
	}
	if params := csi.Subparams(1); len(params) > 0 {
		mod := params[0]
		if mod > 1 {
			key.Mod = fromKittyMod(mod - 1)
		}
		if len(params) > 1 {
			switch params[1] {
			case 2:
				key.IsRepeat = true
			case 3:
				isRelease = true
			}
		}
	}
	if params := csi.Subparams(2); len(params) > 0 {
		r := rune(params[0])
		if unicode.IsPrint(r) {
			key.altRune = key.Rune()
			key.Runes = []rune{r}
		}
	}
	if isRelease {
		return KeyReleaseMsg(key)
	}
	return KeyPressMsg(key)
}
