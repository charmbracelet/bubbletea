package tea

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/mattn/go-localereader"
)

// KeyMsg contains information about a keypress. KeyMsgs are always sent to
// the program's update function. There are a couple general patterns you could
// use to check for keypresses:
//
//	// Switch on the string representation of the key (shorter)
//	switch msg := msg.(type) {
//	case KeyMsg:
//	    switch msg.String() {
//	    case "enter":
//	        fmt.Println("you pressed enter!")
//	    case "a":
//	        fmt.Println("you pressed a!")
//	    }
//	}
//
//	// Switch on the key type (more foolproof)
//	switch msg := msg.(type) {
//	case KeyMsg:
//	    switch msg.Type {
//	    case KeyEnter:
//	        fmt.Println("you pressed enter!")
//	    case KeyRunes:
//	        switch string(msg.Runes) {
//	        case "a":
//	            fmt.Println("you pressed a!")
//	        }
//	    }
//	}
//
// Note that Key.Runes will always contain at least one character, so you can
// always safely call Key.Runes[0]. In most cases Key.Runes will only contain
// one character, though certain input method editors (most notably Chinese
// IMEs) can input multiple runes at once.
type KeyMsg Key

// String returns a string representation for a key message. It's safe (and
// encouraged) for use in key comparison.
func (k KeyMsg) String() (str string) {
	return Key(k).String()
}

// Key contains information about a keypress.
type Key struct {
	Type  KeyType
	Runes []rune
	Alt   bool
}

// String returns a friendly string representation for a key. It's safe (and
// encouraged) for use in key comparison.
//
//	k := Key{Type: KeyEnter}
//	fmt.Println(k)
//	// Output: enter
func (k Key) String() (str string) {
	if k.Alt {
		str += "alt+"
	}
	if k.Type == KeyRunes {
		str += string(k.Runes)
		return str
	} else if s, ok := keyNames[k.Type]; ok {
		str += s
		return str
	}
	return ""
}

// KeyType indicates the key pressed, such as KeyEnter or KeyBreak or KeyCtrlC.
// All other keys will be type KeyRunes. To get the rune value, check the Rune
// method on a Key struct, or use the Key.String() method:
//
//	k := Key{Type: KeyRunes, Runes: []rune{'a'}, Alt: true}
//	if k.Type == KeyRunes {
//
//	    fmt.Println(k.Runes)
//	    // Output: a
//
//	    fmt.Println(k.String())
//	    // Output: alt+a
//
//	}
type KeyType int

func (k KeyType) String() (str string) {
	if s, ok := keyNames[k]; ok {
		return s
	}
	return ""
}

// Control keys. We could do this with an iota, but the values are very
// specific, so we set the values explicitly to avoid any confusion.
//
// See also:
// https://en.wikipedia.org/wiki/C0_and_C1_control_codes
const (
	keyNUL KeyType = 0   // null, \0
	keySOH KeyType = 1   // start of heading
	keySTX KeyType = 2   // start of text
	keyETX KeyType = 3   // break, ctrl+c
	keyEOT KeyType = 4   // end of transmission
	keyENQ KeyType = 5   // enquiry
	keyACK KeyType = 6   // acknowledge
	keyBEL KeyType = 7   // bell, \a
	keyBS  KeyType = 8   // backspace
	keyHT  KeyType = 9   // horizontal tabulation, \t
	keyLF  KeyType = 10  // line feed, \n
	keyVT  KeyType = 11  // vertical tabulation \v
	keyFF  KeyType = 12  // form feed \f
	keyCR  KeyType = 13  // carriage return, \r
	keySO  KeyType = 14  // shift out
	keySI  KeyType = 15  // shift in
	keyDLE KeyType = 16  // data link escape
	keyDC1 KeyType = 17  // device control one
	keyDC2 KeyType = 18  // device control two
	keyDC3 KeyType = 19  // device control three
	keyDC4 KeyType = 20  // device control four
	keyNAK KeyType = 21  // negative acknowledge
	keySYN KeyType = 22  // synchronous idle
	keyETB KeyType = 23  // end of transmission block
	keyCAN KeyType = 24  // cancel
	keyEM  KeyType = 25  // end of medium
	keySUB KeyType = 26  // substitution
	keyESC KeyType = 27  // escape, \e
	keyFS  KeyType = 28  // file separator
	keyGS  KeyType = 29  // group separator
	keyRS  KeyType = 30  // record separator
	keyUS  KeyType = 31  // unit separator
	keyDEL KeyType = 127 // delete. on most systems this is mapped to backspace, I hear
)

// Control key aliases.
const (
	KeyNull      KeyType = keyNUL
	KeyBreak     KeyType = keyETX
	KeyEnter     KeyType = keyCR
	KeyBackspace KeyType = keyDEL
	KeyTab       KeyType = keyHT
	KeyEsc       KeyType = keyESC
	KeyEscape    KeyType = keyESC

	KeyCtrlAt           KeyType = keyNUL // ctrl+@
	KeyCtrlA            KeyType = keySOH
	KeyCtrlB            KeyType = keySTX
	KeyCtrlC            KeyType = keyETX
	KeyCtrlD            KeyType = keyEOT
	KeyCtrlE            KeyType = keyENQ
	KeyCtrlF            KeyType = keyACK
	KeyCtrlG            KeyType = keyBEL
	KeyCtrlH            KeyType = keyBS
	KeyCtrlI            KeyType = keyHT
	KeyCtrlJ            KeyType = keyLF
	KeyCtrlK            KeyType = keyVT
	KeyCtrlL            KeyType = keyFF
	KeyCtrlM            KeyType = keyCR
	KeyCtrlN            KeyType = keySO
	KeyCtrlO            KeyType = keySI
	KeyCtrlP            KeyType = keyDLE
	KeyCtrlQ            KeyType = keyDC1
	KeyCtrlR            KeyType = keyDC2
	KeyCtrlS            KeyType = keyDC3
	KeyCtrlT            KeyType = keyDC4
	KeyCtrlU            KeyType = keyNAK
	KeyCtrlV            KeyType = keySYN
	KeyCtrlW            KeyType = keyETB
	KeyCtrlX            KeyType = keyCAN
	KeyCtrlY            KeyType = keyEM
	KeyCtrlZ            KeyType = keySUB
	KeyCtrlOpenBracket  KeyType = keyESC // ctrl+[
	KeyCtrlBackslash    KeyType = keyFS  // ctrl+\
	KeyCtrlCloseBracket KeyType = keyGS  // ctrl+]
	KeyCtrlCaret        KeyType = keyRS  // ctrl+^
	KeyCtrlUnderscore   KeyType = keyUS  // ctrl+_
	KeyCtrlQuestionMark KeyType = keyDEL // ctrl+?
)

// Other keys.
const (
	KeyRunes KeyType = -(iota + 1)
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyShiftTab
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDown
	KeyDelete
	KeySpace
	KeyCtrlUp
	KeyCtrlDown
	KeyCtrlRight
	KeyCtrlLeft
	KeyShiftUp
	KeyShiftDown
	KeyShiftRight
	KeyShiftLeft
	KeyCtrlShiftUp
	KeyCtrlShiftDown
	KeyCtrlShiftLeft
	KeyCtrlShiftRight
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
)

// Mappings for control keys and other special keys to friendly consts.
var keyNames = map[KeyType]string{
	// Control keys.
	keyNUL: "ctrl+@", // also ctrl+` (that's ctrl+backtick)
	keySOH: "ctrl+a",
	keySTX: "ctrl+b",
	keyETX: "ctrl+c",
	keyEOT: "ctrl+d",
	keyENQ: "ctrl+e",
	keyACK: "ctrl+f",
	keyBEL: "ctrl+g",
	keyBS:  "ctrl+h",
	keyHT:  "tab", // also ctrl+i
	keyLF:  "ctrl+j",
	keyVT:  "ctrl+k",
	keyFF:  "ctrl+l",
	keyCR:  "enter",
	keySO:  "ctrl+n",
	keySI:  "ctrl+o",
	keyDLE: "ctrl+p",
	keyDC1: "ctrl+q",
	keyDC2: "ctrl+r",
	keyDC3: "ctrl+s",
	keyDC4: "ctrl+t",
	keyNAK: "ctrl+u",
	keySYN: "ctrl+v",
	keyETB: "ctrl+w",
	keyCAN: "ctrl+x",
	keyEM:  "ctrl+y",
	keySUB: "ctrl+z",
	keyESC: "esc",
	keyFS:  "ctrl+\\",
	keyGS:  "ctrl+]",
	keyRS:  "ctrl+^",
	keyUS:  "ctrl+_",
	keyDEL: "backspace",

	// Other keys.
	KeyRunes:          "runes",
	KeyUp:             "up",
	KeyDown:           "down",
	KeyRight:          "right",
	KeySpace:          " ", // for backwards compatibility
	KeyLeft:           "left",
	KeyShiftTab:       "shift+tab",
	KeyHome:           "home",
	KeyEnd:            "end",
	KeyPgUp:           "pgup",
	KeyPgDown:         "pgdown",
	KeyDelete:         "delete",
	KeyCtrlUp:         "ctrl+up",
	KeyCtrlDown:       "ctrl+down",
	KeyCtrlRight:      "ctrl+right",
	KeyCtrlLeft:       "ctrl+left",
	KeyShiftUp:        "shift+up",
	KeyShiftDown:      "shift+down",
	KeyShiftRight:     "shift+right",
	KeyShiftLeft:      "shift+left",
	KeyCtrlShiftUp:    "ctrl+shift+up",
	KeyCtrlShiftDown:  "ctrl+shift+down",
	KeyCtrlShiftLeft:  "ctrl+shift+left",
	KeyCtrlShiftRight: "ctrl+shift+right",
	KeyF1:             "f1",
	KeyF2:             "f2",
	KeyF3:             "f3",
	KeyF4:             "f4",
	KeyF5:             "f5",
	KeyF6:             "f6",
	KeyF7:             "f7",
	KeyF8:             "f8",
	KeyF9:             "f9",
	KeyF10:            "f10",
	KeyF11:            "f11",
	KeyF12:            "f12",
	KeyF13:            "f13",
	KeyF14:            "f14",
	KeyF15:            "f15",
	KeyF16:            "f16",
	KeyF17:            "f17",
	KeyF18:            "f18",
	KeyF19:            "f19",
	KeyF20:            "f20",
}

// Sequence mappings.
var sequences = map[string]Key{
	// Arrow keys
	"\x1b[A":     {Type: KeyUp},
	"\x1b[B":     {Type: KeyDown},
	"\x1b[C":     {Type: KeyRight},
	"\x1b[D":     {Type: KeyLeft},
	"\x1b[1;2A":  {Type: KeyShiftUp},
	"\x1b[1;2B":  {Type: KeyShiftDown},
	"\x1b[1;2C":  {Type: KeyShiftRight},
	"\x1b[1;2D":  {Type: KeyShiftLeft},
	"\x1b[OA":    {Type: KeyShiftUp},    // DECCKM
	"\x1b[OB":    {Type: KeyShiftDown},  // DECCKM
	"\x1b[OC":    {Type: KeyShiftRight}, // DECCKM
	"\x1b[OD":    {Type: KeyShiftLeft},  // DECCKM
	"\x1b[a":     {Type: KeyShiftUp},    // urxvt
	"\x1b[b":     {Type: KeyShiftDown},  // urxvt
	"\x1b[c":     {Type: KeyShiftRight}, // urxvt
	"\x1b[d":     {Type: KeyShiftLeft},  // urxvt
	"\x1b[1;3A":  {Type: KeyUp, Alt: true},
	"\x1b[1;3B":  {Type: KeyDown, Alt: true},
	"\x1b[1;3C":  {Type: KeyRight, Alt: true},
	"\x1b[1;3D":  {Type: KeyLeft, Alt: true},
	"\x1b\x1b[A": {Type: KeyUp, Alt: true},    // urxvt
	"\x1b\x1b[B": {Type: KeyDown, Alt: true},  // urxvt
	"\x1b\x1b[C": {Type: KeyRight, Alt: true}, // urxvt
	"\x1b\x1b[D": {Type: KeyLeft, Alt: true},  // urxvt
	"\x1b[1;4A":  {Type: KeyShiftUp, Alt: true},
	"\x1b[1;4B":  {Type: KeyShiftDown, Alt: true},
	"\x1b[1;4C":  {Type: KeyShiftRight, Alt: true},
	"\x1b[1;4D":  {Type: KeyShiftLeft, Alt: true},
	"\x1b\x1b[a": {Type: KeyShiftUp, Alt: true},    // urxvt
	"\x1b\x1b[b": {Type: KeyShiftDown, Alt: true},  // urxvt
	"\x1b\x1b[c": {Type: KeyShiftRight, Alt: true}, // urxvt
	"\x1b\x1b[d": {Type: KeyShiftLeft, Alt: true},  // urxvt
	"\x1b[1;5A":  {Type: KeyCtrlUp},
	"\x1b[1;5B":  {Type: KeyCtrlDown},
	"\x1b[1;5C":  {Type: KeyCtrlRight},
	"\x1b[1;5D":  {Type: KeyCtrlLeft},
	"\x1b[Oa":    {Type: KeyCtrlUp, Alt: true},    // urxvt
	"\x1b[Ob":    {Type: KeyCtrlDown, Alt: true},  // urxvt
	"\x1b[Oc":    {Type: KeyCtrlRight, Alt: true}, // urxvt
	"\x1b[Od":    {Type: KeyCtrlLeft, Alt: true},  // urxvt
	"\x1b[1;6A":  {Type: KeyCtrlShiftUp},
	"\x1b[1;6B":  {Type: KeyCtrlShiftDown},
	"\x1b[1;6C":  {Type: KeyCtrlShiftRight},
	"\x1b[1;6D":  {Type: KeyCtrlShiftLeft},
	"\x1b[1;7A":  {Type: KeyCtrlUp, Alt: true},
	"\x1b[1;7B":  {Type: KeyCtrlDown, Alt: true},
	"\x1b[1;7C":  {Type: KeyCtrlRight, Alt: true},
	"\x1b[1;7D":  {Type: KeyCtrlLeft, Alt: true},
	"\x1b[1;8A":  {Type: KeyCtrlShiftUp, Alt: true},
	"\x1b[1;8B":  {Type: KeyCtrlShiftDown, Alt: true},
	"\x1b[1;8C":  {Type: KeyCtrlShiftRight, Alt: true},
	"\x1b[1;8D":  {Type: KeyCtrlShiftLeft, Alt: true},

	// Miscellaneous keys
	"\x1b[Z":      {Type: KeyShiftTab},
	"\x1b[3~":     {Type: KeyDelete},
	"\x1b[3;3~":   {Type: KeyDelete, Alt: true},
	"\x1b[1~":     {Type: KeyHome},
	"\x1b[1;3H~":  {Type: KeyHome, Alt: true},
	"\x1b[4~":     {Type: KeyEnd},
	"\x1b[1;3F~":  {Type: KeyEnd, Alt: true},
	"\x1b[5~":     {Type: KeyPgUp},
	"\x1b[5;3~":   {Type: KeyPgUp, Alt: true},
	"\x1b[6~":     {Type: KeyPgDown},
	"\x1b[6;3~":   {Type: KeyPgDown, Alt: true},
	"\x1b[7~":     {Type: KeyHome},              // urxvt
	"\x1b[8~":     {Type: KeyEnd},               // urxvt
	"\x1b\x1b[3~": {Type: KeyDelete, Alt: true}, // urxvt
	"\x1b\x1b[5~": {Type: KeyPgUp, Alt: true},   // urxvt
	"\x1b\x1b[6~": {Type: KeyPgDown, Alt: true}, // urxvt
	"\x1b\x1b[7~": {Type: KeyHome, Alt: true},   // urxvt
	"\x1b\x1b[8~": {Type: KeyEnd, Alt: true},    // urxvt

	// Function keys, Linux console
	"\x1b[[A": {Type: KeyF1}, // linux console
	"\x1b[[B": {Type: KeyF2}, // linux console
	"\x1b[[C": {Type: KeyF3}, // linux console
	"\x1b[[D": {Type: KeyF4}, // linux console
	"\x1b[[E": {Type: KeyF5}, // linux console

	// Function keys, X11
	"\x1bOP":     {Type: KeyF1},  // vt100
	"\x1bOQ":     {Type: KeyF2},  // vt100
	"\x1bOR":     {Type: KeyF3},  // vt100
	"\x1bOS":     {Type: KeyF4},  // vt100
	"\x1b[15~":   {Type: KeyF5},  // also urxvt
	"\x1b[17~":   {Type: KeyF6},  // also urxvt
	"\x1b[18~":   {Type: KeyF7},  // also urxvt
	"\x1b[19~":   {Type: KeyF8},  // also urxvt
	"\x1b[20~":   {Type: KeyF9},  // also urxvt
	"\x1b[21~":   {Type: KeyF10}, // also urxvt
	"\x1b[23~":   {Type: KeyF11}, // also urxvt
	"\x1b[24~":   {Type: KeyF12}, // also urxvt
	"\x1b[1;2P":  {Type: KeyF13},
	"\x1b[1;2Q":  {Type: KeyF14},
	"\x1b[1;2R":  {Type: KeyF15},
	"\x1b[1;2S":  {Type: KeyF16},
	"\x1b[15;2~": {Type: KeyF17},
	"\x1b[17;2~": {Type: KeyF18},
	"\x1b[18;2~": {Type: KeyF19},
	"\x1b[19;2~": {Type: KeyF20},

	// Function keys with the alt modifier, X11
	"\x1b[1;3P":  {Type: KeyF1, Alt: true},
	"\x1b[1;3Q":  {Type: KeyF2, Alt: true},
	"\x1b[1;3R":  {Type: KeyF3, Alt: true},
	"\x1b[1;3S":  {Type: KeyF4, Alt: true},
	"\x1b[15;3~": {Type: KeyF5, Alt: true},
	"\x1b[17;3~": {Type: KeyF6, Alt: true},
	"\x1b[18;3~": {Type: KeyF7, Alt: true},
	"\x1b[19;3~": {Type: KeyF8, Alt: true},
	"\x1b[20;3~": {Type: KeyF9, Alt: true},
	"\x1b[21;3~": {Type: KeyF10, Alt: true},
	"\x1b[23;3~": {Type: KeyF11, Alt: true},
	"\x1b[24;3~": {Type: KeyF12, Alt: true},

	// Function keys, urxvt
	"\x1b[11~": {Type: KeyF1},
	"\x1b[12~": {Type: KeyF2},
	"\x1b[13~": {Type: KeyF3},
	"\x1b[14~": {Type: KeyF4},
	"\x1b[25~": {Type: KeyF13},
	"\x1b[26~": {Type: KeyF14},
	"\x1b[28~": {Type: KeyF15},
	"\x1b[29~": {Type: KeyF16},
	"\x1b[31~": {Type: KeyF17},
	"\x1b[32~": {Type: KeyF18},
	"\x1b[33~": {Type: KeyF19},
	"\x1b[34~": {Type: KeyF20},

	// Function keys with the alt modifier, urxvt
	"\x1b\x1b[11~": {Type: KeyF1, Alt: true},
	"\x1b\x1b[12~": {Type: KeyF2, Alt: true},
	"\x1b\x1b[13~": {Type: KeyF3, Alt: true},
	"\x1b\x1b[14~": {Type: KeyF4, Alt: true},
	"\x1b\x1b[25~": {Type: KeyF13, Alt: true},
	"\x1b\x1b[26~": {Type: KeyF14, Alt: true},
	"\x1b\x1b[28~": {Type: KeyF15, Alt: true},
	"\x1b\x1b[29~": {Type: KeyF16, Alt: true},
	"\x1b\x1b[31~": {Type: KeyF17, Alt: true},
	"\x1b\x1b[32~": {Type: KeyF18, Alt: true},
	"\x1b\x1b[33~": {Type: KeyF19, Alt: true},
	"\x1b\x1b[34~": {Type: KeyF20, Alt: true},
}

// Hex code mappings.
var hexes = map[string]Key{
	"1b0d": {Type: KeyEnter, Alt: true},
	"1b7f": {Type: KeyBackspace, Alt: true},

	// Powershell
	"1b4f41": {Type: KeyUp, Alt: false},
	"1b4f42": {Type: KeyDown, Alt: false},
	"1b4f43": {Type: KeyRight, Alt: false},
	"1b4f44": {Type: KeyLeft, Alt: false},
}

// readInputs reads keypress and mouse inputs from a TTY and returns messages
// containing information about the key or mouse events accordingly.
func readInputs(input io.Reader) ([]Msg, error) {
	var buf [256]byte

	// Read and block
	numBytes, err := input.Read(buf[:])
	if err != nil {
		return nil, err
	}
	b := buf[:numBytes]
	b, err = localereader.UTF8(b)
	if err != nil {
		return nil, err
	}

	// Check if it's a mouse event. For now we're parsing X10-type mouse events
	// only.
	mouseEvent, err := parseX10MouseEvents(b)
	if err == nil {
		var m []Msg
		for _, v := range mouseEvent {
			m = append(m, MouseMsg(v))
		}
		return m, nil
	}

	var runeSets [][]rune
	var runes []rune

	// Translate input into runes. In most cases we'll receive exactly one
	// rune, but there are cases, particularly when an input method editor is
	// used, where we can receive multiple runes at once.
	for i, w := 0, 0; i < len(b); i += w {
		r, width := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError {
			return nil, errors.New("could not decode rune")
		}

		if r == '\x1b' && len(runes) > 1 {
			// a new key sequence has started
			runeSets = append(runeSets, runes)
			runes = []rune{}
		}

		runes = append(runes, r)
		w = width
	}
	// add the final set of runes we decoded
	runeSets = append(runeSets, runes)

	if len(runeSets) == 0 {
		return nil, errors.New("received 0 runes from input")
	}

	var msgs []Msg
	for _, runes := range runeSets {
		// Is it a sequence, like an arrow key?
		if k, ok := sequences[string(runes)]; ok {
			msgs = append(msgs, KeyMsg(k))
			continue
		}

		// Some of these need special handling.
		hex := fmt.Sprintf("%x", runes)
		if k, ok := hexes[hex]; ok {
			msgs = append(msgs, KeyMsg(k))
			continue
		}

		// Is the alt key pressed? If so, the buffer will be prefixed with an
		// escape.
		alt := false
		if len(runes) > 1 && runes[0] == 0x1b {
			alt = true
			runes = runes[1:]
		}

		for _, v := range runes {
			// Is the first rune a control character?
			r := KeyType(v)
			if r <= keyUS || r == keyDEL {
				msgs = append(msgs, KeyMsg(Key{Type: r, Alt: alt}))
				continue
			}

			// If it's a space, override the type with KeySpace (but still include
			// the rune).
			if r == ' ' {
				msgs = append(msgs, KeyMsg(Key{Type: KeySpace, Runes: []rune{v}, Alt: alt}))
				continue
			}

			// Welp, just regular, ol' runes.
			msgs = append(msgs, KeyMsg(Key{Type: KeyRunes, Runes: []rune{v}, Alt: alt}))
		}
	}

	return msgs, nil
}
