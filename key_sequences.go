package tea

import "sort"

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
