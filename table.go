package tea

import (
	"strconv"

	"github.com/charmbracelet/x/ansi"
)

// buildKeysTable builds a table of key sequences and their corresponding key
// events based on the VT100/VT200, XTerm, and Urxvt terminal specs.
// TODO: Use flags?
func buildKeysTable(flags int, term string) map[string]Key {
	nul := Key{Runes: []rune{' '}, Type: KeySpace, Mod: ModCtrl} // ctrl+@ or ctrl+space
	if flags&_FlagCtrlAt != 0 {
		nul = Key{Runes: []rune{'@'}, Mod: ModCtrl}
	}

	tab := Key{Type: KeyTab} // ctrl+i or tab
	if flags&_FlagCtrlI != 0 {
		tab = Key{Runes: []rune{'i'}, Mod: ModCtrl}
	}

	enter := Key{Type: KeyEnter} // ctrl+m or enter
	if flags&_FlagCtrlM != 0 {
		enter = Key{Runes: []rune{'m'}, Mod: ModCtrl}
	}

	esc := Key{Type: KeyEscape} // ctrl+[ or escape
	if flags&_FlagCtrlOpenBracket != 0 {
		esc = Key{Runes: []rune{'['}, Mod: ModCtrl} // ctrl+[ or escape
	}

	del := Key{Type: KeyBackspace}
	if flags&_FlagBackspace != 0 {
		del.Type = KeyDelete
	}

	find := Key{Type: KeyHome}
	if flags&_FlagFind != 0 {
		find.Type = KeyFind
	}

	sel := Key{Type: KeyEnd}
	if flags&_FlagSelect != 0 {
		sel.Type = KeySelect
	}

	// The following is a table of key sequences and their corresponding key
	// events based on the VT100/VT200 terminal specs.
	//
	// See: https://vt100.net/docs/vt100-ug/chapter3.html#S3.2
	// See: https://vt100.net/docs/vt220-rm/chapter3.html
	//
	// XXX: These keys may be overwritten by other options like XTerm or
	// Terminfo.
	table := map[string]Key{
		// C0 control characters
		string(byte(ansi.NUL)): nul,
		string(byte(ansi.SOH)): {Runes: []rune{'a'}, Mod: ModCtrl},
		string(byte(ansi.STX)): {Runes: []rune{'b'}, Mod: ModCtrl},
		string(byte(ansi.ETX)): {Runes: []rune{'c'}, Mod: ModCtrl},
		string(byte(ansi.EOT)): {Runes: []rune{'d'}, Mod: ModCtrl},
		string(byte(ansi.ENQ)): {Runes: []rune{'e'}, Mod: ModCtrl},
		string(byte(ansi.ACK)): {Runes: []rune{'f'}, Mod: ModCtrl},
		string(byte(ansi.BEL)): {Runes: []rune{'g'}, Mod: ModCtrl},
		string(byte(ansi.BS)):  {Runes: []rune{'h'}, Mod: ModCtrl},
		string(byte(ansi.HT)):  tab,
		string(byte(ansi.LF)):  {Runes: []rune{'j'}, Mod: ModCtrl},
		string(byte(ansi.VT)):  {Runes: []rune{'k'}, Mod: ModCtrl},
		string(byte(ansi.FF)):  {Runes: []rune{'l'}, Mod: ModCtrl},
		string(byte(ansi.CR)):  enter,
		string(byte(ansi.SO)):  {Runes: []rune{'n'}, Mod: ModCtrl},
		string(byte(ansi.SI)):  {Runes: []rune{'o'}, Mod: ModCtrl},
		string(byte(ansi.DLE)): {Runes: []rune{'p'}, Mod: ModCtrl},
		string(byte(ansi.DC1)): {Runes: []rune{'q'}, Mod: ModCtrl},
		string(byte(ansi.DC2)): {Runes: []rune{'r'}, Mod: ModCtrl},
		string(byte(ansi.DC3)): {Runes: []rune{'s'}, Mod: ModCtrl},
		string(byte(ansi.DC4)): {Runes: []rune{'t'}, Mod: ModCtrl},
		string(byte(ansi.NAK)): {Runes: []rune{'u'}, Mod: ModCtrl},
		string(byte(ansi.SYN)): {Runes: []rune{'v'}, Mod: ModCtrl},
		string(byte(ansi.ETB)): {Runes: []rune{'w'}, Mod: ModCtrl},
		string(byte(ansi.CAN)): {Runes: []rune{'x'}, Mod: ModCtrl},
		string(byte(ansi.EM)):  {Runes: []rune{'y'}, Mod: ModCtrl},
		string(byte(ansi.SUB)): {Runes: []rune{'z'}, Mod: ModCtrl},
		string(byte(ansi.ESC)): esc,
		string(byte(ansi.FS)):  {Runes: []rune{'\\'}, Mod: ModCtrl},
		string(byte(ansi.GS)):  {Runes: []rune{']'}, Mod: ModCtrl},
		string(byte(ansi.RS)):  {Runes: []rune{'^'}, Mod: ModCtrl},
		string(byte(ansi.US)):  {Runes: []rune{'_'}, Mod: ModCtrl},

		// Special keys in G0
		string(byte(ansi.SP)):  {Type: KeySpace, Runes: []rune{' '}},
		string(byte(ansi.DEL)): del,

		// Special keys

		"\x1b[Z": {Type: KeyTab, Mod: ModShift},

		"\x1b[1~": find,
		"\x1b[2~": {Type: KeyInsert},
		"\x1b[3~": {Type: KeyDelete},
		"\x1b[4~": sel,
		"\x1b[5~": {Type: KeyPgUp},
		"\x1b[6~": {Type: KeyPgDown},
		"\x1b[7~": {Type: KeyHome},
		"\x1b[8~": {Type: KeyEnd},

		// Normal mode
		"\x1b[A": {Type: KeyUp},
		"\x1b[B": {Type: KeyDown},
		"\x1b[C": {Type: KeyRight},
		"\x1b[D": {Type: KeyLeft},
		"\x1b[E": {Type: KeyBegin},
		"\x1b[F": {Type: KeyEnd},
		"\x1b[H": {Type: KeyHome},
		"\x1b[P": {Type: KeyF1},
		"\x1b[Q": {Type: KeyF2},
		"\x1b[R": {Type: KeyF3},
		"\x1b[S": {Type: KeyF4},

		// Application Cursor Key Mode (DECCKM)
		"\x1bOA": {Type: KeyUp},
		"\x1bOB": {Type: KeyDown},
		"\x1bOC": {Type: KeyRight},
		"\x1bOD": {Type: KeyLeft},
		"\x1bOE": {Type: KeyBegin},
		"\x1bOF": {Type: KeyEnd},
		"\x1bOH": {Type: KeyHome},
		"\x1bOP": {Type: KeyF1},
		"\x1bOQ": {Type: KeyF2},
		"\x1bOR": {Type: KeyF3},
		"\x1bOS": {Type: KeyF4},

		// Keypad Application Mode (DECKPAM)

		"\x1bOM": {Type: KeyKpEnter},
		"\x1bOX": {Type: KeyKpEqual},
		"\x1bOj": {Type: KeyKpMultiply},
		"\x1bOk": {Type: KeyKpPlus},
		"\x1bOl": {Type: KeyKpComma},
		"\x1bOm": {Type: KeyKpMinus},
		"\x1bOn": {Type: KeyKpDecimal},
		"\x1bOo": {Type: KeyKpDivide},
		"\x1bOp": {Type: KeyKp0},
		"\x1bOq": {Type: KeyKp1},
		"\x1bOr": {Type: KeyKp2},
		"\x1bOs": {Type: KeyKp3},
		"\x1bOt": {Type: KeyKp4},
		"\x1bOu": {Type: KeyKp5},
		"\x1bOv": {Type: KeyKp6},
		"\x1bOw": {Type: KeyKp7},
		"\x1bOx": {Type: KeyKp8},
		"\x1bOy": {Type: KeyKp9},

		// Function keys

		"\x1b[11~": {Type: KeyF1},
		"\x1b[12~": {Type: KeyF2},
		"\x1b[13~": {Type: KeyF3},
		"\x1b[14~": {Type: KeyF4},
		"\x1b[15~": {Type: KeyF5},
		"\x1b[17~": {Type: KeyF6},
		"\x1b[18~": {Type: KeyF7},
		"\x1b[19~": {Type: KeyF8},
		"\x1b[20~": {Type: KeyF9},
		"\x1b[21~": {Type: KeyF10},
		"\x1b[23~": {Type: KeyF11},
		"\x1b[24~": {Type: KeyF12},
		"\x1b[25~": {Type: KeyF13},
		"\x1b[26~": {Type: KeyF14},
		"\x1b[28~": {Type: KeyF15},
		"\x1b[29~": {Type: KeyF16},
		"\x1b[31~": {Type: KeyF17},
		"\x1b[32~": {Type: KeyF18},
		"\x1b[33~": {Type: KeyF19},
		"\x1b[34~": {Type: KeyF20},
	}

	// CSI ~ sequence keys
	csiTildeKeys := map[string]Key{
		"1": find, "2": {Type: KeyInsert},
		"3": {Type: KeyDelete}, "4": sel,
		"5": {Type: KeyPgUp}, "6": {Type: KeyPgDown},
		"7": {Type: KeyHome}, "8": {Type: KeyEnd},
		// There are no 9 and 10 keys
		"11": {Type: KeyF1}, "12": {Type: KeyF2},
		"13": {Type: KeyF3}, "14": {Type: KeyF4},
		"15": {Type: KeyF5}, "17": {Type: KeyF6},
		"18": {Type: KeyF7}, "19": {Type: KeyF8},
		"20": {Type: KeyF9}, "21": {Type: KeyF10},
		"23": {Type: KeyF11}, "24": {Type: KeyF12},
		"25": {Type: KeyF13}, "26": {Type: KeyF14},
		"28": {Type: KeyF15}, "29": {Type: KeyF16},
		"31": {Type: KeyF17}, "32": {Type: KeyF18},
		"33": {Type: KeyF19}, "34": {Type: KeyF20},
	}

	// URxvt keys
	// See https://manpages.ubuntu.com/manpages/trusty/man7/urxvt.7.html#key%20codes
	table["\x1b[a"] = Key{Type: KeyUp, Mod: ModShift}
	table["\x1b[b"] = Key{Type: KeyDown, Mod: ModShift}
	table["\x1b[c"] = Key{Type: KeyRight, Mod: ModShift}
	table["\x1b[d"] = Key{Type: KeyLeft, Mod: ModShift}
	table["\x1bOa"] = Key{Type: KeyUp, Mod: ModCtrl}
	table["\x1bOb"] = Key{Type: KeyDown, Mod: ModCtrl}
	table["\x1bOc"] = Key{Type: KeyRight, Mod: ModCtrl}
	table["\x1bOd"] = Key{Type: KeyLeft, Mod: ModCtrl}
	// TODO: invistigate if shift-ctrl arrow keys collide with DECCKM keys i.e.
	// "\x1bOA", "\x1bOB", "\x1bOC", "\x1bOD"

	// URxvt modifier CSI ~ keys
	for k, v := range csiTildeKeys {
		key := v
		// Normal (no modifier) already defined part of VT100/VT200
		// Shift modifier
		key.Mod = ModShift
		table["\x1b["+k+"$"] = key
		// Ctrl modifier
		key.Mod = ModCtrl
		table["\x1b["+k+"^"] = key
		// Shift-Ctrl modifier
		key.Mod = ModShift | ModCtrl
		table["\x1b["+k+"@"] = key
	}

	// URxvt F keys
	// Note: Shift + F1-F10 generates F11-F20.
	// This means Shift + F1 and Shift + F2 will generate F11 and F12, the same
	// applies to Ctrl + Shift F1 & F2.
	//
	// P.S. Don't like this? Blame URxvt, configure your terminal to use
	// different escapes like XTerm, or switch to a better terminal ¯\_(ツ)_/¯
	//
	// See https://manpages.ubuntu.com/manpages/trusty/man7/urxvt.7.html#key%20codes
	table["\x1b[23$"] = Key{Type: KeyF11, Mod: ModShift}
	table["\x1b[24$"] = Key{Type: KeyF12, Mod: ModShift}
	table["\x1b[25$"] = Key{Type: KeyF13, Mod: ModShift}
	table["\x1b[26$"] = Key{Type: KeyF14, Mod: ModShift}
	table["\x1b[28$"] = Key{Type: KeyF15, Mod: ModShift}
	table["\x1b[29$"] = Key{Type: KeyF16, Mod: ModShift}
	table["\x1b[31$"] = Key{Type: KeyF17, Mod: ModShift}
	table["\x1b[32$"] = Key{Type: KeyF18, Mod: ModShift}
	table["\x1b[33$"] = Key{Type: KeyF19, Mod: ModShift}
	table["\x1b[34$"] = Key{Type: KeyF20, Mod: ModShift}
	table["\x1b[11^"] = Key{Type: KeyF1, Mod: ModCtrl}
	table["\x1b[12^"] = Key{Type: KeyF2, Mod: ModCtrl}
	table["\x1b[13^"] = Key{Type: KeyF3, Mod: ModCtrl}
	table["\x1b[14^"] = Key{Type: KeyF4, Mod: ModCtrl}
	table["\x1b[15^"] = Key{Type: KeyF5, Mod: ModCtrl}
	table["\x1b[17^"] = Key{Type: KeyF6, Mod: ModCtrl}
	table["\x1b[18^"] = Key{Type: KeyF7, Mod: ModCtrl}
	table["\x1b[19^"] = Key{Type: KeyF8, Mod: ModCtrl}
	table["\x1b[20^"] = Key{Type: KeyF9, Mod: ModCtrl}
	table["\x1b[21^"] = Key{Type: KeyF10, Mod: ModCtrl}
	table["\x1b[23^"] = Key{Type: KeyF11, Mod: ModCtrl}
	table["\x1b[24^"] = Key{Type: KeyF12, Mod: ModCtrl}
	table["\x1b[25^"] = Key{Type: KeyF13, Mod: ModCtrl}
	table["\x1b[26^"] = Key{Type: KeyF14, Mod: ModCtrl}
	table["\x1b[28^"] = Key{Type: KeyF15, Mod: ModCtrl}
	table["\x1b[29^"] = Key{Type: KeyF16, Mod: ModCtrl}
	table["\x1b[31^"] = Key{Type: KeyF17, Mod: ModCtrl}
	table["\x1b[32^"] = Key{Type: KeyF18, Mod: ModCtrl}
	table["\x1b[33^"] = Key{Type: KeyF19, Mod: ModCtrl}
	table["\x1b[34^"] = Key{Type: KeyF20, Mod: ModCtrl}
	table["\x1b[23@"] = Key{Type: KeyF11, Mod: ModShift | ModCtrl}
	table["\x1b[24@"] = Key{Type: KeyF12, Mod: ModShift | ModCtrl}
	table["\x1b[25@"] = Key{Type: KeyF13, Mod: ModShift | ModCtrl}
	table["\x1b[26@"] = Key{Type: KeyF14, Mod: ModShift | ModCtrl}
	table["\x1b[28@"] = Key{Type: KeyF15, Mod: ModShift | ModCtrl}
	table["\x1b[29@"] = Key{Type: KeyF16, Mod: ModShift | ModCtrl}
	table["\x1b[31@"] = Key{Type: KeyF17, Mod: ModShift | ModCtrl}
	table["\x1b[32@"] = Key{Type: KeyF18, Mod: ModShift | ModCtrl}
	table["\x1b[33@"] = Key{Type: KeyF19, Mod: ModShift | ModCtrl}
	table["\x1b[34@"] = Key{Type: KeyF20, Mod: ModShift | ModCtrl}

	// Register Alt + <key> combinations
	// XXX: this must come after URxvt but before XTerm keys to register URxvt
	// keys with alt modifier
	tmap := map[string]Key{}
	for seq, key := range table {
		key := key
		key.Mod |= ModAlt
		tmap["\x1b"+seq] = key
	}
	for seq, key := range tmap {
		table[seq] = key
	}

	// XTerm modifiers
	// These are offset by 1 to be compatible with our Mod type.
	// See https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-PC-Style-Function-Keys
	modifiers := []KeyMod{
		ModShift,                              // 1
		ModAlt,                                // 2
		ModShift | ModAlt,                     // 3
		ModCtrl,                               // 4
		ModShift | ModCtrl,                    // 5
		ModAlt | ModCtrl,                      // 6
		ModShift | ModAlt | ModCtrl,           // 7
		ModMeta,                               // 8
		ModMeta | ModShift,                    // 9
		ModMeta | ModAlt,                      // 10
		ModMeta | ModShift | ModAlt,           // 11
		ModMeta | ModCtrl,                     // 12
		ModMeta | ModShift | ModCtrl,          // 13
		ModMeta | ModAlt | ModCtrl,            // 14
		ModMeta | ModShift | ModAlt | ModCtrl, // 15
	}

	// SS3 keypad function keys
	ss3FuncKeys := map[string]Key{
		// These are defined in XTerm
		// Taken from Foot keymap.h and XTerm modifyOtherKeys
		// https://codeberg.org/dnkl/foot/src/branch/master/keymap.h
		"M": {Type: KeyKpEnter}, "X": {Type: KeyKpEqual},
		"j": {Type: KeyKpMultiply}, "k": {Type: KeyKpPlus},
		"l": {Type: KeyKpComma}, "m": {Type: KeyKpMinus},
		"n": {Type: KeyKpDecimal}, "o": {Type: KeyKpDivide},
		"p": {Type: KeyKp0}, "q": {Type: KeyKp1},
		"r": {Type: KeyKp2}, "s": {Type: KeyKp3},
		"t": {Type: KeyKp4}, "u": {Type: KeyKp5},
		"v": {Type: KeyKp6}, "w": {Type: KeyKp7},
		"x": {Type: KeyKp8}, "y": {Type: KeyKp9},
	}

	// XTerm keys
	csiFuncKeys := map[string]Key{
		"A": {Type: KeyUp}, "B": {Type: KeyDown},
		"C": {Type: KeyRight}, "D": {Type: KeyLeft},
		"E": {Type: KeyBegin}, "F": {Type: KeyEnd},
		"H": {Type: KeyHome}, "P": {Type: KeyF1},
		"Q": {Type: KeyF2}, "R": {Type: KeyF3},
		"S": {Type: KeyF4},
	}

	// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
	modifyOtherKeys := map[int]Key{
		ansi.BS:  {Type: KeyBackspace},
		ansi.HT:  {Type: KeyTab},
		ansi.CR:  {Type: KeyEnter},
		ansi.ESC: {Type: KeyEscape},
		ansi.DEL: {Type: KeyBackspace},
	}

	for _, m := range modifiers {
		// XTerm modifier offset +1
		xtermMod := strconv.Itoa(int(m) + 1)

		//  CSI 1 ; <modifier> <func>
		for k, v := range csiFuncKeys {
			// Functions always have a leading 1 param
			seq := "\x1b[1;" + xtermMod + k
			key := v
			key.Mod = m
			table[seq] = key
		}
		// SS3 <modifier> <func>
		for k, v := range ss3FuncKeys {
			seq := "\x1bO" + xtermMod + k
			key := v
			key.Mod = m
			table[seq] = key
		}
		//  CSI <number> ; <modifier> ~
		for k, v := range csiTildeKeys {
			seq := "\x1b[" + k + ";" + xtermMod + "~"
			key := v
			key.Mod = m
			table[seq] = key
		}
		// CSI 27 ; <modifier> ; <code> ~
		for k, v := range modifyOtherKeys {
			code := strconv.Itoa(k)
			seq := "\x1b[27;" + xtermMod + ";" + code + "~"
			key := v
			key.Mod = m
			table[seq] = key
		}
	}

	// Register terminfo keys
	// XXX: this might override keys already registered in table
	if flags&_FlagTerminfo != 0 {
		titable := buildTerminfoKeys(flags, term)
		for seq, key := range titable {
			table[seq] = key
		}
	}

	return table
}
