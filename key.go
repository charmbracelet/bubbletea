package tea

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

// KeyMsg contains information about a keypress. KeyMsgs are always sent to
// the program's update function. There are a couple general patterns you could
// use to check for keypresses:
//
//     // Switch on the string representation of the key (shorter)
//     switch msg := msg.(type) {
//     case KeyMsg:
//         switch msg.String() {
//         case "enter":
//             fmt.Println("you pressed enter!")
//         case "a":
//             fmt.Println("you pressed a!")
//         }
//     }
//
//     // Switch on the key type (more foolproof)
//     switch msg := msg.(type) {
//     case KeyMsg:
//         switch msg.Type {
//         case KeyEnter:
//             fmt.Println("you pressed enter!")
//         case KeyRunes:
//             switch string(msg.Runes) {
//             case "a":
//                 fmt.Println("you pressed a!")
//             }
//         }
//     }
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
//     k := Key{Type: KeyEnter}
//     fmt.Println(k)
//     // Output: enter
//
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
//     k := Key{Type: KeyRunes, Runes: []rune{'a'}, Alt: true}
//     if k.Type == KeyRunes {
//
//         fmt.Println(k.Runes)
//         // Output: a
//
//         fmt.Println(k.String())
//         // Output: alt+a
//
//     }
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
	keySP  KeyType = 32  // space
	keyDEL KeyType = 127 // delete. on most systems this is mapped to backspace, I hear
)

// Control key aliases.
const (
	KeyNull      KeyType = keyNUL
	KeyBreak     KeyType = keyETX
	KeyEnter     KeyType = keyCR
	KeyBackspace KeyType = keyDEL
	KeyTab       KeyType = keyHT
	KeySpace     KeyType = keySP
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
)

// Mapping for control keys to friendly consts.
var keyNames = map[KeyType]string{
	keyNUL: "ctrl+@", // also ctrl+`
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
	keySP:  "space",
	keyDEL: "backspace",

	KeyRunes:    "runes",
	KeyUp:       "up",
	KeyDown:     "down",
	KeyRight:    "right",
	KeyLeft:     "left",
	KeyShiftTab: "shift+tab",
	KeyHome:     "home",
	KeyEnd:      "end",
	KeyPgUp:     "pgup",
	KeyPgDown:   "pgdown",
}

// Mapping for sequences to consts.
var sequences = map[string]KeyType{
	"\x1b[A": KeyUp,
	"\x1b[B": KeyDown,
	"\x1b[C": KeyRight,
	"\x1b[D": KeyLeft,
}

// Mapping for hex codes to consts. Unclear why these won't register as
// sequences.
var hexes = map[string]Key{
	"1b5b5a":       {Type: KeyShiftTab},
	"1b5b337e":     {Type: KeyDelete},
	"1b0d":         {Type: KeyEnter, Alt: true},
	"1b7f":         {Type: KeyBackspace, Alt: true},
	"1b5b48":       {Type: KeyHome},
	"1b5b377e":     {Type: KeyHome}, // urxvt
	"1b5b313b3348": {Type: KeyHome, Alt: true},
	"1b1b5b377e":   {Type: KeyHome, Alt: true}, // urxvt
	"1b5b46":       {Type: KeyEnd},
	"1b5b387e":     {Type: KeyEnd}, // urxvt
	"1b5b313b3346": {Type: KeyEnd, Alt: true},
	"1b1b5b387e":   {Type: KeyEnd, Alt: true}, // urxvt
	"1b5b357e":     {Type: KeyPgUp},
	"1b5b353b337e": {Type: KeyPgUp, Alt: true},
	"1b1b5b357e":   {Type: KeyPgUp, Alt: true}, // urxvt
	"1b5b367e":     {Type: KeyPgDown},
	"1b5b363b337e": {Type: KeyPgDown, Alt: true},
	"1b1b5b367e":   {Type: KeyPgDown, Alt: true}, // urxvt
	"1b5b313b3341": {Type: KeyUp, Alt: true},
	"1b5b313b3342": {Type: KeyDown, Alt: true},
	"1b5b313b3343": {Type: KeyRight, Alt: true},
	"1b5b313b3344": {Type: KeyLeft, Alt: true},

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

	// See if it's a mouse event. For now we're parsing X10-type mouse events
	// only.
	mouseEvent, err := parseX10MouseEvents(buf[:numBytes])
	if err == nil {
		var m []Msg
		for _, v := range mouseEvent {
			m = append(m, MouseMsg(v))
		}
		return m, nil
	}

	// Is it a special sequence, like an arrow key?
	if k, ok := sequences[string(buf[:numBytes])]; ok {
		return []Msg{
			KeyMsg(Key{Type: k}),
		}, nil
	}

	// Some of these need special handling
	hex := fmt.Sprintf("%x", buf[:numBytes])
	if k, ok := hexes[hex]; ok {
		return []Msg{
			KeyMsg(k),
		}, nil
	}

	// Is the alt key pressed? The buffer will be prefixed with an escape
	// sequence if so.
	if numBytes > 1 && buf[0] == 0x1b {
		// Now remove the initial escape sequence and re-process to get the
		// character being pressed in combination with alt.
		c, _ := utf8.DecodeRune(buf[1:])
		if c == utf8.RuneError {
			return nil, errors.New("could not decode rune after removing initial escape")
		}
		return []Msg{
			KeyMsg(Key{Alt: true, Type: KeyRunes, Runes: []rune{c}}),
		}, nil
	}

	var runes []rune
	b := buf[:numBytes]

	// Translate input into runes. In most cases we'll receive exactly one
	// rune, but there are cases, particularly when an input method editor is
	// used, where we can receive multiple runes at once.
	for i, w := 0, 0; i < len(b); i += w {
		r, width := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError {
			return nil, errors.New("could not decode rune")
		}
		runes = append(runes, r)
		w = width
	}

	if len(runes) == 0 {
		return nil, errors.New("received 0 runes from input")
	} else if len(runes) > 1 {
		// We received multiple runes, so we know this isn't a control
		// character, sequence, and so on.
		return []Msg{
			KeyMsg(Key{Type: KeyRunes, Runes: runes}),
		}, nil
	}

	// Is the first rune a control character?
	r := KeyType(runes[0])
	if numBytes == 1 && r <= keyUS || r == keyDEL {
		return []Msg{
			KeyMsg(Key{Type: r}),
		}, nil
	}

	// Welp, it's just a regular, ol' single rune
	return []Msg{
		KeyMsg(Key{Type: KeyRunes, Runes: runes}),
	}, nil
}
