package tea

import (
	"strconv"

	"github.com/charmbracelet/x/ansi"
)

// buildKeysTable builds a table of key sequences and their corresponding key
// events based on the VT100/VT200, XTerm, and Urxvt terminal specs.
// TODO: Use flags?
func buildKeysTable(flags int, term string) map[string]Key {
	nul := Key{Runes: []rune{' '}, Sym: KeySpace, Mod: ModCtrl} // ctrl+@ or ctrl+space
	if flags&_FlagCtrlAt != 0 {
		nul = Key{Runes: []rune{'@'}, Mod: ModCtrl}
	}

	tab := Key{Sym: KeyTab} // ctrl+i or tab
	if flags&_FlagCtrlI != 0 {
		tab = Key{Runes: []rune{'i'}, Mod: ModCtrl}
	}

	enter := Key{Sym: KeyEnter} // ctrl+m or enter
	if flags&_FlagCtrlM != 0 {
		enter = Key{Runes: []rune{'m'}, Mod: ModCtrl}
	}

	esc := Key{Sym: KeyEscape} // ctrl+[ or escape
	if flags&_FlagCtrlOpenBracket != 0 {
		esc = Key{Runes: []rune{'['}, Mod: ModCtrl} // ctrl+[ or escape
	}

	del := Key{Sym: KeyBackspace}
	if flags&_FlagBackspace != 0 {
		del.Sym = KeyDelete
	}

	find := Key{Sym: KeyHome}
	if flags&_FlagFind != 0 {
		find.Sym = KeyFind
	}

	sel := Key{Sym: KeyEnd}
	if flags&_FlagSelect != 0 {
		sel.Sym = KeySelect
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
		string(byte(ansi.SP)):  {Sym: KeySpace, Runes: []rune{' '}},
		string(byte(ansi.DEL)): del,

		// Special keys

		"\x1b[Z": {Sym: KeyTab, Mod: ModShift},

		"\x1b[1~": find,
		"\x1b[2~": {Sym: KeyInsert},
		"\x1b[3~": {Sym: KeyDelete},
		"\x1b[4~": sel,
		"\x1b[5~": {Sym: KeyPgUp},
		"\x1b[6~": {Sym: KeyPgDown},
		"\x1b[7~": {Sym: KeyHome},
		"\x1b[8~": {Sym: KeyEnd},

		// Normal mode
		"\x1b[A": {Sym: KeyUp},
		"\x1b[B": {Sym: KeyDown},
		"\x1b[C": {Sym: KeyRight},
		"\x1b[D": {Sym: KeyLeft},
		"\x1b[E": {Sym: KeyBegin},
		"\x1b[F": {Sym: KeyEnd},
		"\x1b[H": {Sym: KeyHome},
		"\x1b[P": {Sym: KeyF1},
		"\x1b[Q": {Sym: KeyF2},
		"\x1b[R": {Sym: KeyF3},
		"\x1b[S": {Sym: KeyF4},

		// Application Cursor Key Mode (DECCKM)
		"\x1bOA": {Sym: KeyUp},
		"\x1bOB": {Sym: KeyDown},
		"\x1bOC": {Sym: KeyRight},
		"\x1bOD": {Sym: KeyLeft},
		"\x1bOE": {Sym: KeyBegin},
		"\x1bOF": {Sym: KeyEnd},
		"\x1bOH": {Sym: KeyHome},
		"\x1bOP": {Sym: KeyF1},
		"\x1bOQ": {Sym: KeyF2},
		"\x1bOR": {Sym: KeyF3},
		"\x1bOS": {Sym: KeyF4},

		// Keypad Application Mode (DECKPAM)

		"\x1bOM": {Sym: KeyKpEnter},
		"\x1bOX": {Sym: KeyKpEqual},
		"\x1bOj": {Sym: KeyKpMultiply},
		"\x1bOk": {Sym: KeyKpPlus},
		"\x1bOl": {Sym: KeyKpComma},
		"\x1bOm": {Sym: KeyKpMinus},
		"\x1bOn": {Sym: KeyKpDecimal},
		"\x1bOo": {Sym: KeyKpDivide},
		"\x1bOp": {Sym: KeyKp0},
		"\x1bOq": {Sym: KeyKp1},
		"\x1bOr": {Sym: KeyKp2},
		"\x1bOs": {Sym: KeyKp3},
		"\x1bOt": {Sym: KeyKp4},
		"\x1bOu": {Sym: KeyKp5},
		"\x1bOv": {Sym: KeyKp6},
		"\x1bOw": {Sym: KeyKp7},
		"\x1bOx": {Sym: KeyKp8},
		"\x1bOy": {Sym: KeyKp9},

		// Function keys

		"\x1b[11~": {Sym: KeyF1},
		"\x1b[12~": {Sym: KeyF2},
		"\x1b[13~": {Sym: KeyF3},
		"\x1b[14~": {Sym: KeyF4},
		"\x1b[15~": {Sym: KeyF5},
		"\x1b[17~": {Sym: KeyF6},
		"\x1b[18~": {Sym: KeyF7},
		"\x1b[19~": {Sym: KeyF8},
		"\x1b[20~": {Sym: KeyF9},
		"\x1b[21~": {Sym: KeyF10},
		"\x1b[23~": {Sym: KeyF11},
		"\x1b[24~": {Sym: KeyF12},
		"\x1b[25~": {Sym: KeyF13},
		"\x1b[26~": {Sym: KeyF14},
		"\x1b[28~": {Sym: KeyF15},
		"\x1b[29~": {Sym: KeyF16},
		"\x1b[31~": {Sym: KeyF17},
		"\x1b[32~": {Sym: KeyF18},
		"\x1b[33~": {Sym: KeyF19},
		"\x1b[34~": {Sym: KeyF20},
	}

	// CSI ~ sequence keys
	csiTildeKeys := map[string]Key{
		"1": find, "2": {Sym: KeyInsert},
		"3": {Sym: KeyDelete}, "4": sel,
		"5": {Sym: KeyPgUp}, "6": {Sym: KeyPgDown},
		"7": {Sym: KeyHome}, "8": {Sym: KeyEnd},
		// There are no 9 and 10 keys
		"11": {Sym: KeyF1}, "12": {Sym: KeyF2},
		"13": {Sym: KeyF3}, "14": {Sym: KeyF4},
		"15": {Sym: KeyF5}, "17": {Sym: KeyF6},
		"18": {Sym: KeyF7}, "19": {Sym: KeyF8},
		"20": {Sym: KeyF9}, "21": {Sym: KeyF10},
		"23": {Sym: KeyF11}, "24": {Sym: KeyF12},
		"25": {Sym: KeyF13}, "26": {Sym: KeyF14},
		"28": {Sym: KeyF15}, "29": {Sym: KeyF16},
		"31": {Sym: KeyF17}, "32": {Sym: KeyF18},
		"33": {Sym: KeyF19}, "34": {Sym: KeyF20},
	}

	// URxvt keys
	// See https://manpages.ubuntu.com/manpages/trusty/man7/urxvt.7.html#key%20codes
	table["\x1b[a"] = Key{Sym: KeyUp, Mod: ModShift}
	table["\x1b[b"] = Key{Sym: KeyDown, Mod: ModShift}
	table["\x1b[c"] = Key{Sym: KeyRight, Mod: ModShift}
	table["\x1b[d"] = Key{Sym: KeyLeft, Mod: ModShift}
	table["\x1bOa"] = Key{Sym: KeyUp, Mod: ModCtrl}
	table["\x1bOb"] = Key{Sym: KeyDown, Mod: ModCtrl}
	table["\x1bOc"] = Key{Sym: KeyRight, Mod: ModCtrl}
	table["\x1bOd"] = Key{Sym: KeyLeft, Mod: ModCtrl}
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
	table["\x1b[23$"] = Key{Sym: KeyF11, Mod: ModShift}
	table["\x1b[24$"] = Key{Sym: KeyF12, Mod: ModShift}
	table["\x1b[25$"] = Key{Sym: KeyF13, Mod: ModShift}
	table["\x1b[26$"] = Key{Sym: KeyF14, Mod: ModShift}
	table["\x1b[28$"] = Key{Sym: KeyF15, Mod: ModShift}
	table["\x1b[29$"] = Key{Sym: KeyF16, Mod: ModShift}
	table["\x1b[31$"] = Key{Sym: KeyF17, Mod: ModShift}
	table["\x1b[32$"] = Key{Sym: KeyF18, Mod: ModShift}
	table["\x1b[33$"] = Key{Sym: KeyF19, Mod: ModShift}
	table["\x1b[34$"] = Key{Sym: KeyF20, Mod: ModShift}
	table["\x1b[11^"] = Key{Sym: KeyF1, Mod: ModCtrl}
	table["\x1b[12^"] = Key{Sym: KeyF2, Mod: ModCtrl}
	table["\x1b[13^"] = Key{Sym: KeyF3, Mod: ModCtrl}
	table["\x1b[14^"] = Key{Sym: KeyF4, Mod: ModCtrl}
	table["\x1b[15^"] = Key{Sym: KeyF5, Mod: ModCtrl}
	table["\x1b[17^"] = Key{Sym: KeyF6, Mod: ModCtrl}
	table["\x1b[18^"] = Key{Sym: KeyF7, Mod: ModCtrl}
	table["\x1b[19^"] = Key{Sym: KeyF8, Mod: ModCtrl}
	table["\x1b[20^"] = Key{Sym: KeyF9, Mod: ModCtrl}
	table["\x1b[21^"] = Key{Sym: KeyF10, Mod: ModCtrl}
	table["\x1b[23^"] = Key{Sym: KeyF11, Mod: ModCtrl}
	table["\x1b[24^"] = Key{Sym: KeyF12, Mod: ModCtrl}
	table["\x1b[25^"] = Key{Sym: KeyF13, Mod: ModCtrl}
	table["\x1b[26^"] = Key{Sym: KeyF14, Mod: ModCtrl}
	table["\x1b[28^"] = Key{Sym: KeyF15, Mod: ModCtrl}
	table["\x1b[29^"] = Key{Sym: KeyF16, Mod: ModCtrl}
	table["\x1b[31^"] = Key{Sym: KeyF17, Mod: ModCtrl}
	table["\x1b[32^"] = Key{Sym: KeyF18, Mod: ModCtrl}
	table["\x1b[33^"] = Key{Sym: KeyF19, Mod: ModCtrl}
	table["\x1b[34^"] = Key{Sym: KeyF20, Mod: ModCtrl}
	table["\x1b[23@"] = Key{Sym: KeyF11, Mod: ModShift | ModCtrl}
	table["\x1b[24@"] = Key{Sym: KeyF12, Mod: ModShift | ModCtrl}
	table["\x1b[25@"] = Key{Sym: KeyF13, Mod: ModShift | ModCtrl}
	table["\x1b[26@"] = Key{Sym: KeyF14, Mod: ModShift | ModCtrl}
	table["\x1b[28@"] = Key{Sym: KeyF15, Mod: ModShift | ModCtrl}
	table["\x1b[29@"] = Key{Sym: KeyF16, Mod: ModShift | ModCtrl}
	table["\x1b[31@"] = Key{Sym: KeyF17, Mod: ModShift | ModCtrl}
	table["\x1b[32@"] = Key{Sym: KeyF18, Mod: ModShift | ModCtrl}
	table["\x1b[33@"] = Key{Sym: KeyF19, Mod: ModShift | ModCtrl}
	table["\x1b[34@"] = Key{Sym: KeyF20, Mod: ModShift | ModCtrl}

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
		"M": {Sym: KeyKpEnter}, "X": {Sym: KeyKpEqual},
		"j": {Sym: KeyKpMultiply}, "k": {Sym: KeyKpPlus},
		"l": {Sym: KeyKpComma}, "m": {Sym: KeyKpMinus},
		"n": {Sym: KeyKpDecimal}, "o": {Sym: KeyKpDivide},
		"p": {Sym: KeyKp0}, "q": {Sym: KeyKp1},
		"r": {Sym: KeyKp2}, "s": {Sym: KeyKp3},
		"t": {Sym: KeyKp4}, "u": {Sym: KeyKp5},
		"v": {Sym: KeyKp6}, "w": {Sym: KeyKp7},
		"x": {Sym: KeyKp8}, "y": {Sym: KeyKp9},
	}

	// XTerm keys
	csiFuncKeys := map[string]Key{
		"A": {Sym: KeyUp}, "B": {Sym: KeyDown},
		"C": {Sym: KeyRight}, "D": {Sym: KeyLeft},
		"E": {Sym: KeyBegin}, "F": {Sym: KeyEnd},
		"H": {Sym: KeyHome}, "P": {Sym: KeyF1},
		"Q": {Sym: KeyF2}, "R": {Sym: KeyF3},
		"S": {Sym: KeyF4},
	}

	// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
	modifyOtherKeys := map[int]Key{
		ansi.BS:  {Sym: KeyBackspace},
		ansi.HT:  {Sym: KeyTab},
		ansi.CR:  {Sym: KeyEnter},
		ansi.ESC: {Sym: KeyEscape},
		ansi.DEL: {Sym: KeyBackspace},
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
