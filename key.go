package tea

import (
	"errors"
	"io"
	"unicode/utf8"
)

// KeyPressMsg contains information about a keypress
type KeyPressMsg string

var keyNames = map[string]string{
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

	// Was it a special key, like an arrow key?
	if s, ok := keyNames[string(buf[:n])]; ok {
		return s, nil
	}

	// Nope, just a regular key
	c, _ := utf8.DecodeRune(buf[:])
	if c == utf8.RuneError {
		return "", errors.New("no such rune")
	}

	return string(c), nil
}
