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
	} else if s, ok := keyNames[k.Type]; ok {
		return s
	}
	return ""
}

// IsRune returns weather or not the key is a rune
func (k *KeyMsg) IsRune() bool {
	if k.Type == KeyRune {
		return true
	}
	return false
}

type Key struct {
	Type KeyType
	Rune rune
}

// KeyType indicates the key pressed
type KeyType int

// Possible keys
const (
	KeyBreak KeyType = iota
	KeyEnter
	KeyEscape
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyUnitSeparator
	KeyBackspace
	KeyRune
)

// Friendly key names
var keyNames = map[KeyType]string{
	KeyBreak:         "break",
	KeyEnter:         "enter",
	KeyEscape:        "esc",
	KeyUp:            "up",
	KeyDown:          "down",
	KeyRight:         "right",
	KeyLeft:          "left",
	KeyUnitSeparator: "us",
	KeyBackspace:     "backspace",
	KeyRune:          "rune",
}

// Control keys. I know we could do this with an iota, but the values are very
// specific, so we set the values explicitly to avoid any confusion
const (
	keyETX = 3   // break, ctrl+c
	keyLF  = 9   // line-feed, \n
	keyCR  = 13  // carriage return, \r
	keyESC = 27  // escape
	keyUS  = 31  // unit separator
	keyDEL = 127 // delete. on most systems this is mapped to backspace, I hear
)

// Mapping for control keys to friendly consts
var controlKeys = map[int]KeyType{
	keyETX: KeyBreak,
	keyLF:  KeyEnter,
	keyCR:  KeyEnter,
	keyESC: KeyEscape,
	keyUS:  KeyUnitSeparator,
	keyDEL: KeyBackspace,
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
	if n == 1 && c <= keyUS || c == keyDEL {
		if k, ok := controlKeys[n]; ok {
			return Key{Type: k}, nil
		}
	}

	if n == 1 && c <= keyUS {
		if k, ok := controlKeys[int(c)]; ok {
			return Key{Type: k}, nil
		}
	}

	// Is it a special sequence, like an arrow key?
	if k, ok := sequences[string(buf[:n])]; ok {
		return Key{Type: k}, nil
	}

	// Nope, just a regular, ol' rune
	return Key{Type: KeyRune, Rune: c}, nil
}
