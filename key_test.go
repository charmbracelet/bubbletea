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
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

var sequences = buildKeysTable(_FlagTerminfo, "dumb")

func TestKeyString(t *testing.T) {
	t.Run("alt+space", func(t *testing.T) {
		k := KeyPressMsg{Type: KeySpace, Runes: []rune{' '}, Mod: ModAlt}
		if got := k.String(); got != "alt+space" {
			t.Fatalf(`expected a "alt+space ", got %q`, got)
		}
	})

	t.Run("runes", func(t *testing.T) {
		k := KeyPressMsg{Runes: []rune{'a'}}
		if got := k.String(); got != "a" {
			t.Fatalf(`expected an "a", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		k := KeyPressMsg{Type: 99999}
		if got := k.String(); got != "" {
			t.Fatalf(`expected a "unknown", got %q`, got)
		}
	})
}

func TestKeyTypeString(t *testing.T) {
	t.Run("space", func(t *testing.T) {
		if got := KeySpace.String(); got != "space" {
			t.Fatalf(`expected a "space", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if got := KeyType(99999).String(); got != "" {
			t.Fatalf(`expected a "unknown", got %q`, got)
		}
	})
}

type seqTest struct {
	seq  []byte
	msgs []Msg
}

var f3CurPosRegexp = regexp.MustCompile(`\x1b\[1;(\d+)R`)

// buildBaseSeqTests returns sequence tests that are valid for the
// detectSequence() function.
func buildBaseSeqTests() []seqTest {
	td := []seqTest{}
	for seq, key := range sequences {
		k := KeyPressMsg(key)
		st := seqTest{seq: []byte(seq), msgs: []Msg{k}}

		// XXX: This is a special case to handle F3 key sequence and cursor
		// position report having the same sequence. See [parseCsi] for more
		// information.
		if f3CurPosRegexp.MatchString(seq) {
			st.msgs = []Msg{k, CursorPositionMsg{Row: 1, Column: int(key.Mod) + 1}}
		}
		td = append(td, st)
	}

	// Additional special cases.
	td = append(td,
		// Unrecognized CSI sequence.
		seqTest{
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Msg{
				UnknownMsg([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'}),
			},
		},
		// A lone space character.
		seqTest{
			[]byte{' '},
			[]Msg{
				KeyPressMsg{Type: KeySpace, Runes: []rune{' '}},
			},
		},
		// An escape character with the alt modifier.
		seqTest{
			[]byte{'\x1b', ' '},
			[]Msg{
				KeyPressMsg{Type: KeySpace, Runes: []rune{' '}, Mod: ModAlt},
			},
		},
	)
	return td
}

func TestParseSequence(t *testing.T) {
	td := buildBaseSeqTests()
	td = append(td,
		// Xterm modifyOtherKeys CSI 27 ; <modifier> ; <code> ~
		seqTest{
			[]byte("\x1b[27;3;20320~"),
			[]Msg{KeyPressMsg{Runes: []rune{'你'}, Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;65~"),
			[]Msg{KeyPressMsg{Runes: []rune{'A'}, Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;8~"),
			[]Msg{KeyPressMsg{Type: KeyBackspace, Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;27~"),
			[]Msg{KeyPressMsg{Type: KeyEscape, Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;127~"),
			[]Msg{KeyPressMsg{Type: KeyBackspace, Mod: ModAlt}},
		},

		// Kitty keyboard / CSI u (fixterms)
		seqTest{
			[]byte("\x1b[1B"),
			[]Msg{KeyPressMsg{Type: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[1;B"),
			[]Msg{KeyPressMsg{Type: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[1;4B"),
			[]Msg{KeyPressMsg{Mod: ModShift | ModAlt, Type: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[8~"),
			[]Msg{KeyPressMsg{Type: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[8;~"),
			[]Msg{KeyPressMsg{Type: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[8;10~"),
			[]Msg{KeyPressMsg{Mod: ModShift | ModMeta, Type: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[27;4u"),
			[]Msg{KeyPressMsg{Mod: ModShift | ModAlt, Type: KeyEscape}},
		},
		seqTest{
			[]byte("\x1b[127;4u"),
			[]Msg{KeyPressMsg{Mod: ModShift | ModAlt, Type: KeyBackspace}},
		},
		seqTest{
			[]byte("\x1b[57358;4u"),
			[]Msg{KeyPressMsg{Mod: ModShift | ModAlt, Type: KeyCapsLock}},
		},
		seqTest{
			[]byte("\x1b[9;2u"),
			[]Msg{KeyPressMsg{Mod: ModShift, Type: KeyTab}},
		},
		seqTest{
			[]byte("\x1b[195;u"),
			[]Msg{KeyPressMsg{Runes: []rune{'Ã'}, Type: KeyRunes}},
		},
		seqTest{
			[]byte("\x1b[20320;2u"),
			[]Msg{KeyPressMsg{Runes: []rune{'你'}, Mod: ModShift, Type: KeyRunes}},
		},
		seqTest{
			[]byte("\x1b[195;:1u"),
			[]Msg{KeyPressMsg{Runes: []rune{'Ã'}, Type: KeyRunes}},
		},
		seqTest{
			[]byte("\x1b[195;2:3u"),
			[]Msg{KeyReleaseMsg{Runes: []rune{'Ã'}, Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:2u"),
			[]Msg{KeyPressMsg{Runes: []rune{'Ã'}, IsRepeat: true, Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:1u"),
			[]Msg{KeyPressMsg{Runes: []rune{'Ã'}, Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:3u"),
			[]Msg{KeyReleaseMsg{Runes: []rune{'Ã'}, Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[97;2;65u"),
			[]Msg{KeyPressMsg{Runes: []rune{'A'}, Mod: ModShift, altRune: 'a'}},
		},
		seqTest{
			[]byte("\x1b[97;;229u"),
			[]Msg{KeyPressMsg{Runes: []rune{'å'}, altRune: 'a'}},
		},

		// focus/blur
		seqTest{
			[]byte{'\x1b', '[', 'I'},
			[]Msg{
				FocusMsg{},
			},
		},
		seqTest{
			[]byte{'\x1b', '[', 'O'},
			[]Msg{
				BlurMsg{},
			},
		},
		// Mouse event.
		seqTest{
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Msg{
				MouseWheelMsg{X: 32, Y: 16, Button: MouseWheelUp},
			},
		},
		// SGR Mouse event.
		seqTest{
			[]byte("\x1b[<0;33;17M"),
			[]Msg{
				MouseClickMsg{X: 32, Y: 16, Button: MouseLeft},
			},
		},
		// Runes.
		seqTest{
			[]byte{'a'},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}},
			},
		},
		seqTest{
			[]byte{'\x1b', 'a'},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}, Mod: ModAlt},
			},
		},
		seqTest{
			[]byte{'a', 'a', 'a'},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}},
				KeyPressMsg{Runes: []rune{'a'}},
				KeyPressMsg{Runes: []rune{'a'}},
			},
		},
		// Multi-byte rune.
		seqTest{
			[]byte("☃"),
			[]Msg{
				KeyPressMsg{Runes: []rune{'☃'}},
			},
		},
		seqTest{
			[]byte("\x1b☃"),
			[]Msg{
				KeyPressMsg{Runes: []rune{'☃'}, Mod: ModAlt},
			},
		},
		// Standalone control chacters.
		seqTest{
			[]byte{'\x1b'},
			[]Msg{
				KeyPressMsg{Type: KeyEscape},
			},
		},
		seqTest{
			[]byte{ansi.SOH},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}, Mod: ModCtrl},
			},
		},
		seqTest{
			[]byte{'\x1b', ansi.SOH},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}, Mod: ModCtrl | ModAlt},
			},
		},
		seqTest{
			[]byte{ansi.NUL},
			[]Msg{
				KeyPressMsg{Runes: []rune{' '}, Type: KeySpace, Mod: ModCtrl},
			},
		},
		seqTest{
			[]byte{'\x1b', ansi.NUL},
			[]Msg{
				KeyPressMsg{Runes: []rune{' '}, Type: KeySpace, Mod: ModCtrl | ModAlt},
			},
		},
		// C1 control characters.
		seqTest{
			[]byte{'\x80'},
			[]Msg{
				KeyPressMsg{Runes: []rune{0x80 - '@'}, Mod: ModCtrl | ModAlt},
			},
		},
	)

	if runtime.GOOS != "windows" {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		td = append(td, seqTest{
			[]byte{'\xfe'},
			[]Msg{
				UnknownMsg(rune(0xfe)),
			},
		})
	}

	for _, tc := range td {
		t.Run(fmt.Sprintf("%q", string(tc.seq)), func(t *testing.T) {
			var events []Msg
			buf := tc.seq
			for len(buf) > 0 {
				width, msg := parseSequence(buf)
				switch msg := msg.(type) {
				case multiMsg:
					events = append(events, msg...)
				default:
					events = append(events, msg)
				}
				buf = buf[width:]
			}
			if !reflect.DeepEqual(tc.msgs, events) {
				t.Errorf("\nexpected event:\n    %#v\ngot:\n    %#v", tc.msgs, events)
			}
		})
	}
}

func TestReadLongInput(t *testing.T) {
	expect := make([]Msg, 1000)
	for i := 0; i < 1000; i++ {
		expect[i] = KeyPressMsg{Runes: []rune{'a'}}
	}
	input := strings.Repeat("a", 1000)
	drv, err := newDriver(strings.NewReader(input), "dumb", 0)
	if err != nil {
		t.Fatalf("unexpected input driver error: %v", err)
	}

	var msgs []Msg
	for {
		events, err := drv.ReadEvents()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unexpected input error: %v", err)
		}
		msgs = append(msgs, events...)
	}

	if !reflect.DeepEqual(expect, msgs) {
		t.Errorf("unexpected messages, expected:\n    %+v\ngot:\n    %+v", expect, msgs)
	}
}

func TestReadInput(t *testing.T) {
	type test struct {
		keyname string
		in      []byte
		out     []Msg
	}
	testData := []test{
		{
			"a",
			[]byte{'a'},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}},
			},
		},
		{
			"space",
			[]byte{' '},
			[]Msg{
				KeyPressMsg{Type: KeySpace, Runes: []rune{' '}},
			},
		},
		{
			"a alt+a",
			[]byte{'a', '\x1b', 'a'},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}},
				KeyPressMsg{Runes: []rune{'a'}, Mod: ModAlt},
			},
		},
		{
			"a alt+a a",
			[]byte{'a', '\x1b', 'a', 'a'},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}},
				KeyPressMsg{Runes: []rune{'a'}, Mod: ModAlt},
				KeyPressMsg{Runes: []rune{'a'}},
			},
		},
		{
			"ctrl+a",
			[]byte{byte(ansi.SOH)},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}, Mod: ModCtrl},
			},
		},
		{
			"ctrl+a ctrl+b",
			[]byte{byte(ansi.SOH), byte(ansi.STX)},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}, Mod: ModCtrl},
				KeyPressMsg{Runes: []rune{'b'}, Mod: ModCtrl},
			},
		},
		{
			"alt+a",
			[]byte{byte(0x1b), 'a'},
			[]Msg{
				KeyPressMsg{Mod: ModAlt, Runes: []rune{'a'}},
			},
		},
		{
			"a b c d",
			[]byte{'a', 'b', 'c', 'd'},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}},
				KeyPressMsg{Runes: []rune{'b'}},
				KeyPressMsg{Runes: []rune{'c'}},
				KeyPressMsg{Runes: []rune{'d'}},
			},
		},
		{
			"up",
			[]byte("\x1b[A"),
			[]Msg{
				KeyPressMsg{Type: KeyUp},
			},
		},
		{
			"wheel up",
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Msg{
				MouseWheelMsg{X: 32, Y: 16, Button: MouseWheelUp},
			},
		},
		{
			"left motion release",
			[]byte{
				'\x1b', '[', 'M', byte(32) + 0b0010_0000, byte(32 + 33), byte(16 + 33),
				'\x1b', '[', 'M', byte(32) + 0b0000_0011, byte(64 + 33), byte(32 + 33),
			},
			[]Msg{
				MouseMotionMsg{X: 32, Y: 16, Button: MouseLeft},
				MouseReleaseMsg{X: 64, Y: 32, Button: MouseNone},
			},
		},
		{
			"shift+tab",
			[]byte{'\x1b', '[', 'Z'},
			[]Msg{
				KeyPressMsg{Type: KeyTab, Mod: ModShift},
			},
		},
		{
			"enter",
			[]byte{'\r'},
			[]Msg{KeyPressMsg{Type: KeyEnter}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\r'},
			[]Msg{
				KeyPressMsg{Type: KeyEnter, Mod: ModAlt},
			},
		},
		{
			"insert",
			[]byte{'\x1b', '[', '2', '~'},
			[]Msg{
				KeyPressMsg{Type: KeyInsert},
			},
		},
		{
			"ctrl+alt+a",
			[]byte{'\x1b', byte(ansi.SOH)},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}, Mod: ModCtrl | ModAlt},
			},
		},
		{
			"CSI?----X?",
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Msg{UnknownMsg([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'})},
		},
		// Powershell sequences.
		{
			"up",
			[]byte{'\x1b', 'O', 'A'},
			[]Msg{KeyPressMsg{Type: KeyUp}},
		},
		{
			"down",
			[]byte{'\x1b', 'O', 'B'},
			[]Msg{KeyPressMsg{Type: KeyDown}},
		},
		{
			"right",
			[]byte{'\x1b', 'O', 'C'},
			[]Msg{KeyPressMsg{Type: KeyRight}},
		},
		{
			"left",
			[]byte{'\x1b', 'O', 'D'},
			[]Msg{KeyPressMsg{Type: KeyLeft}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\x0d'},
			[]Msg{KeyPressMsg{Type: KeyEnter, Mod: ModAlt}},
		},
		{
			"alt+backspace",
			[]byte{'\x1b', '\x7f'},
			[]Msg{KeyPressMsg{Type: KeyBackspace, Mod: ModAlt}},
		},
		{
			"ctrl+space",
			[]byte{'\x00'},
			[]Msg{KeyPressMsg{Type: KeySpace, Runes: []rune{' '}, Mod: ModCtrl}},
		},
		{
			"ctrl+alt+space",
			[]byte{'\x1b', '\x00'},
			[]Msg{KeyPressMsg{Type: KeySpace, Runes: []rune{' '}, Mod: ModCtrl | ModAlt}},
		},
		{
			"esc",
			[]byte{'\x1b'},
			[]Msg{KeyPressMsg{Type: KeyEscape}},
		},
		{
			"alt+esc",
			[]byte{'\x1b', '\x1b'},
			[]Msg{KeyPressMsg{Type: KeyEscape, Mod: ModAlt}},
		},
		{
			"a b o",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', ' ', 'b',
				'\x1b', '[', '2', '0', '1', '~',
				'o',
			},
			[]Msg{
				PasteStartMsg{},
				PasteMsg("a b"),
				PasteEndMsg{},
				KeyPressMsg{Runes: []rune{'o'}},
			},
		},
		{
			"a\x03\nb",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', '\x03', '\n', 'b',
				'\x1b', '[', '2', '0', '1', '~',
			},
			[]Msg{
				PasteStartMsg{},
				PasteMsg("a\x03\nb"),
				PasteEndMsg{},
			},
		},
		{
			"?0xfe?",
			[]byte{'\xfe'},
			[]Msg{
				UnknownMsg(rune(0xfe)),
			},
		},
		{
			"a ?0xfe?   b",
			[]byte{'a', '\xfe', ' ', 'b'},
			[]Msg{
				KeyPressMsg{Runes: []rune{'a'}},
				UnknownMsg(rune(0xfe)),
				KeyPressMsg{Type: KeySpace, Runes: []rune{' '}},
				KeyPressMsg{Runes: []rune{'b'}},
			},
		},
	}

	for i, td := range testData {
		t.Run(fmt.Sprintf("%d: %s", i, td.keyname), func(t *testing.T) {
			msgs := testReadInputs(t, bytes.NewReader(td.in))
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

			if len(msgs) != len(td.out) {
				t.Fatalf("unexpected message list length: got %d, expected %d\n  got: %#v\n  expected: %#v\n", len(msgs), len(td.out), msgs, td.out)
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

	dr, err := newDriver(input, "dumb", 0)
	if err != nil {
		t.Fatalf("unexpected input driver error: %v", err)
	}

	// The messages we're consuming.
	msgsC := make(chan Msg)

	// Start the reader in the background.
	wg.Add(1)
	go func() {
		defer wg.Done()
		var events []Msg
		events, inputErr = dr.ReadEvents()
	out:
		for _, ev := range events {
			select {
			case msgsC <- ev:
			case <-ctx.Done():
				break out
			}
		}
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
			res.names = append(res.names, "ctrl+"+prefix+"a")
			res.lengths = append(res.lengths, 1+esclen)

		case 1, 2:
			// A sequence.
			seqi := r.Intn(len(allseqs))
			s := allseqs[seqi]
			if strings.Contains(s.name, "alt+") || strings.Contains(s.name, "meta+") {
				esclen = 0
				prefix = ""
				alt = 0
			}
			if alt == 1 {
				res.data = append(res.data, '\x1b')
			}
			res.data = append(res.data, s.seq...)
			if strings.HasPrefix(s.name, "ctrl+") {
				prefix = "ctrl+" + prefix
			}
			name := prefix + strings.TrimPrefix(s.name, "ctrl+")
			res.names = append(res.names, name)
			res.lengths = append(res.lengths, len(s.seq)+esclen)
		}
	}
	return res
}

func FuzzParseSequence(f *testing.F) {
	for seq := range sequences {
		f.Add(seq)
	}
	f.Add("\x1b]52;?\x07")                      // OSC 52
	f.Add("\x1b]11;rgb:0000/0000/0000\x1b\\")   // OSC 11
	f.Add("\x1bP>|charm terminal(0.1.2)\x1b\\") // DCS (XTVERSION)
	f.Add("\x1b_Gi=123\x1b\\")                  // APC
	f.Fuzz(func(t *testing.T, seq string) {
		n, _ := parseSequence([]byte(seq))
		if n == 0 && seq != "" {
			t.Errorf("expected a non-zero width for %q", seq)
		}
	})
}

// BenchmarkDetectSequenceMap benchmarks the map-based sequence
// detector.
func BenchmarkDetectSequenceMap(b *testing.B) {
	td := genRandomDataWithSeed(123, 10000)
	for i := 0; i < b.N; i++ {
		for j, w := 0, 0; j < len(td.data); j += w {
			w, _ = parseSequence(td.data[j:])
		}
	}
}
