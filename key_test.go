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
		k := KeyPressMsg{typ: KeySpace, runes: []rune{' '}, mod: ModAlt}
		if got := k.String(); got != "alt+space" {
			t.Fatalf(`expected a "alt+space ", got %q`, got)
		}
	})

	t.Run("runes", func(t *testing.T) {
		k := KeyPressMsg{runes: []rune{'a'}}
		if got := k.String(); got != "a" {
			t.Fatalf(`expected an "a", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		k := KeyPressMsg{typ: 99999}
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
			st.msgs = []Msg{k, CursorPositionMsg{Row: 1, Column: int(key.mod) + 1}}
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
				KeyPressMsg{typ: KeySpace, runes: []rune{' '}},
			},
		},
		// An escape character with the alt modifier.
		seqTest{
			[]byte{'\x1b', ' '},
			[]Msg{
				KeyPressMsg{typ: KeySpace, runes: []rune{' '}, mod: ModAlt},
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
			[]Msg{KeyPressMsg{runes: []rune{'你'}, mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;65~"),
			[]Msg{KeyPressMsg{runes: []rune{'A'}, mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;8~"),
			[]Msg{KeyPressMsg{typ: KeyBackspace, mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;27~"),
			[]Msg{KeyPressMsg{typ: KeyEscape, mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;127~"),
			[]Msg{KeyPressMsg{typ: KeyBackspace, mod: ModAlt}},
		},

		// Kitty keyboard / CSI u (fixterms)
		seqTest{
			[]byte("\x1b[1B"),
			[]Msg{KeyPressMsg{typ: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[1;B"),
			[]Msg{KeyPressMsg{typ: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[1;4B"),
			[]Msg{KeyPressMsg{mod: ModShift | ModAlt, typ: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[8~"),
			[]Msg{KeyPressMsg{typ: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[8;~"),
			[]Msg{KeyPressMsg{typ: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[8;10~"),
			[]Msg{KeyPressMsg{mod: ModShift | ModMeta, typ: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[27;4u"),
			[]Msg{KeyPressMsg{mod: ModShift | ModAlt, typ: KeyEscape}},
		},
		seqTest{
			[]byte("\x1b[127;4u"),
			[]Msg{KeyPressMsg{mod: ModShift | ModAlt, typ: KeyBackspace}},
		},
		seqTest{
			[]byte("\x1b[57358;4u"),
			[]Msg{KeyPressMsg{mod: ModShift | ModAlt, typ: KeyCapsLock}},
		},
		seqTest{
			[]byte("\x1b[9;2u"),
			[]Msg{KeyPressMsg{mod: ModShift, typ: KeyTab}},
		},
		seqTest{
			[]byte("\x1b[195;u"),
			[]Msg{KeyPressMsg{runes: []rune{'Ã'}, typ: KeyRunes}},
		},
		seqTest{
			[]byte("\x1b[20320;2u"),
			[]Msg{KeyPressMsg{runes: []rune{'你'}, mod: ModShift, typ: KeyRunes}},
		},
		seqTest{
			[]byte("\x1b[195;:1u"),
			[]Msg{KeyPressMsg{runes: []rune{'Ã'}, typ: KeyRunes}},
		},
		seqTest{
			[]byte("\x1b[195;2:3u"),
			[]Msg{KeyReleaseMsg{runes: []rune{'Ã'}, mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:2u"),
			[]Msg{KeyPressMsg{runes: []rune{'Ã'}, isRepeat: true, mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:1u"),
			[]Msg{KeyPressMsg{runes: []rune{'Ã'}, mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:3u"),
			[]Msg{KeyReleaseMsg{runes: []rune{'Ã'}, mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[97;2;65u"),
			[]Msg{KeyPressMsg{runes: []rune{'A'}, mod: ModShift, altRune: 'a'}},
		},
		seqTest{
			[]byte("\x1b[97;;229u"),
			[]Msg{KeyPressMsg{runes: []rune{'å'}, altRune: 'a'}},
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
				MouseWheelMsg{x: 32, y: 16, button: MouseWheelUp},
			},
		},
		// SGR Mouse event.
		seqTest{
			[]byte("\x1b[<0;33;17M"),
			[]Msg{
				MouseClickMsg{x: 32, y: 16, button: MouseLeft},
			},
		},
		// Runes.
		seqTest{
			[]byte{'a'},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}},
			},
		},
		seqTest{
			[]byte{'\x1b', 'a'},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}, mod: ModAlt},
			},
		},
		seqTest{
			[]byte{'a', 'a', 'a'},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}},
				KeyPressMsg{runes: []rune{'a'}},
				KeyPressMsg{runes: []rune{'a'}},
			},
		},
		// Multi-byte rune.
		seqTest{
			[]byte("☃"),
			[]Msg{
				KeyPressMsg{runes: []rune{'☃'}},
			},
		},
		seqTest{
			[]byte("\x1b☃"),
			[]Msg{
				KeyPressMsg{runes: []rune{'☃'}, mod: ModAlt},
			},
		},
		// Standalone control chacters.
		seqTest{
			[]byte{'\x1b'},
			[]Msg{
				KeyPressMsg{typ: KeyEscape},
			},
		},
		seqTest{
			[]byte{ansi.SOH},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}, mod: ModCtrl},
			},
		},
		seqTest{
			[]byte{'\x1b', ansi.SOH},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}, mod: ModCtrl | ModAlt},
			},
		},
		seqTest{
			[]byte{ansi.NUL},
			[]Msg{
				KeyPressMsg{runes: []rune{' '}, typ: KeySpace, mod: ModCtrl},
			},
		},
		seqTest{
			[]byte{'\x1b', ansi.NUL},
			[]Msg{
				KeyPressMsg{runes: []rune{' '}, typ: KeySpace, mod: ModCtrl | ModAlt},
			},
		},
		// C1 control characters.
		seqTest{
			[]byte{'\x80'},
			[]Msg{
				KeyPressMsg{runes: []rune{0x80 - '@'}, mod: ModCtrl | ModAlt},
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
		expect[i] = KeyPressMsg{runes: []rune{'a'}}
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
				KeyPressMsg{runes: []rune{'a'}},
			},
		},
		{
			"space",
			[]byte{' '},
			[]Msg{
				KeyPressMsg{typ: KeySpace, runes: []rune{' '}},
			},
		},
		{
			"a alt+a",
			[]byte{'a', '\x1b', 'a'},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}},
				KeyPressMsg{runes: []rune{'a'}, mod: ModAlt},
			},
		},
		{
			"a alt+a a",
			[]byte{'a', '\x1b', 'a', 'a'},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}},
				KeyPressMsg{runes: []rune{'a'}, mod: ModAlt},
				KeyPressMsg{runes: []rune{'a'}},
			},
		},
		{
			"ctrl+a",
			[]byte{byte(ansi.SOH)},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}, mod: ModCtrl},
			},
		},
		{
			"ctrl+a ctrl+b",
			[]byte{byte(ansi.SOH), byte(ansi.STX)},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}, mod: ModCtrl},
				KeyPressMsg{runes: []rune{'b'}, mod: ModCtrl},
			},
		},
		{
			"alt+a",
			[]byte{byte(0x1b), 'a'},
			[]Msg{
				KeyPressMsg{mod: ModAlt, runes: []rune{'a'}},
			},
		},
		{
			"a b c d",
			[]byte{'a', 'b', 'c', 'd'},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}},
				KeyPressMsg{runes: []rune{'b'}},
				KeyPressMsg{runes: []rune{'c'}},
				KeyPressMsg{runes: []rune{'d'}},
			},
		},
		{
			"up",
			[]byte("\x1b[A"),
			[]Msg{
				KeyPressMsg{typ: KeyUp},
			},
		},
		{
			"wheel up",
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Msg{
				MouseWheelMsg{x: 32, y: 16, button: MouseWheelUp},
			},
		},
		{
			"left motion release",
			[]byte{
				'\x1b', '[', 'M', byte(32) + 0b0010_0000, byte(32 + 33), byte(16 + 33),
				'\x1b', '[', 'M', byte(32) + 0b0000_0011, byte(64 + 33), byte(32 + 33),
			},
			[]Msg{
				MouseMotionMsg{x: 32, y: 16, button: MouseLeft},
				MouseReleaseMsg{x: 64, y: 32, button: MouseNone},
			},
		},
		{
			"shift+tab",
			[]byte{'\x1b', '[', 'Z'},
			[]Msg{
				KeyPressMsg{typ: KeyTab, mod: ModShift},
			},
		},
		{
			"enter",
			[]byte{'\r'},
			[]Msg{KeyPressMsg{typ: KeyEnter}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\r'},
			[]Msg{
				KeyPressMsg{typ: KeyEnter, mod: ModAlt},
			},
		},
		{
			"insert",
			[]byte{'\x1b', '[', '2', '~'},
			[]Msg{
				KeyPressMsg{typ: KeyInsert},
			},
		},
		{
			"ctrl+alt+a",
			[]byte{'\x1b', byte(ansi.SOH)},
			[]Msg{
				KeyPressMsg{runes: []rune{'a'}, mod: ModCtrl | ModAlt},
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
			[]Msg{KeyPressMsg{typ: KeyUp}},
		},
		{
			"down",
			[]byte{'\x1b', 'O', 'B'},
			[]Msg{KeyPressMsg{typ: KeyDown}},
		},
		{
			"right",
			[]byte{'\x1b', 'O', 'C'},
			[]Msg{KeyPressMsg{typ: KeyRight}},
		},
		{
			"left",
			[]byte{'\x1b', 'O', 'D'},
			[]Msg{KeyPressMsg{typ: KeyLeft}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\x0d'},
			[]Msg{KeyPressMsg{typ: KeyEnter, mod: ModAlt}},
		},
		{
			"alt+backspace",
			[]byte{'\x1b', '\x7f'},
			[]Msg{KeyPressMsg{typ: KeyBackspace, mod: ModAlt}},
		},
		{
			"ctrl+space",
			[]byte{'\x00'},
			[]Msg{KeyPressMsg{typ: KeySpace, runes: []rune{' '}, mod: ModCtrl}},
		},
		{
			"ctrl+alt+space",
			[]byte{'\x1b', '\x00'},
			[]Msg{KeyPressMsg{typ: KeySpace, runes: []rune{' '}, mod: ModCtrl | ModAlt}},
		},
		{
			"esc",
			[]byte{'\x1b'},
			[]Msg{KeyPressMsg{typ: KeyEscape}},
		},
		{
			"alt+esc",
			[]byte{'\x1b', '\x1b'},
			[]Msg{KeyPressMsg{typ: KeyEscape, mod: ModAlt}},
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
				KeyPressMsg{runes: []rune{'o'}},
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
				KeyPressMsg{runes: []rune{'a'}},
				UnknownMsg(rune(0xfe)),
				KeyPressMsg{typ: KeySpace, runes: []rune{' '}},
				KeyPressMsg{runes: []rune{'b'}},
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
