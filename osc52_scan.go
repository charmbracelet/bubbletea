package tea

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
)

// Control bytes used by OSC52 clipboard sequences. A sequence is introduced by
// ESC ']' (or the 8-bit equivalent 0x9d) and terminated by BEL, ST (ESC '\'),
// or the 8-bit ST.
const (
	oscByte  = 0x1b
	osc8Byte = 0x9d
	stByte   = 0x9c
	belByte  = 0x07
	canByte  = 0x18
	subByte  = 0x1a
)

// osc52Prefix is the OSC command number for clipboard operations.
const osc52Prefix = "52;"

// osc52MaxPending bounds how many bytes are held back while waiting for an
// OSC52 sequence to complete. A sequence longer than this is passed through
// untouched rather than being intercepted.
const osc52MaxPending = 1 << 20

// osc52Parse is the outcome of parsing an OSC52 sequence.
type osc52Parse int

const (
	// osc52NeedMore means the sequence is incomplete and more bytes are
	// required before it can be parsed.
	osc52NeedMore osc52Parse = iota

	// osc52Pass means the sequence is complete but is not a usable OSC52
	// sequence, and should be forwarded untouched.
	osc52Pass

	// osc52Handle means the sequence is a well-formed OSC52 sequence that
	// should be handled by the bridge.
	osc52Handle
)

// osc52Bridge intercepts OSC52 clipboard sequences in a terminal output stream
// and performs the corresponding clipboard operations through a
// [ClipboardBackend].
//
// This is what allows programs running on terminals without OSC52 support,
// such as Apple's Terminal.app, to copy and paste. Set requests are consumed,
// read requests are answered by injecting the response into the program's
// terminal input via reply. All other bytes are forwarded to w untouched.
type osc52Bridge struct {
	backend ClipboardBackend
	w       io.Writer
	reply   io.Writer

	// pending holds bytes that may be the start of an OSC52 sequence which
	// has not been fully written yet.
	pending []byte
}

// newOSC52Bridge returns a bridge that performs clipboard operations through
// backend, forwards all other output to w, and injects clipboard replies into
// reply.
func newOSC52Bridge(backend ClipboardBackend, w, reply io.Writer) *osc52Bridge {
	return &osc52Bridge{backend: backend, w: w, reply: reply}
}

// Write filters OSC52 sequences out of p. Everything else is forwarded to the
// underlying writer.
func (b *osc52Bridge) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	buf := p
	if len(b.pending) > 0 {
		buf = append(b.pending, p...)
		b.pending = nil
	}

	out := make([]byte, 0, len(buf))
	for len(buf) > 0 {
		start, found := indexOSC52(buf)
		if !found {
			// Nothing to intercept. Forward everything except a trailing
			// partial introducer, which is held back until more data
			// arrives so a sequence split across writes can still be
			// intercepted.
			n := len(buf) - osc52PrefixLen(buf)
			out = append(out, buf[:n]...)
			b.pending = append(b.pending, buf[n:]...)
			break
		}

		out = append(out, buf[:start]...)
		rest := buf[start:]

		seqLen, selection, payload, terminator, status := parseOSC52(rest)
		switch status {
		case osc52NeedMore:
			if len(rest) > osc52MaxPending {
				// Never hold back unbounded data. Forward the sequence
				// untouched instead of intercepting it.
				out = append(out, rest...)
			} else {
				b.pending = append(b.pending, rest...)
			}
			rest = nil
		case osc52Pass:
			out = append(out, rest[:seqLen]...)
		case osc52Handle:
			if err := b.handle(selection, payload, terminator); err != nil {
				// Fail open: forward the sequence untouched so terminals
				// that do understand OSC52 can still act on it.
				out = append(out, rest[:seqLen]...)
			}
		}

		if rest == nil {
			break
		}
		buf = rest[seqLen:]
	}

	if len(out) == 0 {
		return len(p), nil
	}
	if _, err := b.w.Write(out); err != nil {
		return 0, fmt.Errorf("bubbletea: writing bridged output: %w", err)
	}
	return len(p), nil
}

// handle performs the clipboard operation described by a parsed OSC52
// sequence.
func (b *osc52Bridge) handle(selection byte, payload []byte, terminator byte) error {
	if bytes.Equal(payload, []byte("?")) {
		content, err := b.backend.Get(selection)
		if err != nil {
			return fmt.Errorf("bubbletea: reading clipboard: %w", err)
		}
		if _, err := b.reply.Write(encodeOSC52(selection, content, terminator)); err != nil {
			return fmt.Errorf("bubbletea: writing clipboard reply: %w", err)
		}
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(string(payload))
	if err != nil {
		return fmt.Errorf("bubbletea: invalid OSC52 payload: %w", err)
	}
	if err := b.backend.Set(selection, string(decoded)); err != nil {
		return fmt.Errorf("bubbletea: writing clipboard: %w", err)
	}
	return nil
}

// indexOSC52 returns the index of the first OSC52 introducer in b.
func indexOSC52(b []byte) (int, bool) {
	for i := 0; i < len(b); i++ {
		switch b[i] {
		case osc8Byte:
			if bytes.HasPrefix(b[i+1:], []byte(osc52Prefix)) {
				return i, true
			}
		case oscByte:
			if i+1 < len(b) && b[i+1] == ']' && bytes.HasPrefix(b[i+2:], []byte(osc52Prefix)) {
				return i, true
			}
		}
	}
	return 0, false
}

// osc52PrefixLen returns the length of the longest suffix of b that could be
// the beginning of an OSC52 introducer. Those bytes must not be forwarded
// before it is known whether the sequence should be intercepted.
//
// A trailing ESC is held back as well, since a sequence split between writes
// may start with it. Use [osc52Bridge.Flush] to emit anything that is still
// held back once the stream ends.
func osc52PrefixLen(b []byte) int {
	longest := 0
	for _, prefix := range []string{"\x1b", "\x1b]", "\x1b]5", "\x1b]52", "\x9d", "\x9d5", "\x9d52"} {
		if len(prefix) > longest && bytes.HasSuffix(b, []byte(prefix)) {
			longest = len(prefix)
		}
	}
	return longest
}

// Flush writes any bytes held back while waiting for a sequence to complete.
// It should be called once no more output is expected.
func (b *osc52Bridge) Flush() error {
	if len(b.pending) == 0 {
		return nil
	}
	pending := b.pending
	b.pending = nil
	if _, err := b.w.Write(pending); err != nil {
		return fmt.Errorf("bubbletea: writing bridged output: %w", err)
	}
	return nil
}

// parseOSC52 parses an OSC52 sequence at the start of b, which must begin with
// an OSC52 introducer. It returns the length of the sequence, the selection,
// the still-encoded payload, the terminator byte, and whether the sequence is
// complete.
func parseOSC52(b []byte) (n int, selection byte, payload []byte, terminator byte, status osc52Parse) {
	bodyStart := 1
	if b[0] == oscByte {
		bodyStart = 2
	}
	body := b[bodyStart:]

	end, termLen, found := findOSCTerminator(body)
	if !found {
		return 0, 0, nil, 0, osc52NeedMore
	}

	n = bodyStart + end + termLen
	terminator = body[end]
	if terminator == canByte || terminator == subByte {
		// The sequence was cancelled by its writer; leave it alone.
		return n, 0, nil, terminator, osc52Pass
	}
	if end < len(osc52Prefix) || !bytes.HasPrefix(body[:end], []byte(osc52Prefix)) {
		return n, 0, nil, terminator, osc52Pass
	}

	rest := body[len(osc52Prefix):end]
	semi := bytes.IndexByte(rest, ';')
	if semi < 1 {
		// Missing a selection or a payload separator.
		return n, 0, nil, terminator, osc52Pass
	}
	return n, rest[0], rest[semi+1:], terminator, osc52Handle
}

// findOSCTerminator scans body for the terminator of an OSC string. It returns
// the length of the string data preceding the terminator, the length of the
// terminator itself, and whether one was found.
func findOSCTerminator(body []byte) (end, termLen int, found bool) {
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case belByte, stByte, canByte, subByte:
			return i, 1, true
		case oscByte:
			if i+1 < len(body) && body[i+1] == '\\' {
				return i, 2, true
			}
		}
	}
	return 0, 0, false
}

// encodeOSC52 builds an OSC52 sequence carrying content for the given
// selection, terminated the same way the request was.
func encodeOSC52(selection byte, content string, terminator byte) []byte {
	seq := make([]byte, 0, len(content)+len(osc52Prefix)+8)
	seq = append(seq, oscByte, ']')
	seq = append(seq, osc52Prefix...)
	seq = append(seq, selection, ';')
	seq = base64.StdEncoding.AppendEncode(seq, []byte(content))
	switch terminator {
	case oscByte:
		seq = append(seq, oscByte, '\\')
	case stByte:
		seq = append(seq, stByte)
	default:
		seq = append(seq, belByte)
	}
	return seq
}
