package tea

import (
	"errors"
	"io"
	"unicode/utf8"
)

// KeyPressMsg contains information about a keypress
type KeyPressMsg string

// ReadKey reads keypress input from a TTY and returns a string representation
// of a key
func ReadKey(r io.Reader) (string, error) {
	var buf [256]byte

	// Read and block
	_, err := r.Read(buf[:])
	if err != nil {
		return "", err
	}

	// TODO: non-rune keys like arrows, meta keys, and so on

	// Read "normal" key
	c, _ := utf8.DecodeRune(buf[:])
	if c == utf8.RuneError {
		return "", errors.New("no such rune")
	}

	return string(c), nil
}
