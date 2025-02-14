package tea

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestKeyString(t *testing.T) {
	t.Run("alt+space", func(t *testing.T) {
		if got := KeyMsg(Key{
			Type: KeySpace,
			Alt:  true,
		}).String(); got != "alt+ " {
			t.Fatalf(`expected a "alt+ ", got %q`, got)
		}
	})

	t.Run("runes", func(t *testing.T) {
		if got := KeyMsg(Key{
			Type:  KeyRunes,
			Runes: []rune{'a'},
		}).String(); got != "a" {
			t.Fatalf(`expected an "a", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if got := KeyMsg(Key{
			Type: KeyType(99999),
		}).String(); got != "" {
			t.Fatalf(`expected a "", got %q`, got)
		}
	})
}

func TestKeyTypeString(t *testing.T) {
	t.Run("space", func(t *testing.T) {
		if got := KeySpace.String(); got != " " {
			t.Fatalf(`expected a " ", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if got := KeyType(99999).String(); got != "" {
			t.Fatalf(`expected a "", got %q`, got)
		}
	})
}

type seqTest struct {
	seq []byte
	msg Msg
}

// buildBaseSeqTests returns sequence tests that are valid for the
// detectSequence() function.
func buildBaseSeqTests() []seqTest {
	td := []seqTest{}
	for seq, key := range sequences {
		key := key
		td = append(td, seqTest{[]byte(seq), KeyMsg(key)})
		if !key.Alt {
			key.Alt = true
			td = append(td, seqTest{[]byte("\x1b" + seq), KeyMsg(key)})
		}
	}
	// Add all the control characters.
	for i := keyNUL + 1; i <= keyDEL; i++ {
		if i == keyESC {
			// Not handled in detectSequence(), so not part of the base test
			// suite.
			continue
		}
		td = append(td, seqTest{[]byte{byte(i)}, KeyMsg{Type: i}})
		td = append(td, seqTest{[]byte{'\x1b', byte(i)}, KeyMsg{Type: i, Alt: true}})
		if i == keyUS {
			i = keyDEL - 1
		}
	}

	// Additional special cases.
	td = append(td,
		// Unrecognized CSI sequence.
		seqTest{
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			unknownCSISequenceMsg([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'}),
		},
		// A lone space character.
		seqTest{
			[]byte{' '},
			KeyMsg{Type: KeySpace, Runes: []rune(" ")},
		},
		// An escape character with the alt modifier.
		seqTest{
			[]byte{'\x1b', ' '},
			KeyMsg{Type: KeySpace, Runes: []rune(" "), Alt: true},
		},
	)
	return td
}

func TestDetectSequence(t *testing.T) {
	td := buildBaseSeqTests()
	for _, tc := range td {
		t.Run(fmt.Sprintf("%q", string(tc.seq)), func(t *testing.T) {
			hasSeq, width, msg := detectSequence(tc.seq)
			if !hasSeq {
				t.Fatalf("no sequence found")
			}
			if width != len(tc.seq) {
				t.Errorf("parser did not consume the entire input: got %d, expected %d", width, len(tc.seq))
			}
			if !reflect.DeepEqual(tc.msg, msg) {
				t.Errorf("expected event %#v (%T), got %#v (%T)", tc.msg, tc.msg, msg, msg)
			}
		})
	}
}

func TestDetectOneMsg(t *testing.T) {
	td := buildBaseSeqTests()
	// Add tests for the inputs that detectOneMsg() can parse, but
	// detectSequence() cannot.
	td = append(td,
		// focus/blur
		seqTest{
			[]byte{'\x1b', '[', 'I'},
			FocusMsg{},
		},
		seqTest{
			[]byte{'\x1b', '[', 'O'},
			BlurMsg{},
		},
		// Mouse event.
		seqTest{
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			MouseMsg{X: 32, Y: 16, Type: MouseWheelUp, Button: MouseButtonWheelUp, Action: MouseActionPress},
		},
		// SGR Mouse event.
		seqTest{
			[]byte("\x1b[<0;33;17M"),
			MouseMsg{X: 32, Y: 16, Type: MouseLeft, Button: MouseButtonLeft, Action: MouseActionPress},
		},
		// Runes.
		seqTest{
			[]byte{'a'},
			KeyMsg{Type: KeyRunes, Runes: []rune("a")},
		},
		seqTest{
			[]byte{'\x1b', 'a'},
			KeyMsg{Type: KeyRunes, Runes: []rune("a"), Alt: true},
		},
		seqTest{
			[]byte{'a', 'a', 'a'},
			KeyMsg{Type: KeyRunes, Runes: []rune("aaa")},
		},
		// Multi-byte rune.
		seqTest{
			[]byte("☃"),
			KeyMsg{Type: KeyRunes, Runes: []rune("☃")},
		},
		seqTest{
			[]byte("\x1b☃"),
			KeyMsg{Type: KeyRunes, Runes: []rune("☃"), Alt: true},
		},
		// Standalone control chacters.
		seqTest{
			[]byte{'\x1b'},
			KeyMsg{Type: KeyEscape},
		},
		seqTest{
			[]byte{byte(keySOH)},
			KeyMsg{Type: KeyCtrlA},
		},
		seqTest{
			[]byte{'\x1b', byte(keySOH)},
			KeyMsg{Type: KeyCtrlA, Alt: true},
		},
		seqTest{
			[]byte{byte(keyNUL)},
			KeyMsg{Type: KeyCtrlAt},
		},
		seqTest{
			[]byte{'\x1b', byte(keyNUL)},
			KeyMsg{Type: KeyCtrlAt, Alt: true},
		},
		// Invalid characters.
		seqTest{
			[]byte{'\x80'},
			unknownInputByteMsg(0x80),
		},
	)

	if runtime.GOOS != "windows" {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		td = append(td, seqTest{
			[]byte{'\xfe'},
			unknownInputByteMsg(0xfe),
		})
	}

	for _, tc := range td {
		t.Run(fmt.Sprintf("%q", string(tc.seq)), func(t *testing.T) {
			width, msg := detectOneMsg(tc.seq, false /* canHaveMoreData */)
			if width != len(tc.seq) {
				t.Errorf("parser did not consume the entire input: got %d, expected %d", width, len(tc.seq))
			}
			if !reflect.DeepEqual(tc.msg, msg) {
				t.Errorf("expected event %#v (%T), got %#v (%T)", tc.msg, tc.msg, msg, msg)
			}
		})
	}
}

func TestReadLongInput(t *testing.T) {
	input := strings.Repeat("a", 1000)
	msgs := testReadInputs(t, bytes.NewReader([]byte(input)))
	if len(msgs) != 1 {
		t.Errorf("expected 1 messages, got %d", len(msgs))
	}
	km := msgs[0]
	k := Key(km.(KeyMsg))
	if k.Type != KeyRunes {
		t.Errorf("expected key runes, got %d", k.Type)
	}
	if len(k.Runes) != 1000 || !reflect.DeepEqual(k.Runes, []rune(input)) {
		t.Errorf("unexpected runes: %+v", k)
	}
	if k.Alt {
		t.Errorf("unexpected alt")
	}
}

type chunkedBytesReader struct {
	data     []byte
	segments []int
	current  int
}

func (cr *chunkedBytesReader) Read(p []byte) (int, error) {
	if cr.current >= len(cr.segments)-1 {
		return 0, io.EOF
	}
	data := cr.data[cr.segments[cr.current]:cr.segments[cr.current+1]]
	n, err := bytes.NewReader(data).Read(p)
	cr.current++
	return n, err
}

func TestReadInput(t *testing.T) {
	type test struct {
		keyname  string
		in       []byte
		out      []Msg
		segments []int
	}
	type testOptionFunc func(*test)
	newTest := func(keyname string, in []byte, out []Msg, opts ...testOptionFunc) test {
		t := &test{keyname: keyname, in: in, out: out}
		for _, opt := range opts {
			opt(t)
		}
		return *t
	}
	withSegment := func(segments ...int) testOptionFunc {
		return func(t *test) { t.segments = segments }
	}

	testData := []test{
		newTest(
			"a",
			[]byte{'a'},
			[]Msg{
				KeyMsg{
					Type:  KeyRunes,
					Runes: []rune{'a'},
				},
			},
		),
		newTest(
			" ",
			[]byte{' '},
			[]Msg{
				KeyMsg{
					Type:  KeySpace,
					Runes: []rune{' '},
				},
			},
		),
		newTest(
			"a alt+a",
			[]byte{'a', '\x1b', 'a'},
			[]Msg{
				KeyMsg{Type: KeyRunes, Runes: []rune{'a'}},
				KeyMsg{Type: KeyRunes, Runes: []rune{'a'}, Alt: true},
			},
		),
		newTest(
			"a alt+a a",
			[]byte{'a', '\x1b', 'a', 'a'},
			[]Msg{
				KeyMsg{Type: KeyRunes, Runes: []rune{'a'}},
				KeyMsg{Type: KeyRunes, Runes: []rune{'a'}, Alt: true},
				KeyMsg{Type: KeyRunes, Runes: []rune{'a'}},
			},
		),
		newTest(
			"ctrl+a",
			[]byte{byte(keySOH)},
			[]Msg{
				KeyMsg{
					Type: KeyCtrlA,
				},
			},
		),
		newTest(
			"ctrl+a ctrl+b",
			[]byte{byte(keySOH), byte(keySTX)},
			[]Msg{
				KeyMsg{Type: KeyCtrlA},
				KeyMsg{Type: KeyCtrlB},
			},
		),
		newTest(
			"alt+a",
			[]byte{byte(0x1b), 'a'},
			[]Msg{
				KeyMsg{
					Type:  KeyRunes,
					Alt:   true,
					Runes: []rune{'a'},
				},
			},
		),
		newTest(
			"abcd",
			[]byte{'a', 'b', 'c', 'd'},
			[]Msg{
				KeyMsg{
					Type:  KeyRunes,
					Runes: []rune{'a', 'b', 'c', 'd'},
				},
			},
		),
		newTest(
			"up",
			[]byte("\x1b[A"),
			[]Msg{
				KeyMsg{
					Type: KeyUp,
				},
			},
		),
		newTest(
			"wheel up",
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Msg{
				MouseMsg{
					X:      32,
					Y:      16,
					Type:   MouseWheelUp,
					Button: MouseButtonWheelUp,
					Action: MouseActionPress,
				},
			},
		),
		newTest(
			"left motion release",
			[]byte{
				'\x1b', '[', 'M', byte(32) + 0b0010_0000, byte(32 + 33), byte(16 + 33),
				'\x1b', '[', 'M', byte(32) + 0b0000_0011, byte(64 + 33), byte(32 + 33),
			},
			[]Msg{
				MouseMsg(MouseEvent{
					X:      32,
					Y:      16,
					Type:   MouseLeft,
					Button: MouseButtonLeft,
					Action: MouseActionMotion,
				}),
				MouseMsg(MouseEvent{
					X:      64,
					Y:      32,
					Type:   MouseRelease,
					Button: MouseButtonNone,
					Action: MouseActionRelease,
				}),
			},
		),
		newTest(
			"shift+tab",
			[]byte{'\x1b', '[', 'Z'},
			[]Msg{
				KeyMsg{
					Type: KeyShiftTab,
				},
			},
		),
		newTest(
			"enter",
			[]byte{'\r'},
			[]Msg{KeyMsg{Type: KeyEnter}},
		),
		newTest(
			"alt+enter",
			[]byte{'\x1b', '\r'},
			[]Msg{
				KeyMsg{
					Type: KeyEnter,
					Alt:  true,
				},
			},
		),
		newTest(
			"insert",
			[]byte{'\x1b', '[', '2', '~'},
			[]Msg{
				KeyMsg{
					Type: KeyInsert,
				},
			},
		),
		newTest(
			"alt+ctrl+a",
			[]byte{'\x1b', byte(keySOH)},
			[]Msg{
				KeyMsg{
					Type: KeyCtrlA,
					Alt:  true,
				},
			},
		),
		newTest(
			"?CSI[45 45 45 45 88]?",
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Msg{unknownCSISequenceMsg([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'})},
		),
		// Powershell sequences.
		newTest(
			"up",
			[]byte{'\x1b', 'O', 'A'},
			[]Msg{KeyMsg{Type: KeyUp}},
		),
		newTest(
			"down",
			[]byte{'\x1b', 'O', 'B'},
			[]Msg{KeyMsg{Type: KeyDown}},
		),
		newTest(
			"right",
			[]byte{'\x1b', 'O', 'C'},
			[]Msg{KeyMsg{Type: KeyRight}},
		),
		newTest(
			"left",
			[]byte{'\x1b', 'O', 'D'},
			[]Msg{KeyMsg{Type: KeyLeft}},
		),
		newTest(
			"alt+enter",
			[]byte{'\x1b', '\x0d'},
			[]Msg{KeyMsg{Type: KeyEnter, Alt: true}},
		),
		newTest(
			"alt+backspace",
			[]byte{'\x1b', '\x7f'},
			[]Msg{KeyMsg{Type: KeyBackspace, Alt: true}},
		),
		newTest(
			"ctrl+@",
			[]byte{'\x00'},
			[]Msg{KeyMsg{Type: KeyCtrlAt}},
		),
		newTest(
			"alt+ctrl+@",
			[]byte{'\x1b', '\x00'},
			[]Msg{KeyMsg{Type: KeyCtrlAt, Alt: true}},
		),
		newTest(
			"esc",
			[]byte{'\x1b'},
			[]Msg{KeyMsg{Type: KeyEsc}},
		),
		newTest(
			"alt+esc",
			[]byte{'\x1b', '\x1b'},
			[]Msg{KeyMsg{Type: KeyEsc, Alt: true}},
		),
		newTest(
			"[a b] o",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', ' ', 'b',
				'\x1b', '[', '2', '0', '1', '~',
				'o',
			},
			[]Msg{
				KeyMsg{Type: KeyRunes, Runes: []rune("a b"), Paste: true},
				KeyMsg{Type: KeyRunes, Runes: []rune("o")},
			},
		),
		newTest(
			"[a\x03\nb]",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', '\x03', '\n', 'b',
				'\x1b', '[', '2', '0', '1', '~',
			},
			[]Msg{
				KeyMsg{Type: KeyRunes, Runes: []rune("a\x03\nb"), Paste: true},
			},
		),
	}
	if runtime.GOOS != "windows" {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		testData = append(testData,
			newTest(
				"?0xfe?",
				[]byte{'\xfe'},
				[]Msg{unknownInputByteMsg(0xfe)},
			),
			newTest(
				"a ?0xfe?   b",
				[]byte{'a', '\xfe', ' ', 'b'},
				[]Msg{
					KeyMsg{Type: KeyRunes, Runes: []rune{'a'}},
					unknownInputByteMsg(0xfe),
					KeyMsg{Type: KeySpace, Runes: []rune{' '}},
					KeyMsg{Type: KeyRunes, Runes: []rune{'b'}},
				},
			),
		)
	}

	// Incomplete UTF-8 sequences
	testData = append(testData,
		// 2-bytes
		newTest(
			"α βγ",
			[]byte{
				'\xce', '\xb1', // α
				'\xce', '\xb2', // β (splitted)
				'\xce', '\xb3', // γ
			},
			[]Msg{
				KeyMsg{Type: KeyRunes, Runes: []rune{'α'}},
				KeyMsg{Type: KeyRunes, Runes: []rune{'β', 'γ'}},
			},
			withSegment(3),
		),
		// 3-bytes
		newTest(
			"一 二三",
			[]byte{
				'\xe4', '\xb8', '\x80', // 一
				'\xe4', '\xba', '\x8c', // 二 (splitted)
				'\xe4', '\xb8', '\x89', // 三
			},
			[]Msg{
				KeyMsg{Type: KeyRunes, Runes: []rune{'一'}},
				KeyMsg{Type: KeyRunes, Runes: []rune{'二', '三'}},
			},
			withSegment(4, 5),
		),
		// 4-bytes
		newTest(
			"🧋 🫧📼",
			[]byte{
				'\xf0', '\x9f', '\xa7', '\x8b', // 🧋
				'\xf0', '\x9f', '\xab', '\xa7', // 🫧 (splitted)
				'\xf0', '\x9f', '\x93', '\xbc', // 📼
			},
			[]Msg{
				KeyMsg{Type: KeyRunes, Runes: []rune{'🧋'}},
				KeyMsg{Type: KeyRunes, Runes: []rune{'🫧', '📼'}},
			},
			withSegment(5, 6, 7),
		),
	)

	newBytesReader := func(bs []byte, segments []int) io.Reader {
		if segments == nil {
			return bytes.NewReader(bs)
		}
		segments = append([]int{0}, segments...)
		segments = append(segments, len(bs))
		return &chunkedBytesReader{
			data:     bs,
			segments: segments,
			current:  0,
		}
	}
	for i, td := range testData {
		t.Run(fmt.Sprintf("%d: %s", i, td.keyname), func(t *testing.T) {
			msgs := testReadInputs(t, newBytesReader(td.in, td.segments))
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

			title := buf.String()
			if title != td.keyname {
				t.Errorf("expected message titles:\n  %s\ngot:\n  %s", td.keyname, title)
			}

			if len(msgs) != len(td.out) {
				t.Fatalf("unexpected message list length: got %d, expected %d\n%#v", len(msgs), len(td.out), msgs)
			}

			if !reflect.DeepEqual(td.out, msgs) {
				t.Fatalf("expected:\n%#v\ngot:\n%#v", td.out, msgs)
			}
		})
	}
}

func testReadInputs(t *testing.T, input io.Reader) []Msg {
	// We'll check that the input reader finishes at the end
	// without error.
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
		inputErr = readAnsiInputs(ctx, msgsC, input)
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
	src := rand.NewSource(s)
	r := rand.New(src)

	// allseqs contains all the sequences, in sorted order. We sort
	// to make the test deterministic (when the seed is also fixed).
	type seqpair struct {
		seq  string
		name string
	}
	var allseqs []seqpair
	for seq, key := range sequences {
		allseqs = append(allseqs, seqpair{seq, key.String()})
	}
	sort.Slice(allseqs, func(i, j int) bool { return allseqs[i].seq < allseqs[j].seq })

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
			res.data = append(res.data, 1)
			res.names = append(res.names, prefix+"ctrl+a")
			res.lengths = append(res.lengths, 1+esclen)

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
			res.data = append(res.data, s.seq...)
			res.names = append(res.names, prefix+s.name)
			res.lengths = append(res.lengths, len(s.seq)+esclen)
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
	t *testing.T, detectSequence func(input []byte) (hasSeq bool, width int, msg Msg),
) {
	for i := 0; i < 10; i++ {
		t.Run("", func(t *testing.T) {
			td := genRandomData(func(s int64) { t.Logf("using random seed: %d", s) }, 1000)

			t.Logf("%#v", td)

			// tn is the event number in td.
			// i is the cursor in the input data.
			// w is the length of the last sequence detected.
			for tn, i, w := 0, 0, 0; i < len(td.data); tn, i = tn+1, i+w {
				hasSequence, width, msg := detectSequence(td.data[i:])
				if !hasSequence {
					t.Fatalf("at %d (ev %d): failed to find sequence", i, tn)
				}
				if width != td.lengths[tn] {
					t.Errorf("at %d (ev %d): expected width %d, got %d", i, tn, td.lengths[tn], width)
				}
				w = width

				s, ok := msg.(fmt.Stringer)
				if !ok {
					t.Errorf("at %d (ev %d): expected stringer event, got %T", i, tn, msg)
				} else {
					if td.names[tn] != s.String() {
						t.Errorf("at %d (ev %d): expected event %q, got %q", i, tn, td.names[tn], s.String())
					}
				}
			}
		})
	}
}

// TestDetectRandomSequencesMap checks that the map-based sequence
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
