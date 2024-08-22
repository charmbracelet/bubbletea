package tea

import (
	"strconv"

	"github.com/charmbracelet/x/ansi"
)

// buildKeysTable builds a table of key sequences and their corresponding key
// events based on the VT100/VT200, XTerm, and Urxvt terminal specs.
// TODO: Use flags?
func buildKeysTable(flags int, term string) map[string]key {
	nul := key{runes: []rune{' '}, typ: KeySpace, mod: ModCtrl} // ctrl+@ or ctrl+space
	if flags&_FlagCtrlAt != 0 {
		nul = key{runes: []rune{'@'}, mod: ModCtrl}
	}

	tab := key{typ: KeyTab} // ctrl+i or tab
	if flags&_FlagCtrlI != 0 {
		tab = key{runes: []rune{'i'}, mod: ModCtrl}
	}

	enter := key{typ: KeyEnter} // ctrl+m or enter
	if flags&_FlagCtrlM != 0 {
		enter = key{runes: []rune{'m'}, mod: ModCtrl}
	}

	esc := key{typ: KeyEscape} // ctrl+[ or escape
	if flags&_FlagCtrlOpenBracket != 0 {
		esc = key{runes: []rune{'['}, mod: ModCtrl} // ctrl+[ or escape
	}

	del := key{typ: KeyBackspace}
	if flags&_FlagBackspace != 0 {
		del.typ = KeyDelete
	}

	find := key{typ: KeyHome}
	if flags&_FlagFind != 0 {
		find.typ = KeyFind
	}

	sel := key{typ: KeyEnd}
	if flags&_FlagSelect != 0 {
		sel.typ = KeySelect
	}

	// The following is a table of key sequences and their corresponding key
	// events based on the VT100/VT200 terminal specs.
	//
	// See: https://vt100.net/docs/vt100-ug/chapter3.html#S3.2
	// See: https://vt100.net/docs/vt220-rm/chapter3.html
	//
	// XXX: These keys may be overwritten by other options like XTerm or
	// Terminfo.
	table := map[string]key{
		// C0 control characters
		string(byte(ansi.NUL)): nul,
		string(byte(ansi.SOH)): {runes: []rune{'a'}, mod: ModCtrl},
		string(byte(ansi.STX)): {runes: []rune{'b'}, mod: ModCtrl},
		string(byte(ansi.ETX)): {runes: []rune{'c'}, mod: ModCtrl},
		string(byte(ansi.EOT)): {runes: []rune{'d'}, mod: ModCtrl},
		string(byte(ansi.ENQ)): {runes: []rune{'e'}, mod: ModCtrl},
		string(byte(ansi.ACK)): {runes: []rune{'f'}, mod: ModCtrl},
		string(byte(ansi.BEL)): {runes: []rune{'g'}, mod: ModCtrl},
		string(byte(ansi.BS)):  {runes: []rune{'h'}, mod: ModCtrl},
		string(byte(ansi.HT)):  tab,
		string(byte(ansi.LF)):  {runes: []rune{'j'}, mod: ModCtrl},
		string(byte(ansi.VT)):  {runes: []rune{'k'}, mod: ModCtrl},
		string(byte(ansi.FF)):  {runes: []rune{'l'}, mod: ModCtrl},
		string(byte(ansi.CR)):  enter,
		string(byte(ansi.SO)):  {runes: []rune{'n'}, mod: ModCtrl},
		string(byte(ansi.SI)):  {runes: []rune{'o'}, mod: ModCtrl},
		string(byte(ansi.DLE)): {runes: []rune{'p'}, mod: ModCtrl},
		string(byte(ansi.DC1)): {runes: []rune{'q'}, mod: ModCtrl},
		string(byte(ansi.DC2)): {runes: []rune{'r'}, mod: ModCtrl},
		string(byte(ansi.DC3)): {runes: []rune{'s'}, mod: ModCtrl},
		string(byte(ansi.DC4)): {runes: []rune{'t'}, mod: ModCtrl},
		string(byte(ansi.NAK)): {runes: []rune{'u'}, mod: ModCtrl},
		string(byte(ansi.SYN)): {runes: []rune{'v'}, mod: ModCtrl},
		string(byte(ansi.ETB)): {runes: []rune{'w'}, mod: ModCtrl},
		string(byte(ansi.CAN)): {runes: []rune{'x'}, mod: ModCtrl},
		string(byte(ansi.EM)):  {runes: []rune{'y'}, mod: ModCtrl},
		string(byte(ansi.SUB)): {runes: []rune{'z'}, mod: ModCtrl},
		string(byte(ansi.ESC)): esc,
		string(byte(ansi.FS)):  {runes: []rune{'\\'}, mod: ModCtrl},
		string(byte(ansi.GS)):  {runes: []rune{']'}, mod: ModCtrl},
		string(byte(ansi.RS)):  {runes: []rune{'^'}, mod: ModCtrl},
		string(byte(ansi.US)):  {runes: []rune{'_'}, mod: ModCtrl},

		// Special keys in G0
		string(byte(ansi.SP)):  {typ: KeySpace, runes: []rune{' '}},
		string(byte(ansi.DEL)): del,

		// Special keys

		"\x1b[Z": {typ: KeyTab, mod: ModShift},

		"\x1b[1~": find,
		"\x1b[2~": {typ: KeyInsert},
		"\x1b[3~": {typ: KeyDelete},
		"\x1b[4~": sel,
		"\x1b[5~": {typ: KeyPgUp},
		"\x1b[6~": {typ: KeyPgDown},
		"\x1b[7~": {typ: KeyHome},
		"\x1b[8~": {typ: KeyEnd},

		// Normal mode
		"\x1b[A": {typ: KeyUp},
		"\x1b[B": {typ: KeyDown},
		"\x1b[C": {typ: KeyRight},
		"\x1b[D": {typ: KeyLeft},
		"\x1b[E": {typ: KeyBegin},
		"\x1b[F": {typ: KeyEnd},
		"\x1b[H": {typ: KeyHome},
		"\x1b[P": {typ: KeyF1},
		"\x1b[Q": {typ: KeyF2},
		"\x1b[R": {typ: KeyF3},
		"\x1b[S": {typ: KeyF4},

		// Application Cursor Key Mode (DECCKM)
		"\x1bOA": {typ: KeyUp},
		"\x1bOB": {typ: KeyDown},
		"\x1bOC": {typ: KeyRight},
		"\x1bOD": {typ: KeyLeft},
		"\x1bOE": {typ: KeyBegin},
		"\x1bOF": {typ: KeyEnd},
		"\x1bOH": {typ: KeyHome},
		"\x1bOP": {typ: KeyF1},
		"\x1bOQ": {typ: KeyF2},
		"\x1bOR": {typ: KeyF3},
		"\x1bOS": {typ: KeyF4},

		// Keypad Application Mode (DECKPAM)

		"\x1bOM": {typ: KeyKpEnter},
		"\x1bOX": {typ: KeyKpEqual},
		"\x1bOj": {typ: KeyKpMultiply},
		"\x1bOk": {typ: KeyKpPlus},
		"\x1bOl": {typ: KeyKpComma},
		"\x1bOm": {typ: KeyKpMinus},
		"\x1bOn": {typ: KeyKpDecimal},
		"\x1bOo": {typ: KeyKpDivide},
		"\x1bOp": {typ: KeyKp0},
		"\x1bOq": {typ: KeyKp1},
		"\x1bOr": {typ: KeyKp2},
		"\x1bOs": {typ: KeyKp3},
		"\x1bOt": {typ: KeyKp4},
		"\x1bOu": {typ: KeyKp5},
		"\x1bOv": {typ: KeyKp6},
		"\x1bOw": {typ: KeyKp7},
		"\x1bOx": {typ: KeyKp8},
		"\x1bOy": {typ: KeyKp9},

		// Function keys

		"\x1b[11~": {typ: KeyF1},
		"\x1b[12~": {typ: KeyF2},
		"\x1b[13~": {typ: KeyF3},
		"\x1b[14~": {typ: KeyF4},
		"\x1b[15~": {typ: KeyF5},
		"\x1b[17~": {typ: KeyF6},
		"\x1b[18~": {typ: KeyF7},
		"\x1b[19~": {typ: KeyF8},
		"\x1b[20~": {typ: KeyF9},
		"\x1b[21~": {typ: KeyF10},
		"\x1b[23~": {typ: KeyF11},
		"\x1b[24~": {typ: KeyF12},
		"\x1b[25~": {typ: KeyF13},
		"\x1b[26~": {typ: KeyF14},
		"\x1b[28~": {typ: KeyF15},
		"\x1b[29~": {typ: KeyF16},
		"\x1b[31~": {typ: KeyF17},
		"\x1b[32~": {typ: KeyF18},
		"\x1b[33~": {typ: KeyF19},
		"\x1b[34~": {typ: KeyF20},
	}

	// CSI ~ sequence keys
	csiTildeKeys := map[string]key{
		"1": find, "2": {typ: KeyInsert},
		"3": {typ: KeyDelete}, "4": sel,
		"5": {typ: KeyPgUp}, "6": {typ: KeyPgDown},
		"7": {typ: KeyHome}, "8": {typ: KeyEnd},
		// There are no 9 and 10 keys
		"11": {typ: KeyF1}, "12": {typ: KeyF2},
		"13": {typ: KeyF3}, "14": {typ: KeyF4},
		"15": {typ: KeyF5}, "17": {typ: KeyF6},
		"18": {typ: KeyF7}, "19": {typ: KeyF8},
		"20": {typ: KeyF9}, "21": {typ: KeyF10},
		"23": {typ: KeyF11}, "24": {typ: KeyF12},
		"25": {typ: KeyF13}, "26": {typ: KeyF14},
		"28": {typ: KeyF15}, "29": {typ: KeyF16},
		"31": {typ: KeyF17}, "32": {typ: KeyF18},
		"33": {typ: KeyF19}, "34": {typ: KeyF20},
	}

	// URxvt keys
	// See https://manpages.ubuntu.com/manpages/trusty/man7/urxvt.7.html#key%20codes
	table["\x1b[a"] = key{typ: KeyUp, mod: ModShift}
	table["\x1b[b"] = key{typ: KeyDown, mod: ModShift}
	table["\x1b[c"] = key{typ: KeyRight, mod: ModShift}
	table["\x1b[d"] = key{typ: KeyLeft, mod: ModShift}
	table["\x1bOa"] = key{typ: KeyUp, mod: ModCtrl}
	table["\x1bOb"] = key{typ: KeyDown, mod: ModCtrl}
	table["\x1bOc"] = key{typ: KeyRight, mod: ModCtrl}
	table["\x1bOd"] = key{typ: KeyLeft, mod: ModCtrl}
	// TODO: invistigate if shift-ctrl arrow keys collide with DECCKM keys i.e.
	// "\x1bOA", "\x1bOB", "\x1bOC", "\x1bOD"

	// URxvt modifier CSI ~ keys
	for k, v := range csiTildeKeys {
		key := v
		// Normal (no modifier) already defined part of VT100/VT200
		// Shift modifier
		key.mod = ModShift
		table["\x1b["+k+"$"] = key
		// Ctrl modifier
		key.mod = ModCtrl
		table["\x1b["+k+"^"] = key
		// Shift-Ctrl modifier
		key.mod = ModShift | ModCtrl
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
	table["\x1b[23$"] = key{typ: KeyF11, mod: ModShift}
	table["\x1b[24$"] = key{typ: KeyF12, mod: ModShift}
	table["\x1b[25$"] = key{typ: KeyF13, mod: ModShift}
	table["\x1b[26$"] = key{typ: KeyF14, mod: ModShift}
	table["\x1b[28$"] = key{typ: KeyF15, mod: ModShift}
	table["\x1b[29$"] = key{typ: KeyF16, mod: ModShift}
	table["\x1b[31$"] = key{typ: KeyF17, mod: ModShift}
	table["\x1b[32$"] = key{typ: KeyF18, mod: ModShift}
	table["\x1b[33$"] = key{typ: KeyF19, mod: ModShift}
	table["\x1b[34$"] = key{typ: KeyF20, mod: ModShift}
	table["\x1b[11^"] = key{typ: KeyF1, mod: ModCtrl}
	table["\x1b[12^"] = key{typ: KeyF2, mod: ModCtrl}
	table["\x1b[13^"] = key{typ: KeyF3, mod: ModCtrl}
	table["\x1b[14^"] = key{typ: KeyF4, mod: ModCtrl}
	table["\x1b[15^"] = key{typ: KeyF5, mod: ModCtrl}
	table["\x1b[17^"] = key{typ: KeyF6, mod: ModCtrl}
	table["\x1b[18^"] = key{typ: KeyF7, mod: ModCtrl}
	table["\x1b[19^"] = key{typ: KeyF8, mod: ModCtrl}
	table["\x1b[20^"] = key{typ: KeyF9, mod: ModCtrl}
	table["\x1b[21^"] = key{typ: KeyF10, mod: ModCtrl}
	table["\x1b[23^"] = key{typ: KeyF11, mod: ModCtrl}
	table["\x1b[24^"] = key{typ: KeyF12, mod: ModCtrl}
	table["\x1b[25^"] = key{typ: KeyF13, mod: ModCtrl}
	table["\x1b[26^"] = key{typ: KeyF14, mod: ModCtrl}
	table["\x1b[28^"] = key{typ: KeyF15, mod: ModCtrl}
	table["\x1b[29^"] = key{typ: KeyF16, mod: ModCtrl}
	table["\x1b[31^"] = key{typ: KeyF17, mod: ModCtrl}
	table["\x1b[32^"] = key{typ: KeyF18, mod: ModCtrl}
	table["\x1b[33^"] = key{typ: KeyF19, mod: ModCtrl}
	table["\x1b[34^"] = key{typ: KeyF20, mod: ModCtrl}
	table["\x1b[23@"] = key{typ: KeyF11, mod: ModShift | ModCtrl}
	table["\x1b[24@"] = key{typ: KeyF12, mod: ModShift | ModCtrl}
	table["\x1b[25@"] = key{typ: KeyF13, mod: ModShift | ModCtrl}
	table["\x1b[26@"] = key{typ: KeyF14, mod: ModShift | ModCtrl}
	table["\x1b[28@"] = key{typ: KeyF15, mod: ModShift | ModCtrl}
	table["\x1b[29@"] = key{typ: KeyF16, mod: ModShift | ModCtrl}
	table["\x1b[31@"] = key{typ: KeyF17, mod: ModShift | ModCtrl}
	table["\x1b[32@"] = key{typ: KeyF18, mod: ModShift | ModCtrl}
	table["\x1b[33@"] = key{typ: KeyF19, mod: ModShift | ModCtrl}
	table["\x1b[34@"] = key{typ: KeyF20, mod: ModShift | ModCtrl}

	// Register Alt + <key> combinations
	// XXX: this must come after URxvt but before XTerm keys to register URxvt
	// keys with alt modifier
	tmap := map[string]key{}
	for seq, key := range table {
		key := key
		key.mod |= ModAlt
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
	ss3FuncKeys := map[string]key{
		// These are defined in XTerm
		// Taken from Foot keymap.h and XTerm modifyOtherKeys
		// https://codeberg.org/dnkl/foot/src/branch/master/keymap.h
		"M": {typ: KeyKpEnter}, "X": {typ: KeyKpEqual},
		"j": {typ: KeyKpMultiply}, "k": {typ: KeyKpPlus},
		"l": {typ: KeyKpComma}, "m": {typ: KeyKpMinus},
		"n": {typ: KeyKpDecimal}, "o": {typ: KeyKpDivide},
		"p": {typ: KeyKp0}, "q": {typ: KeyKp1},
		"r": {typ: KeyKp2}, "s": {typ: KeyKp3},
		"t": {typ: KeyKp4}, "u": {typ: KeyKp5},
		"v": {typ: KeyKp6}, "w": {typ: KeyKp7},
		"x": {typ: KeyKp8}, "y": {typ: KeyKp9},
	}

	// XTerm keys
	csiFuncKeys := map[string]key{
		"A": {typ: KeyUp}, "B": {typ: KeyDown},
		"C": {typ: KeyRight}, "D": {typ: KeyLeft},
		"E": {typ: KeyBegin}, "F": {typ: KeyEnd},
		"H": {typ: KeyHome}, "P": {typ: KeyF1},
		"Q": {typ: KeyF2}, "R": {typ: KeyF3},
		"S": {typ: KeyF4},
	}

	// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
	modifyOtherKeys := map[int]key{
		ansi.BS:  {typ: KeyBackspace},
		ansi.HT:  {typ: KeyTab},
		ansi.CR:  {typ: KeyEnter},
		ansi.ESC: {typ: KeyEscape},
		ansi.DEL: {typ: KeyBackspace},
	}

	for _, m := range modifiers {
		// XTerm modifier offset +1
		xtermMod := strconv.Itoa(int(m) + 1)

		//  CSI 1 ; <modifier> <func>
		for k, v := range csiFuncKeys {
			// Functions always have a leading 1 param
			seq := "\x1b[1;" + xtermMod + k
			key := v
			key.mod = m
			table[seq] = key
		}
		// SS3 <modifier> <func>
		for k, v := range ss3FuncKeys {
			seq := "\x1bO" + xtermMod + k
			key := v
			key.mod = m
			table[seq] = key
		}
		//  CSI <number> ; <modifier> ~
		for k, v := range csiTildeKeys {
			seq := "\x1b[" + k + ";" + xtermMod + "~"
			key := v
			key.mod = m
			table[seq] = key
		}
		// CSI 27 ; <modifier> ; <code> ~
		for k, v := range modifyOtherKeys {
			code := strconv.Itoa(k)
			seq := "\x1b[27;" + xtermMod + ";" + code + "~"
			key := v
			key.mod = m
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
