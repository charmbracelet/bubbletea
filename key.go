package tea

import (
	"errors"
	"io"
	"unicode/utf8"
)

// KeyPressMsg contains information about a keypress
type KeyPressMsg string

// Control keys
const (
	keyETX = 3  // break, ctrl+c
	keyLF  = 9  // line-feed, \n
	keyCR  = 13 // carriage return, \r
	keyESC = 27
	keyUS  = 31
)

var controlNames = map[int]string{
	keyETX: "ctrl+c",
	keyLF:  "enter",
	keyCR:  "enter",
	keyESC: "esc",
	keyUS:  "us",
}

var sequenceNames = map[string]string{
	"\x1b[A": "up",
	"\x1b[B": "down",
	"\x1b[C": "right",
	"\x1b[D": "left",
}

// ReadKey reads keypress input from a TTY and returns a string representation
// of a key
func ReadKey(r io.Reader) (string, error) {
	var buf [256]byte

	// Read and block
	n, err := r.Read(buf[:])
	if err != nil {
		return "", err
	}

	// Get rune
	c, _ := utf8.DecodeRune(buf[:])
	if c == utf8.RuneError {
		return "", errors.New("no such rune")
	}

	// Is it a control character?
	if n == 1 && c <= keyUS {
		if s, ok := controlNames[int(c)]; ok {
			return s, nil
		}
	}

	// Is it a special sequence, like an arrow key?
	if s, ok := sequenceNames[string(buf[:n])]; ok {
		return s, nil
	}

	// Nope, just a regular, ol' rune
	return string(c), nil
}
