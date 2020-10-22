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
//     // Switch on the type (safer)
//     switch msg := msg.(type) {
//     case KeyMsg:
//         switch msg.Type {
//         case KeyEnter:
//             fmt.Println("you pressed enter!")
//         case KeyRune:
//             switch msg.Rune {
//             case 'a':
//                 fmt.Println("you pressed a!")
//             }
//         }
//     }
//
//     // Switch on the string representation of the key (shorter)
//     switch msg := msg.(type) {
//     case KeyMsg:
//         switch msg.String() {
//         case "enter":
//             fmt.Println("you pressed enter!")
//         case "a':
//             fmt.Println("you pressed a!")
//         }
//     }
type KeyMsg Key

// String returns a friendly name for a key.
//
//     k := KeyType{Type: KeyEnter}
//     fmt.Println(k)
//     // Output: enter
func (k *KeyMsg) String() (str string) {
	if k.Alt {
		str += "alt+"
	}
	if k.Type == KeyRune {
		str += string(k.Rune)
		return str
	} else if s, ok := keyNames[int(k.Type)]; ok {
		str += s
		return str
	}
	return ""
}

// IsRune returns whether or not the key is a rune.
func (k *KeyMsg) IsRune() bool {
	return k.Type == KeyRune
}

// Key contains information about a keypress.
type Key struct {
	Type KeyType
	Rune rune
	Alt  bool
}

// KeyType indicates the key pressed, such as KeyEnter or KeyBreak or
// KeyCtrlC. All other keys will be type KeyRune. To get the rune value, check
// the Rune method on a Key struct, or use the Key.String() method:
//
//     k := Key{Type: KeyRune, Rune: 'a', Alt: true}
//     if k.Type == KeyRune {
//
//         fmt.Println(k.Rune)
//         // Output: a
//
//         fmt.Println(k.String())
//         // Output: alt+a
//
//     }
type KeyType int

// Control keys. I know we could do this with an iota, but the values are very
// specific, so we set the values explicitly to avoid any confusion.
//
// See also:
// https://en.wikipedia.org/wiki/C0_and_C1_control_codes
const (
	keyNUL = 0   // null, \0
	keySOH = 1   // start of heading
	keySTX = 2   // start of text
	keyETX = 3   // break, ctrl+c
	keyEOT = 4   // end of transmission
	keyENQ = 5   // enquiry
	keyACK = 6   // acknowledge
	keyBEL = 7   // bell, \a
	keyBS  = 8   // backspace
	keyHT  = 9   // horizontal tabulation, \t
	keyLF  = 10  // line feed, \n
	keyVT  = 11  // vertical tabulation \v
	keyFF  = 12  // form feed \f
	keyCR  = 13  // carriage return, \r
	keySO  = 14  // shift out
	keySI  = 15  // shift in
	keyDLE = 16  // data link escape
	keyDC1 = 17  // device control one
	keyDC2 = 18  // device control two
	keyDC3 = 19  // device control three
	keyDC4 = 20  // device control four
	keyNAK = 21  // negative acknowledge
	keySYN = 22  // synchronous idle
	keyETB = 23  // end of transmission block
	keyCAN = 24  // cancel
	keyEM  = 25  // end of medium
	keySUB = 26  // substitution
	keyESC = 27  // escape, \e
	keyFS  = 28  // file separator
	keyGS  = 29  // group separator
	keyRS  = 30  // record separator
	keyUS  = 31  // unit separator
	keySP  = 32  // space
	keyDEL = 127 // delete. on most systems this is mapped to backspace, I hear
)

// Control key aliases.
const (
	KeyNull      = keyNUL
	KeyBreak     = keyETX
	KeyEnter     = keyCR
	KeyBackspace = keyDEL
	KeyTab       = keyHT
	KeySpace     = keySP
	KeyEsc       = keyESC
	KeyEscape    = keyESC

	KeyCtrlAt           = keyNUL // ctrl+@
	KeyCtrlA            = keySOH
	KeyCtrlB            = keySTX
	KeyCtrlC            = keyETX
	KeyCtrlD            = keyEOT
	KeyCtrlE            = keyENQ
	KeyCtrlF            = keyACK
	KeyCtrlG            = keyBEL
	KeyCtrlH            = keyBS
	KeyCtrlI            = keyHT
	KeyCtrlJ            = keyLF
	KeyCtrlK            = keyVT
	KeyCtrlL            = keyFF
	KeyCtrlM            = keyCR
	KeyCtrlN            = keySO
	KeyCtrlO            = keySI
	KeyCtrlP            = keyDLE
	KeyCtrlQ            = keyDC1
	KeyCtrlR            = keyDC2
	KeyCtrlS            = keyDC3
	KeyCtrlT            = keyDC4
	KeyCtrlU            = keyNAK
	KeyCtrlV            = keySYN
	KeyCtrlW            = keyETB
	KeyCtrlX            = keyCAN
	KeyCtrlY            = keyEM
	KeyCtrlZ            = keySUB
	KeyCtrlOpenBracket  = keyESC // ctrl+[
	KeyCtrlBackslash    = keyFS  // ctrl+\
	KeyCtrlCloseBracket = keyGS  // ctrl+]
	KeyCtrlCaret        = keyRS  // ctrl+^
	KeyCtrlUnderscore   = keyUS  // ctrl+_
	KeyCtrlQuestionMark = keyDEL // ctrl+?
)

// Other keys.
const (
	KeyRune = -(iota + 1)
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
var keyNames = map[int]string{
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

	KeyRune:     "rune",
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

// readInput reads keypress and mouse input from a TTY and returns a message
// containing information about the key or mouse event accordingly.
func readInput(r io.Reader) (Msg, error) {
	var buf [256]byte

	// Read and block
	numBytes, err := r.Read(buf[:])
	if err != nil {
		return nil, err
	}

	// See if it's a mouse event. For now we're parsing X10-type mouse events
	// only.
	mouseEvent, err := parseX10MouseEvent(buf[:numBytes])
	if err == nil {
		return MouseMsg(mouseEvent), nil
	}

	hex := fmt.Sprintf("%x", buf[:numBytes])

	// Some of these need special handling
	if k, ok := hexes[hex]; ok {
		return KeyMsg(k), nil
	}

	// Get unicode value
	char, _ := utf8.DecodeRune(buf[:])
	if char == utf8.RuneError {
		return nil, errors.New("could not decode rune")
	}

	// Is it a control character?
	if numBytes == 1 && char <= keyUS || char == keyDEL {
		return KeyMsg(Key{Type: KeyType(char)}), nil
	}

	// Is it a special sequence, like an arrow key?
	if k, ok := sequences[string(buf[:numBytes])]; ok {
		return KeyMsg(Key{Type: k}), nil
	}

	// Is the alt key pressed? The buffer will be prefixed with an escape
	// sequence if so
	if numBytes > 1 && buf[0] == 0x1b {
		// Now remove the initial escape sequence and re-process to get the
		// character.
		c, _ := utf8.DecodeRune(buf[1:])
		if c == utf8.RuneError {
			return nil, errors.New("could not decode rune after removing initial escape")
		}
		return KeyMsg(Key{Alt: true, Type: KeyRune, Rune: c}), nil
	}

	// Just a regular, ol' rune
	return KeyMsg(Key{Type: KeyRune, Rune: char}), nil
}
