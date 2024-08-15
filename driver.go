package tea

import (
	"bytes"
	"io"
	"unicode/utf8"

	"github.com/muesli/cancelreader"
)

// driver represents an ANSI terminal input driver.
// It reads input events and parses ANSI sequences from the terminal input
// buffer.
type driver struct {
	rd    cancelreader.CancelReader
	table map[string]Key // table is a lookup table for key sequences.

	term string // term is the terminal name $TERM.

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	buf [256]byte // do we need a larger buffer?

	// prevMouseState keeps track of the previous mouse state to determine mouse
	// up button events.
	prevMouseState uint32 // nolint: unused

	// lastWinsizeEvent keeps track of the last window size event to prevent
	// multiple size events from firing.
	lastWinsizeEventX, lastWinsizeEventY int16 // nolint: unused

	flags int // control the behavior of the driver.
}

// newDriver returns a new ANSI input driver.
// This driver uses ANSI control codes compatible with VT100/VT200 terminals,
// and XTerm. It supports reading Terminfo databases to overwrite the default
// key sequences.
func newDriver(r io.Reader, term string, flags int) (*driver, error) {
	d := new(driver)
	cr, err := newCancelreader(r)
	if err != nil {
		return nil, err
	}

	d.rd = cr
	d.table = buildKeysTable(flags, term)
	d.term = term
	d.flags = flags
	return d, nil
}

// Cancel cancels the underlying reader.
func (d *driver) Cancel() bool {
	return d.rd.Cancel()
}

// Close closes the underlying reader.
func (d *driver) Close() error {
	return d.rd.Close()
}

func (d *driver) readEvents() (msgs []Msg, err error) {
	nb, err := d.rd.Read(d.buf[:])
	if err != nil {
		return nil, err
	}

	buf := d.buf[:nb]

	// Lookup table first
	if bytes.HasPrefix(buf, []byte{'\x1b'}) {
		if k, ok := d.table[string(buf)]; ok {
			msgs = append(msgs, KeyPressMsg(k))
			return
		}
	}

	var i int
	for i < len(buf) {
		nb, ev := parseSequence(buf[i:])

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := ev.(PasteEndMsg); !ok {
				d.paste = append(d.paste, buf[i])
				i++
				continue
			}
		}

		switch ev.(type) {
		case UnknownMsg:
			// If the sequence is not recognized by the parser, try looking it up.
			if k, ok := d.table[string(buf[i:i+nb])]; ok {
				ev = KeyPressMsg(k)
			}
		case PasteStartMsg:
			d.paste = []byte{}
		case PasteEndMsg:
			// Decode the captured data into runes.
			var paste []rune
			for len(d.paste) > 0 {
				r, w := utf8.DecodeRune(d.paste)
				if r != utf8.RuneError {
					paste = append(paste, r)
				}
				d.paste = d.paste[w:]
			}
			d.paste = nil // reset the buffer
			msgs = append(msgs, PasteMsg(paste))
		case nil:
			i++
			continue
		}

		if mevs, ok := ev.(multiMsg); ok {
			msgs = append(msgs, []Msg(mevs)...)
		} else {
			msgs = append(msgs, ev)
		}
		i += nb
	}

	return
}
