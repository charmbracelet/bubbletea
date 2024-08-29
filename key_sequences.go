package tea

import (
	"bytes"
	"sort"
	"unicode/utf8"
)

// extSequences is used by the map-based algorithm below. It contains
// the sequences plus their alternatives with an escape character
// prefixed, plus the control chars, plus the space.
// It does not contain the NUL character, which is handled specially
// by detectOneMsg.
var extSequences = func() map[string]Key {
	s := map[string]Key{}
	for seq, key := range sequences {
		key := key
		s[seq] = key
		if !key.Alt {
			key.Alt = true
			s["\x1b"+seq] = key
		}
	}
	for i := keyNUL + 1; i <= keyDEL; i++ {
		if i == keyESC {
			continue
		}
		s[string([]byte{byte(i)})] = Key{Type: i}
		s[string([]byte{'\x1b', byte(i)})] = Key{Type: i, Alt: true}
		if i == keyUS {
			i = keyDEL - 1
		}
	}
	s[" "] = Key{Type: KeySpace, Runes: spaceRunes}
	s["\x1b "] = Key{Type: KeySpace, Alt: true, Runes: spaceRunes}
	s["\x1b\x1b"] = Key{Type: KeyEscape, Alt: true}
	return s
}()

// seqLengths is the sizes of valid sequences, starting with the
// largest size.
var seqLengths = func() []int {
	sizes := map[int]struct{}{}
	for seq := range extSequences {
		sizes[len(seq)] = struct{}{}
	}
	lsizes := make([]int, 0, len(sizes))
	for sz := range sizes {
		lsizes = append(lsizes, sz)
	}
	sort.Slice(lsizes, func(i, j int) bool { return lsizes[i] > lsizes[j] })
	return lsizes
}()

// detectSequence uses a longest prefix match over the input
// sequence and a hash map.
func detectSequence(input []byte) (hasSeq bool, width int, msg Msg) {
	seqs := extSequences
	for _, sz := range seqLengths {
		if sz > len(input) {
			continue
		}
		prefix := input[:sz]
		key, ok := seqs[string(prefix)]
		if ok {
			return true, sz, KeyMsg(key)
		}
	}
	// Is this an unknown CSI sequence?
	if loc := unknownCSIRe.FindIndex(input); loc != nil {
		return true, loc[1], unknownCSISequenceMsg(input[:loc[1]])
	}

	return false, 0, nil
}

// detectBracketedPaste detects an input pasted while bracketed
// paste mode was enabled.
//
// Note: this function is a no-op if bracketed paste was not enabled
// on the terminal, since in that case we'd never see this
// particular escape sequence.
func detectBracketedPaste(input []byte) (hasBp bool, width int, msg Msg) {
	// Detect the start sequence.
	const bpStart = "\x1b[200~"
	if len(input) < len(bpStart) || string(input[:len(bpStart)]) != bpStart {
		return false, 0, nil
	}

	// Skip over the start sequence.
	input = input[len(bpStart):]

	// If we saw the start sequence, then we must have an end sequence
	// as well. Find it.
	const bpEnd = "\x1b[201~"
	idx := bytes.Index(input, []byte(bpEnd))
	inputLen := len(bpStart) + idx + len(bpEnd)
	if idx == -1 {
		// We have encountered the end of the input buffer without seeing
		// the marker for the end of the bracketed paste.
		// Tell the outer loop we have done a short read and we want more.
		return true, 0, nil
	}

	// The paste is everything in-between.
	paste := input[:idx]

	// All there is in-between is runes, not to be interpreted further.
	k := Key{Type: KeyRunes, Paste: true}
	for len(paste) > 0 {
		r, w := utf8.DecodeRune(paste)
		if r != utf8.RuneError {
			k.Runes = append(k.Runes, r)
		}
		paste = paste[w:]
	}

	return true, inputLen, KeyMsg(k)
}

// detectReportFocus detects a focus report sequence.
// nolint: gomnd
func detectReportFocus(input []byte) (hasRF bool, width int, msg Msg) {
	switch {
	case bytes.Equal(input, []byte("\x1b[I")):
		return true, 3, FocusMsg{}
	case bytes.Equal(input, []byte("\x1b[O")):
		return true, 3, BlurMsg{}
	}
	return false, 0, nil
}
