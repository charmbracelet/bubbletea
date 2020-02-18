package tea

import (
	"errors"
	"io"
	"unicode/utf8"
)

// KeyPressMsg contains information about a keypress
type KeyMsg Key

// String returns a friendly name for a key
func (k *KeyMsg) String() string {
	if k.Type == KeyRune {
		return string(k.Rune)
	} else if s, ok := keyNames[int(k.Type)]; ok {
		return s
	}
	return ""
}

// IsRune returns weather or not the key is a rune
func (k *KeyMsg) IsRune() bool {
	return k.Type == KeyRune
}

// Key contains information about a keypress
type Key struct {
	Type KeyType
	Rune rune
}

// KeyType indicates the key pressed
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

// Aliases
const (
	KeyBreak     = keyETX
	KeyEnter     = keyCR
	KeyBackspace = keyBS
	KeySpace     = keySP
	KeyEsc       = keyESC
	KeyEscape    = keyESC
	KeyDelete    = keyDEL

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
	KeyCtrlV            = keyETB
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

const (
	KeyRune = -(iota + 1)
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
)

// Mapping for control keys to friendly consts
var keyNames = map[int]string{
	keyNUL: "ctrl+@", // also ctrl+`
	keySOH: "ctrl+a",
	keySTX: "ctrl+b",
	keyETX: "ctrl+c",
	keyEOT: "ctrl+d",
	keyENQ: "ctrl+e",
	keyACK: "ctrl+f",
	keyBEL: "ctrl+g",
	keyBS:  "backspace", // also ctrl+h
	keyHT:  "tab",       // also ctrl+i
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
	keyDEL: "delete",

	KeyRune:  "rune",
	KeyUp:    "up",
	KeyDown:  "down",
	KeyRight: "right",
	KeyLeft:  "left",
}

// Mapping for sequences to consts
var sequences = map[string]KeyType{
	"\x1b[A": KeyUp,
	"\x1b[B": KeyDown,
	"\x1b[C": KeyRight,
	"\x1b[D": KeyLeft,
}

// ReadKey reads keypress input from a TTY and returns a string representation
// of a key
func ReadKey(r io.Reader) (Key, error) {
	var buf [256]byte

	// Read and block
	n, err := r.Read(buf[:])
	if err != nil {
		return Key{}, err
	}

	// Get rune
	c, _ := utf8.DecodeRune(buf[:])
	if c == utf8.RuneError {
		return Key{}, errors.New("no such rune")
	}

	// Is it a control character?
	if /*n == 1 &&*/ c <= keyUS || c == keyDEL {
		return Key{Type: KeyType(c)}, nil
	}

	// Is it a special sequence, like an arrow key?
	if k, ok := sequences[string(buf[:n])]; ok {
		return Key{Type: k}, nil
	}

	// Nope, just a regular, ol' rune
	return Key{Type: KeyRune, Rune: c}, nil
}
