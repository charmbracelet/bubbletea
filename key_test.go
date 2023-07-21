package tea

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestKeyString(t *testing.T) {
	for name, test := range map[string]struct {
		key      Key
		expected string
	}{
		"alt+space": {key: Key{Type: KeySpace, Alt: true}, expected: "alt+ "},
		"runes":     {key: Key{Type: KeyRunes, Runes: []rune{'a'}}, expected: "a"},
		"invalid":   {key: Key{Type: KeyType(99999)}, expected: ""},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, MsgKey(test.key).String())
		})
	}
}

func TestKeyTypeString(t *testing.T) {
	for name, test := range map[string]struct {
		keyType  KeyType
		expected string
	}{
		"space": {
			keyType:  KeySpace,
			expected: " ",
		},
		"invalid": {
			keyType:  KeyType(99999),
			expected: "",
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.keyType.String())
		})
	}
}

type seqTest struct {
	seq []byte
	msg Msg
}

// buildBaseSeqTests returns sequence tests that are valid for the
// detectSequence() function.
func buildBaseSeqTests() []seqTest {
	tests := []seqTest{}
	for seq, key := range sequences {
		key := key
		tests = append(tests, seqTest{[]byte(seq), MsgKey(key)})
		if !key.Alt {
			key.Alt = true
			tests = append(tests, seqTest{[]byte("\x1b" + seq), MsgKey(key)})
		}
	}
	// Add all the control characters.
	for i := keyNUL + 1; i <= keyDEL; i++ {
		if i == keyESC {
			// Not handled in detectSequence(), so not part of the base test
			// suite.
			continue
		}
		tests = append(tests,
			seqTest{[]byte{byte(i)}, MsgKey{Type: i}},
			seqTest{[]byte{'\x1b', byte(i)}, MsgKey{Type: i, Alt: true}},
		)
		if i == keyUS {
			i = keyDEL - 1
		}
	}

	// Additional special cases.
	tests = append(tests,
		// Unrecognized CSI sequence.
		seqTest{
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			unknownCSISequenceMsg([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'}),
		},
		// A lone space character.
		seqTest{
			[]byte{' '},
			MsgKey{Type: KeySpace, Runes: []rune(" ")},
		},
		// An escape character with the alt modifier.
		seqTest{
			[]byte{'\x1b', ' '},
			MsgKey{Type: KeySpace, Runes: []rune(" "), Alt: true},
		},
	)
	return tests
}

func TestDetectSequence(t *testing.T) {
	for _, test := range buildBaseSeqTests() {
		t.Run(fmt.Sprintf("%q", string(test.seq)), func(t *testing.T) {
			hasSeq, width, msg := detectSequence(test.seq)
			assert.True(t, hasSeq, "no sequence found")
			assert.Len(t, test.seq, width, "parser did not consume the entire input")
			assert.Equal(t, test.msg, msg)
		})
	}
}

func TestDetectOneMsg(t *testing.T) {
	tests := buildBaseSeqTests()
	// Add tests for the inputs that detectOneMsg() can parse, but
	// detectSequence() cannot.
	tests = append(tests,
		// Mouse event.
		seqTest{
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			MouseMsg{X: 32, Y: 16, Type: MouseWheelUp},
		},
		// Runes.
		seqTest{[]byte{'a'}, MsgKey{Type: KeyRunes, Runes: []rune("a")}},
		seqTest{[]byte{'\x1b', 'a'}, MsgKey{Type: KeyRunes, Runes: []rune("a"), Alt: true}},
		seqTest{[]byte{'a', 'a', 'a'}, MsgKey{Type: KeyRunes, Runes: []rune("aaa")}},
		// Multi-byte rune.
		seqTest{[]byte("☃"), MsgKey{Type: KeyRunes, Runes: []rune("☃")}},
		seqTest{[]byte("\x1b☃"), MsgKey{Type: KeyRunes, Runes: []rune("☃"), Alt: true}},
		// Standalone control chacters.
		seqTest{[]byte{'\x1b'}, MsgKey{Type: KeyEscape}},
		seqTest{[]byte{byte(keySOH)}, MsgKey{Type: KeyCtrlA}},
		seqTest{[]byte{'\x1b', byte(keySOH)}, MsgKey{Type: KeyCtrlA, Alt: true}},
		seqTest{[]byte{byte(keyNUL)}, MsgKey{Type: KeyCtrlAt}},
		seqTest{[]byte{'\x1b', byte(keyNUL)}, MsgKey{Type: KeyCtrlAt, Alt: true}},
		// Invalid characters.
		seqTest{[]byte{'\x80'}, unknownInputByteMsg(0x80)},
	)

	if runtime.GOOS != "windows" {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		tests = append(tests, seqTest{
			[]byte{'\xfe'},
			unknownInputByteMsg(0xfe),
		})
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%q", string(test.seq)), func(t *testing.T) {
			width, msg := detectOneMsg(test.seq)
			assert.Len(t, test.seq, width)
			assert.Equal(t, test.msg, msg)
		})
	}
}

func TestReadInput(t *testing.T) {
	type test struct {
		keyname string
		in      []byte
		out     []Msg
	}
	tests := []test{
		{
			"a",
			[]byte{'a'},
			[]Msg{
				MsgKey{
					Type:  KeyRunes,
					Runes: []rune{'a'},
				},
			},
		},
		{
			" ",
			[]byte{' '},
			[]Msg{
				MsgKey{
					Type:  KeySpace,
					Runes: []rune{' '},
				},
			},
		},
		{
			"a alt+a",
			[]byte{'a', '\x1b', 'a'},
			[]Msg{
				MsgKey{Type: KeyRunes, Runes: []rune{'a'}},
				MsgKey{Type: KeyRunes, Runes: []rune{'a'}, Alt: true},
			},
		},
		{
			"a alt+a a",
			[]byte{'a', '\x1b', 'a', 'a'},
			[]Msg{
				MsgKey{Type: KeyRunes, Runes: []rune{'a'}},
				MsgKey{Type: KeyRunes, Runes: []rune{'a'}, Alt: true},
				MsgKey{Type: KeyRunes, Runes: []rune{'a'}},
			},
		},
		{
			"ctrl+a",
			[]byte{byte(keySOH)},
			[]Msg{
				MsgKey{
					Type: KeyCtrlA,
				},
			},
		},
		{
			"ctrl+a ctrl+b",
			[]byte{byte(keySOH), byte(keySTX)},
			[]Msg{
				MsgKey{Type: KeyCtrlA},
				MsgKey{Type: KeyCtrlB},
			},
		},
		{
			"alt+a",
			[]byte{byte(0x1b), 'a'},
			[]Msg{
				MsgKey{
					Type:  KeyRunes,
					Alt:   true,
					Runes: []rune{'a'},
				},
			},
		},
		{
			"abcd",
			[]byte{'a', 'b', 'c', 'd'},
			[]Msg{
				MsgKey{
					Type:  KeyRunes,
					Runes: []rune{'a', 'b', 'c', 'd'},
				},
			},
		},
		{
			"up",
			[]byte("\x1b[A"),
			[]Msg{
				MsgKey{
					Type: KeyUp,
				},
			},
		},
		{
			"wheel up",
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Msg{
				MouseMsg{
					X:    32,
					Y:    16,
					Type: MouseWheelUp,
				},
			},
		},
		{
			"left release",
			[]byte{
				'\x1b', '[', 'M', byte(32) + 0b0010_0000, byte(32 + 33), byte(16 + 33),
				'\x1b', '[', 'M', byte(32) + 0b0000_0011, byte(64 + 33), byte(32 + 33),
			},
			[]Msg{
				MouseMsg(MouseEvent{
					X:    32,
					Y:    16,
					Type: MouseLeft,
				}),
				MouseMsg(MouseEvent{
					X:    64,
					Y:    32,
					Type: MouseRelease,
				}),
			},
		},
		{
			"shift+tab",
			[]byte{'\x1b', '[', 'Z'},
			[]Msg{
				MsgKey{
					Type: KeyShiftTab,
				},
			},
		},
		{
			"enter",
			[]byte{'\r'},
			[]Msg{MsgKey{Type: KeyEnter}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\r'},
			[]Msg{
				MsgKey{
					Type: KeyEnter,
					Alt:  true,
				},
			},
		},
		{
			"insert",
			[]byte{'\x1b', '[', '2', '~'},
			[]Msg{
				MsgKey{
					Type: KeyInsert,
				},
			},
		},
		{
			"alt+ctrl+a",
			[]byte{'\x1b', byte(keySOH)},
			[]Msg{
				MsgKey{
					Type: KeyCtrlA,
					Alt:  true,
				},
			},
		},
		{
			"?CSI[45 45 45 45 88]?",
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Msg{unknownCSISequenceMsg([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'})},
		},
		// Powershell sequences.
		{
			"up",
			[]byte{'\x1b', 'O', 'A'},
			[]Msg{MsgKey{Type: KeyUp}},
		},
		{
			"down",
			[]byte{'\x1b', 'O', 'B'},
			[]Msg{MsgKey{Type: KeyDown}},
		},
		{
			"right",
			[]byte{'\x1b', 'O', 'C'},
			[]Msg{MsgKey{Type: KeyRight}},
		},
		{
			"left",
			[]byte{'\x1b', 'O', 'D'},
			[]Msg{MsgKey{Type: KeyLeft}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\x0d'},
			[]Msg{MsgKey{Type: KeyEnter, Alt: true}},
		},
		{
			"alt+backspace",
			[]byte{'\x1b', '\x7f'},
			[]Msg{MsgKey{Type: KeyBackspace, Alt: true}},
		},
		{
			"ctrl+@",
			[]byte{'\x00'},
			[]Msg{MsgKey{Type: KeyCtrlAt}},
		},
		{
			"alt+ctrl+@",
			[]byte{'\x1b', '\x00'},
			[]Msg{MsgKey{Type: KeyCtrlAt, Alt: true}},
		},
		{
			"esc",
			[]byte{'\x1b'},
			[]Msg{MsgKey{Type: KeyEsc}},
		},
		{
			"alt+esc",
			[]byte{'\x1b', '\x1b'},
			[]Msg{MsgKey{Type: KeyEsc, Alt: true}},
		},
		// Bracketed paste does not work yet.
		{
			"?CSI[50 48 48 126]? a   b ?CSI[50 48 49 126]?",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', ' ', 'b',
				'\x1b', '[', '2', '0', '1', '~',
			},
			[]Msg{
				// What we expect once bracketed paste is recognized properly:
				//
				//  MsgKey{Type: KeyRunes, Runes: []rune("a b")},
				//
				// What we get instead (for now):
				unknownCSISequenceMsg{0x1b, 0x5b, 0x32, 0x30, 0x30, 0x7e},
				MsgKey{Type: KeyRunes, Runes: []rune{'a'}},
				MsgKey{Type: KeySpace, Runes: []rune{' '}},
				MsgKey{Type: KeyRunes, Runes: []rune{'b'}},
				unknownCSISequenceMsg{0x1b, 0x5b, 0x32, 0x30, 0x31, 0x7e},
			},
		},
	}
	if runtime.GOOS != "windows" {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		tests = append(tests,
			test{
				"?0xfe?",
				[]byte{'\xfe'},
				[]Msg{unknownInputByteMsg(0xfe)},
			},
			test{
				"a ?0xfe?   b",
				[]byte{'a', '\xfe', ' ', 'b'},
				[]Msg{
					MsgKey{Type: KeyRunes, Runes: []rune{'a'}},
					unknownInputByteMsg(0xfe),
					MsgKey{Type: KeySpace, Runes: []rune{' '}},
					MsgKey{Type: KeyRunes, Runes: []rune{'b'}},
				},
			},
		)
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, test.keyname), func(t *testing.T) {
			msgs := testReadInputs(t, bytes.NewReader(test.in))
			var buf strings.Builder
			for i, msg := range msgs {
				if i > 0 {
					buf.WriteByte(' ')
				}
				if s, ok := msg.(fmt.Stringer); ok {
					buf.WriteString(s.String())
				} else {
					fmt.Fprintf(&buf, "%#v:%T", msg, msg)
				}
			}

			assert.Equal(t, test.keyname, buf.String())
			assert.Equal(t, test.out, msgs)
		})
	}
}

func testReadInputs(t *testing.T, input io.Reader) []Msg {
	// We'll check that the input reader finishes at the end without error.
	var wg sync.WaitGroup
	var inputErr error
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
		if inputErr != nil && !errors.Is(inputErr, io.EOF) {
			t.Fatalf("unexpected input error: %v", inputErr)
		}
	}()

	// The messages we're consuming.
	msgsC := make(chan Msg)

	// Start the reader in the background.
	wg.Add(1)
	go func() {
		defer wg.Done()
		inputErr = readInputs(ctx, msgsC, input)
		msgsC <- nil
	}()

	var msgs []Msg
loop:
	for {
		select {
		case msg := <-msgsC:
			if msg == nil {
				// end of input marker for the test.
				break loop
			}
			msgs = append(msgs, msg)
		case <-time.After(2 * time.Second):
			t.Errorf("timeout waiting for input event")
			break loop
		}
	}
	return msgs
}

// randTest defines the test input and expected output for a sequence
// of interleaved control sequences and control characters.
type randTest struct {
	data    []byte
	lengths []int
	names   []string
}

// seed is the random seed to randomize the input. This helps check
// that all the sequences get ultimately exercised.
var seed = flag.Int64("seed", 0, "random seed (0 to autoselect)")

// genRandomData generates a randomized test, with a random seed unless
// the seed flag was set.
func genRandomData(logfn func(int64), length int) randTest {
	// We'll use a random source. However, we give the user the option
	// to override it to a specific value for reproduceability.
	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}
	// Inform the user so they know what to reuse to get the same data.
	logfn(s)
	return genRandomDataWithSeed(s, length)
}

// genRandomDataWithSeed generates a randomized test with a fixed seed.
func genRandomDataWithSeed(s int64, length int) randTest {
	r := rand.New(rand.NewSource(s)) //nolint:gosec

	// allseqs contains all the sequences, in sorted order. We sort
	// to make the test deterministic (when the seed is also fixed).
	type seqpair struct {
		seq  string
		name string
	}
	allseqs := lo.MapToSlice(sequences, func(seq string, key Key) seqpair {
		return seqpair{seq, key.String()}
	})
	sort.Slice(allseqs, func(i, j int) bool {
		return allseqs[i].seq < allseqs[j].seq
	})

	// res contains the computed test.
	var res randTest
	for len(res.data) < length {
		alt := r.Intn(2)
		prefix := ""
		esclen := 0
		if alt == 1 {
			prefix = "alt+"
			esclen = 1
		}
		kind := r.Intn(3)
		switch kind {
		case 0:
			// A control character.
			if alt == 1 {
				res.data = append(res.data, '\x1b')
			}
			res = randTest{
				data:    append(res.data, 1),
				names:   append(res.names, prefix+"ctrl+a"),
				lengths: append(res.lengths, 1+esclen),
			}
		case 1, 2:
			// A sequence.
			seqi := r.Intn(len(allseqs))
			s := allseqs[seqi]
			if strings.HasPrefix(s.name, "alt+") {
				esclen = 0
				prefix = ""
				alt = 0
			}
			if alt == 1 {
				res.data = append(res.data, '\x1b')
			}
			res = randTest{
				data:    append(res.data, s.seq...),
				names:   append(res.names, prefix+s.name),
				lengths: append(res.lengths, len(s.seq)+esclen),
			}
		}
	}
	return res
}

// TestDetectRandomSequencesLex checks that the lex-generated sequence
// detector works over concatenations of random sequences.
func TestDetectRandomSequencesLex(t *testing.T) {
	runTestDetectSequence(t, detectSequence)
}

func runTestDetectSequence(
	t *testing.T,
	detectSequence func(input []byte) (hasSeq bool, width int, msg Msg),
) {
	for wtf := 0; wtf < 10; wtf++ {
		t.Run(fmt.Sprintf("#%d", wtf), func(t *testing.T) {
			td := genRandomData(func(s int64) { t.Logf("using random seed: %d", s) }, 1000)

			t.Logf("%#v", td)

			// tn is the event number in td.
			// i is the cursor in the input data.
			// w is the length of the last sequence detected.
			for tn, i, w := 0, 0, 0; i < len(td.data); tn, i = tn+1, i+w {
				hasSequence, width, msg := detectSequence(td.data[i:])
				assert.True(t, hasSequence, "at %d (ev %d): failed to find sequence", i, tn)
				assert.Equal(t, td.lengths[tn], width)
				w = width

				assert.Equal(t, td.names[tn], msg.(fmt.Stringer).String(), "at %d (ev %d)", i, tn)
			}
		})
	}
}

// TestDetectRandomSequencesLex checks that the map-based sequence
// detector works over concatenations of random sequences.
func TestDetectRandomSequencesMap(t *testing.T) {
	runTestDetectSequence(t, detectSequence)
}

// BenchmarkDetectSequenceMap benchmarks the map-based sequence
// detector.
func BenchmarkDetectSequenceMap(b *testing.B) {
	td := genRandomDataWithSeed(123, 10000)
	for i := 0; i < b.N; i++ {
		for j, w := 0, 0; j < len(td.data); j += w {
			_, w, _ = detectSequence(td.data[j:])
		}
	}
}
