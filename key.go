package tea

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

// KeyMsg contains information about a keypress. KeyMsgs are always sent to
// the program's update function. There are a couple general patterns you could
// use to check for keypresses:
//
//	// Switch on the string representation of the key (shorter)
//	switch msg := msg.(type) {
//	case KeyMsg:
//	    switch msg.String() {
//	    case "enter":
//	        fmt.Println("you pressed enter!")
//	    case "a":
//	        fmt.Println("you pressed a!")
//	    }
//	}
//
//	// Switch on the key type (more foolproof)
//	switch msg := msg.(type) {
//	case KeyMsg:
//	    switch msg.Type {
//	    case KeyEnter:
//	        fmt.Println("you pressed enter!")
//	    case KeyRunes:
//	        switch string(msg.Runes) {
//	        case "a":
//	            fmt.Println("you pressed a!")
//	        }
//	    }
//	}
//
// Note that Key.Runes will always contain at least one character, so you can
// always safely call Key.Runes[0]. In most cases Key.Runes will only contain
// one character, though certain input method editors (most notably Chinese
// IMEs) can input multiple runes at once.
type KeyMsg Key

// String returns a string representation for a key message. It's safe (and
// encouraged) for use in key comparison.
func (k KeyMsg) String() (str string) {
	return Key(k).String()
}

// Key contains information about a keypress.
type Key struct {
	Type  KeyType
	Runes []rune
	Alt   bool
	Paste bool
}

// String returns a friendly string representation for a key. It's safe (and
// encouraged) for use in key comparison.
//
//	k := Key{Type: KeyEnter}
//	fmt.Println(k)
//	// Output: enter
func (k Key) String() (str string) {
	var buf strings.Builder
	if k.Alt {
		buf.WriteString("alt+")
	}
	if k.Type == KeyRunes {
		if k.Paste {
			// Note: bubbles/keys bindings currently do string compares to
			// recognize shortcuts. Since pasted text should never activate
			// shortcuts, we need to ensure that the binding code doesn't
			// match Key events that result from pastes. We achieve this
			// here by enclosing pastes in '[...]' so that the string
			// comparison in Matches() fails in that case.
			buf.WriteByte('[')
		}
		buf.WriteString(string(k.Runes))
		if k.Paste {
			buf.WriteByte(']')
		}
		return buf.String()
	} else if s, ok := keyNames[k.Type]; ok {
		buf.WriteString(s)
		return buf.String()
	}
	return ""
}

// KeyType indicates the key pressed, such as KeyEnter or KeyBreak or KeyCtrlC.
// All other keys will be type KeyRunes. To get the rune value, check the Rune
// method on a Key struct, or use the Key.String() method:
//
//	k := Key{Type: KeyRunes, Runes: []rune{'a'}, Alt: true}
//	if k.Type == KeyRunes {
//
//	    fmt.Println(k.Runes)
//	    // Output: a
//
//	    fmt.Println(k.String())
//	    // Output: alt+a
//
//	}
type KeyType int

func (k KeyType) String() (str string) {
	if s, ok := keyNames[k]; ok {
		return s
	}
	return ""
}

// Control keys. We could do this with an iota, but the values are very
// specific, so we set the values explicitly to avoid any confusion.
//
// See also:
// https://en.wikipedia.org/wiki/C0_and_C1_control_codes
const (
	keyNUL KeyType = 0   // null, \0
	keySOH KeyType = 1   // start of heading
	keySTX KeyType = 2   // start of text
	keyETX KeyType = 3   // break, ctrl+c
	keyEOT KeyType = 4   // end of transmission
	keyENQ KeyType = 5   // enquiry
	keyACK KeyType = 6   // acknowledge
	keyBEL KeyType = 7   // bell, \a
	keyBS  KeyType = 8   // backspace
	keyHT  KeyType = 9   // horizontal tabulation, \t
	keyLF  KeyType = 10  // line feed, \n
	keyVT  KeyType = 11  // vertical tabulation \v
	keyFF  KeyType = 12  // form feed \f
	keyCR  KeyType = 13  // carriage return, \r
	keySO  KeyType = 14  // shift out
	keySI  KeyType = 15  // shift in
	keyDLE KeyType = 16  // data link escape
	keyDC1 KeyType = 17  // device control one
	keyDC2 KeyType = 18  // device control two
	keyDC3 KeyType = 19  // device control three
	keyDC4 KeyType = 20  // device control four
	keyNAK KeyType = 21  // negative acknowledge
	keySYN KeyType = 22  // synchronous idle
	keyETB KeyType = 23  // end of transmission block
	keyCAN KeyType = 24  // cancel
	keyEM  KeyType = 25  // end of medium
	keySUB KeyType = 26  // substitution
	keyESC KeyType = 27  // escape, \e
	keyFS  KeyType = 28  // file separator
	keyGS  KeyType = 29  // group separator
	keyRS  KeyType = 30  // record separator
	keyUS  KeyType = 31  // unit separator
	keyDEL KeyType = 127 // delete. on most systems this is mapped to backspace, I hear
)

// Control key aliases.
const (
	KeyNull      KeyType = keyNUL
	KeyBreak     KeyType = keyETX
	KeyEnter     KeyType = keyCR
	KeyBackspace KeyType = keyDEL
	KeyTab       KeyType = keyHT
	KeyEsc       KeyType = keyESC
	KeyEscape    KeyType = keyESC

	KeyCtrlAt           KeyType = keyNUL // ctrl+@
	KeyCtrlA            KeyType = keySOH
	KeyCtrlB            KeyType = keySTX
	KeyCtrlC            KeyType = keyETX
	KeyCtrlD            KeyType = keyEOT
	KeyCtrlE            KeyType = keyENQ
	KeyCtrlF            KeyType = keyACK
	KeyCtrlG            KeyType = keyBEL
	KeyCtrlH            KeyType = keyBS
	KeyCtrlI            KeyType = keyHT
	KeyCtrlJ            KeyType = keyLF
	KeyCtrlK            KeyType = keyVT
	KeyCtrlL            KeyType = keyFF
	KeyCtrlM            KeyType = keyCR
	KeyCtrlN            KeyType = keySO
	KeyCtrlO            KeyType = keySI
	KeyCtrlP            KeyType = keyDLE
	KeyCtrlQ            KeyType = keyDC1
	KeyCtrlR            KeyType = keyDC2
	KeyCtrlS            KeyType = keyDC3
	KeyCtrlT            KeyType = keyDC4
	KeyCtrlU            KeyType = keyNAK
	KeyCtrlV            KeyType = keySYN
	KeyCtrlW            KeyType = keyETB
	KeyCtrlX            KeyType = keyCAN
	KeyCtrlY            KeyType = keyEM
	KeyCtrlZ            KeyType = keySUB
	KeyCtrlOpenBracket  KeyType = keyESC // ctrl+[
	KeyCtrlBackslash    KeyType = keyFS  // ctrl+\
	KeyCtrlCloseBracket KeyType = keyGS  // ctrl+]
	KeyCtrlCaret        KeyType = keyRS  // ctrl+^
	KeyCtrlUnderscore   KeyType = keyUS  // ctrl+_
	KeyCtrlQuestionMark KeyType = keyDEL // ctrl+?
)

// Other keys.
const (
	KeyRunes KeyType = -(iota + 1)
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyShiftTab
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDown
	KeyCtrlPgUp
	KeyCtrlPgDown
	KeyDelete
	KeyInsert
	KeySpace
	KeyCtrlUp
	KeyCtrlDown
	KeyCtrlRight
	KeyCtrlLeft
	KeyCtrlHome
	KeyCtrlEnd
	KeyShiftUp
	KeyShiftDown
	KeyShiftRight
	KeyShiftLeft
	KeyShiftHome
	KeyShiftEnd
	KeyCtrlShiftUp
	KeyCtrlShiftDown
	KeyCtrlShiftLeft
	KeyCtrlShiftRight
	KeyCtrlShiftHome
	KeyCtrlShiftEnd
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
)

// Mappings for control keys and other special keys to friendly consts.
var keyNames = map[KeyType]string{
	// Control keys.
	keyNUL: "ctrl+@", // also ctrl+` (that's ctrl+backtick)
	keySOH: "ctrl+a",
	keySTX: "ctrl+b",
	keyETX: "ctrl+c",
	keyEOT: "ctrl+d",
	keyENQ: "ctrl+e",
	keyACK: "ctrl+f",
	keyBEL: "ctrl+g",
	keyBS:  "ctrl+h",
	keyHT:  "tab", // also ctrl+i
	keyLF:  "ctrl+j",
	keyVT:  "ctrl+k",
	keyFF:  "ctrl+l",
	keyCR:  "enter",
	keySO:  "ctrl+n",
	keySI:  "ctrl+o",
	keyDLE: "ctrl+p",
	keyDC1: "ctrl+q",
	keyDC2: "ctrl+r",
	keyDC3: "ctrl+s",
	keyDC4: "ctrl+t",
	keyNAK: "ctrl+u",
	keySYN: "ctrl+v",
	keyETB: "ctrl+w",
	keyCAN: "ctrl+x",
	keyEM:  "ctrl+y",
	keySUB: "ctrl+z",
	keyESC: "esc",
	keyFS:  "ctrl+\\",
	keyGS:  "ctrl+]",
	keyRS:  "ctrl+^",
	keyUS:  "ctrl+_",
	keyDEL: "backspace",

	// Other keys.
	KeyRunes:          "runes",
	KeyUp:             "up",
	KeyDown:           "down",
	KeyRight:          "right",
	KeySpace:          " ", // for backwards compatibility
	KeyLeft:           "left",
	KeyShiftTab:       "shift+tab",
	KeyHome:           "home",
	KeyEnd:            "end",
	KeyCtrlHome:       "ctrl+home",
	KeyCtrlEnd:        "ctrl+end",
	KeyShiftHome:      "shift+home",
	KeyShiftEnd:       "shift+end",
	KeyCtrlShiftHome:  "ctrl+shift+home",
	KeyCtrlShiftEnd:   "ctrl+shift+end",
	KeyPgUp:           "pgup",
	KeyPgDown:         "pgdown",
	KeyCtrlPgUp:       "ctrl+pgup",
	KeyCtrlPgDown:     "ctrl+pgdown",
	KeyDelete:         "delete",
	KeyInsert:         "insert",
	KeyCtrlUp:         "ctrl+up",
	KeyCtrlDown:       "ctrl+down",
	KeyCtrlRight:      "ctrl+right",
	KeyCtrlLeft:       "ctrl+left",
	KeyShiftUp:        "shift+up",
	KeyShiftDown:      "shift+down",
	KeyShiftRight:     "shift+right",
	KeyShiftLeft:      "shift+left",
	KeyCtrlShiftUp:    "ctrl+shift+up",
	KeyCtrlShiftDown:  "ctrl+shift+down",
	KeyCtrlShiftLeft:  "ctrl+shift+left",
	KeyCtrlShiftRight: "ctrl+shift+right",
	KeyF1:             "f1",
	KeyF2:             "f2",
	KeyF3:             "f3",
	KeyF4:             "f4",
	KeyF5:             "f5",
	KeyF6:             "f6",
	KeyF7:             "f7",
	KeyF8:             "f8",
	KeyF9:             "f9",
	KeyF10:            "f10",
	KeyF11:            "f11",
	KeyF12:            "f12",
	KeyF13:            "f13",
	KeyF14:            "f14",
	KeyF15:            "f15",
	KeyF16:            "f16",
	KeyF17:            "f17",
	KeyF18:            "f18",
	KeyF19:            "f19",
	KeyF20:            "f20",
}

// Sequence mappings.
var sequences = map[string]Key{
	// Arrow keys
	"\x1b[A":    {Type: KeyUp},
	"\x1b[B":    {Type: KeyDown},
	"\x1b[C":    {Type: KeyRight},
	"\x1b[D":    {Type: KeyLeft},
	"\x1b[1;2A": {Type: KeyShiftUp},
	"\x1b[1;2B": {Type: KeyShiftDown},
	"\x1b[1;2C": {Type: KeyShiftRight},
	"\x1b[1;2D": {Type: KeyShiftLeft},
	"\x1b[OA":   {Type: KeyShiftUp},    // DECCKM
	"\x1b[OB":   {Type: KeyShiftDown},  // DECCKM
	"\x1b[OC":   {Type: KeyShiftRight}, // DECCKM
	"\x1b[OD":   {Type: KeyShiftLeft},  // DECCKM
	"\x1b[a":    {Type: KeyShiftUp},    // urxvt
	"\x1b[b":    {Type: KeyShiftDown},  // urxvt
	"\x1b[c":    {Type: KeyShiftRight}, // urxvt
	"\x1b[d":    {Type: KeyShiftLeft},  // urxvt
	"\x1b[1;3A": {Type: KeyUp, Alt: true},
	"\x1b[1;3B": {Type: KeyDown, Alt: true},
	"\x1b[1;3C": {Type: KeyRight, Alt: true},
	"\x1b[1;3D": {Type: KeyLeft, Alt: true},

	"\x1b[1;4A": {Type: KeyShiftUp, Alt: true},
	"\x1b[1;4B": {Type: KeyShiftDown, Alt: true},
	"\x1b[1;4C": {Type: KeyShiftRight, Alt: true},
	"\x1b[1;4D": {Type: KeyShiftLeft, Alt: true},

	"\x1b[1;5A": {Type: KeyCtrlUp},
	"\x1b[1;5B": {Type: KeyCtrlDown},
	"\x1b[1;5C": {Type: KeyCtrlRight},
	"\x1b[1;5D": {Type: KeyCtrlLeft},
	"\x1b[Oa":   {Type: KeyCtrlUp, Alt: true},    // urxvt
	"\x1b[Ob":   {Type: KeyCtrlDown, Alt: true},  // urxvt
	"\x1b[Oc":   {Type: KeyCtrlRight, Alt: true}, // urxvt
	"\x1b[Od":   {Type: KeyCtrlLeft, Alt: true},  // urxvt
	"\x1b[1;6A": {Type: KeyCtrlShiftUp},
	"\x1b[1;6B": {Type: KeyCtrlShiftDown},
	"\x1b[1;6C": {Type: KeyCtrlShiftRight},
	"\x1b[1;6D": {Type: KeyCtrlShiftLeft},
	"\x1b[1;7A": {Type: KeyCtrlUp, Alt: true},
	"\x1b[1;7B": {Type: KeyCtrlDown, Alt: true},
	"\x1b[1;7C": {Type: KeyCtrlRight, Alt: true},
	"\x1b[1;7D": {Type: KeyCtrlLeft, Alt: true},
	"\x1b[1;8A": {Type: KeyCtrlShiftUp, Alt: true},
	"\x1b[1;8B": {Type: KeyCtrlShiftDown, Alt: true},
	"\x1b[1;8C": {Type: KeyCtrlShiftRight, Alt: true},
	"\x1b[1;8D": {Type: KeyCtrlShiftLeft, Alt: true},

	// Miscellaneous keys
	"\x1b[Z": {Type: KeyShiftTab},

	"\x1b[2~":   {Type: KeyInsert},
	"\x1b[3;2~": {Type: KeyInsert, Alt: true},

	"\x1b[3~":   {Type: KeyDelete},
	"\x1b[3;3~": {Type: KeyDelete, Alt: true},

	"\x1b[5~":   {Type: KeyPgUp},
	"\x1b[5;3~": {Type: KeyPgUp, Alt: true},
	"\x1b[5;5~": {Type: KeyCtrlPgUp},
	"\x1b[5^":   {Type: KeyCtrlPgUp}, // urxvt
	"\x1b[5;7~": {Type: KeyCtrlPgUp, Alt: true},

	"\x1b[6~":   {Type: KeyPgDown},
	"\x1b[6;3~": {Type: KeyPgDown, Alt: true},
	"\x1b[6;5~": {Type: KeyCtrlPgDown},
	"\x1b[6^":   {Type: KeyCtrlPgDown}, // urxvt
	"\x1b[6;7~": {Type: KeyCtrlPgDown, Alt: true},

	"\x1b[1~":   {Type: KeyHome},
	"\x1b[H":    {Type: KeyHome},                     // xterm, lxterm
	"\x1b[1;3H": {Type: KeyHome, Alt: true},          // xterm, lxterm
	"\x1b[1;5H": {Type: KeyCtrlHome},                 // xterm, lxterm
	"\x1b[1;7H": {Type: KeyCtrlHome, Alt: true},      // xterm, lxterm
	"\x1b[1;2H": {Type: KeyShiftHome},                // xterm, lxterm
	"\x1b[1;4H": {Type: KeyShiftHome, Alt: true},     // xterm, lxterm
	"\x1b[1;6H": {Type: KeyCtrlShiftHome},            // xterm, lxterm
	"\x1b[1;8H": {Type: KeyCtrlShiftHome, Alt: true}, // xterm, lxterm

	"\x1b[4~":   {Type: KeyEnd},
	"\x1b[F":    {Type: KeyEnd},                     // xterm, lxterm
	"\x1b[1;3F": {Type: KeyEnd, Alt: true},          // xterm, lxterm
	"\x1b[1;5F": {Type: KeyCtrlEnd},                 // xterm, lxterm
	"\x1b[1;7F": {Type: KeyCtrlEnd, Alt: true},      // xterm, lxterm
	"\x1b[1;2F": {Type: KeyShiftEnd},                // xterm, lxterm
	"\x1b[1;4F": {Type: KeyShiftEnd, Alt: true},     // xterm, lxterm
	"\x1b[1;6F": {Type: KeyCtrlShiftEnd},            // xterm, lxterm
	"\x1b[1;8F": {Type: KeyCtrlShiftEnd, Alt: true}, // xterm, lxterm

	"\x1b[7~": {Type: KeyHome},          // urxvt
	"\x1b[7^": {Type: KeyCtrlHome},      // urxvt
	"\x1b[7$": {Type: KeyShiftHome},     // urxvt
	"\x1b[7@": {Type: KeyCtrlShiftHome}, // urxvt

	"\x1b[8~": {Type: KeyEnd},          // urxvt
	"\x1b[8^": {Type: KeyCtrlEnd},      // urxvt
	"\x1b[8$": {Type: KeyShiftEnd},     // urxvt
	"\x1b[8@": {Type: KeyCtrlShiftEnd}, // urxvt

	// Function keys, Linux console
	"\x1b[[A": {Type: KeyF1}, // linux console
	"\x1b[[B": {Type: KeyF2}, // linux console
	"\x1b[[C": {Type: KeyF3}, // linux console
	"\x1b[[D": {Type: KeyF4}, // linux console
	"\x1b[[E": {Type: KeyF5}, // linux console

	// Function keys, X11
	"\x1bOP": {Type: KeyF1}, // vt100, xterm
	"\x1bOQ": {Type: KeyF2}, // vt100, xterm
	"\x1bOR": {Type: KeyF3}, // vt100, xterm
	"\x1bOS": {Type: KeyF4}, // vt100, xterm

	"\x1b[1;3P": {Type: KeyF1, Alt: true}, // vt100, xterm
	"\x1b[1;3Q": {Type: KeyF2, Alt: true}, // vt100, xterm
	"\x1b[1;3R": {Type: KeyF3, Alt: true}, // vt100, xterm
	"\x1b[1;3S": {Type: KeyF4, Alt: true}, // vt100, xterm

	"\x1b[11~": {Type: KeyF1}, // urxvt
	"\x1b[12~": {Type: KeyF2}, // urxvt
	"\x1b[13~": {Type: KeyF3}, // urxvt
	"\x1b[14~": {Type: KeyF4}, // urxvt

	"\x1b[15~": {Type: KeyF5}, // vt100, xterm, also urxvt

	"\x1b[15;3~": {Type: KeyF5, Alt: true}, // vt100, xterm, also urxvt

	"\x1b[17~": {Type: KeyF6},  // vt100, xterm, also urxvt
	"\x1b[18~": {Type: KeyF7},  // vt100, xterm, also urxvt
	"\x1b[19~": {Type: KeyF8},  // vt100, xterm, also urxvt
	"\x1b[20~": {Type: KeyF9},  // vt100, xterm, also urxvt
	"\x1b[21~": {Type: KeyF10}, // vt100, xterm, also urxvt

	"\x1b[17;3~": {Type: KeyF6, Alt: true},  // vt100, xterm
	"\x1b[18;3~": {Type: KeyF7, Alt: true},  // vt100, xterm
	"\x1b[19;3~": {Type: KeyF8, Alt: true},  // vt100, xterm
	"\x1b[20;3~": {Type: KeyF9, Alt: true},  // vt100, xterm
	"\x1b[21;3~": {Type: KeyF10, Alt: true}, // vt100, xterm

	"\x1b[23~": {Type: KeyF11}, // vt100, xterm, also urxvt
	"\x1b[24~": {Type: KeyF12}, // vt100, xterm, also urxvt

	"\x1b[23;3~": {Type: KeyF11, Alt: true}, // vt100, xterm
	"\x1b[24;3~": {Type: KeyF12, Alt: true}, // vt100, xterm

	"\x1b[1;2P": {Type: KeyF13},
	"\x1b[1;2Q": {Type: KeyF14},

	"\x1b[25~": {Type: KeyF13}, // vt100, xterm, also urxvt
	"\x1b[26~": {Type: KeyF14}, // vt100, xterm, also urxvt

	"\x1b[25;3~": {Type: KeyF13, Alt: true}, // vt100, xterm
	"\x1b[26;3~": {Type: KeyF14, Alt: true}, // vt100, xterm

	"\x1b[1;2R": {Type: KeyF15},
	"\x1b[1;2S": {Type: KeyF16},

	"\x1b[28~": {Type: KeyF15}, // vt100, xterm, also urxvt
	"\x1b[29~": {Type: KeyF16}, // vt100, xterm, also urxvt

	"\x1b[28;3~": {Type: KeyF15, Alt: true}, // vt100, xterm
	"\x1b[29;3~": {Type: KeyF16, Alt: true}, // vt100, xterm

	"\x1b[15;2~": {Type: KeyF17},
	"\x1b[17;2~": {Type: KeyF18},
	"\x1b[18;2~": {Type: KeyF19},
	"\x1b[19;2~": {Type: KeyF20},

	"\x1b[31~": {Type: KeyF17},
	"\x1b[32~": {Type: KeyF18},
	"\x1b[33~": {Type: KeyF19},
	"\x1b[34~": {Type: KeyF20},

	// Powershell sequences.
	"\x1bOA": {Type: KeyUp, Alt: false},
	"\x1bOB": {Type: KeyDown, Alt: false},
	"\x1bOC": {Type: KeyRight, Alt: false},
	"\x1bOD": {Type: KeyLeft, Alt: false},
}

// unknownInputByteMsg is reported by the input reader when an invalid
// utf-8 byte is detected on the input. Currently, it is not handled
// further by bubbletea. However, having this event makes it possible
// to troubleshoot invalid inputs.
type unknownInputByteMsg byte

func (u unknownInputByteMsg) String() string {
	return fmt.Sprintf("?%#02x?", int(u))
}

// unknownCSISequenceMsg is reported by the input reader when an
// unrecognized CSI sequence is detected on the input. Currently, it
// is not handled further by bubbletea. However, having this event
// makes it possible to troubleshoot invalid inputs.
type unknownCSISequenceMsg []byte

func (u unknownCSISequenceMsg) String() string {
	return fmt.Sprintf("?CSI%+v?", []byte(u)[2:])
}

var spaceRunes = []rune{' '}

// readAnsiInputs reads keypress and mouse inputs from a TTY and produces messages
// containing information about the key or mouse events accordingly.
func readAnsiInputs(ctx context.Context, msgs chan<- Msg, input io.Reader) error {
	var buf [256]byte

	var leftOverFromPrevIteration []byte
loop:
	for {
		// Read and block.
		numBytes, err := input.Read(buf[:])
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
		b := buf[:numBytes]
		if leftOverFromPrevIteration != nil {
			b = append(leftOverFromPrevIteration, b...)
		}

		// If we had a short read (numBytes < len(buf)), we're sure that
		// the end of this read is an event boundary, so there is no doubt
		// if we are encountering the end of the buffer while parsing a message.
		// However, if we've succeeded in filling up the buffer, there may
		// be more data in the OS buffer ready to be read in, to complete
		// the last message in the input. In that case, we will retry with
		// the left over data in the next iteration.
		canHaveMoreData := numBytes == len(buf)

		var i, w int
		for i, w = 0, 0; i < len(b); i += w {
			var msg Msg
			w, msg = detectOneMsg(b[i:], canHaveMoreData)
			if w == 0 {
				// Expecting more bytes beyond the current buffer. Try waiting
				// for more input.
				leftOverFromPrevIteration = make([]byte, 0, len(b[i:])+len(buf))
				leftOverFromPrevIteration = append(leftOverFromPrevIteration, b[i:]...)
				continue loop
			}

			select {
			case msgs <- msg:
			case <-ctx.Done():
				err := ctx.Err()
				if err != nil {
					err = fmt.Errorf("found context error while reading input: %w", err)
				}
				return err
			}
		}
		leftOverFromPrevIteration = nil
	}
}

var (
	unknownCSIRe  = regexp.MustCompile(`^\x1b\[[\x30-\x3f]*[\x20-\x2f]*[\x40-\x7e]`)
	mouseSGRRegex = regexp.MustCompile(`(\d+);(\d+);(\d+)([Mm])`)
)

func detectOneMsg(b []byte, canHaveMoreData bool) (w int, msg Msg) {
	// Detect mouse events.
	// X10 mouse events have a length of 6 bytes
	const mouseEventX10Len = 6
	if len(b) >= mouseEventX10Len && b[0] == '\x1b' && b[1] == '[' {
		switch b[2] {
		case 'M':
			return mouseEventX10Len, MouseMsg(parseX10MouseEvent(b))
		case '<':
			if matchIndices := mouseSGRRegex.FindSubmatchIndex(b[3:]); matchIndices != nil {
				// SGR mouse events length is the length of the match plus the length of the escape sequence
				mouseEventSGRLen := matchIndices[1] + 3 //nolint:gomnd
				return mouseEventSGRLen, MouseMsg(parseSGRMouseEvent(b))
			}
		}
	}

	// Detect focus events.
	var foundRF bool
	foundRF, w, msg = detectReportFocus(b)
	if foundRF {
		return w, msg
	}

	// Detect bracketed paste.
	var foundbp bool
	foundbp, w, msg = detectBracketedPaste(b)
	if foundbp {
		return w, msg
	}

	// Detect escape sequence and control characters other than NUL,
	// possibly with an escape character in front to mark the Alt
	// modifier.
	var foundSeq bool
	foundSeq, w, msg = detectSequence(b)
	if foundSeq {
		return w, msg
	}

	// No non-NUL control character or escape sequence.
	// If we are seeing at least an escape character, remember it for later below.
	alt := false
	i := 0
	if b[0] == '\x1b' {
		alt = true
		i++
	}

	// Are we seeing a standalone NUL? This is not handled by detectSequence().
	if i < len(b) && b[i] == 0 {
		return i + 1, KeyMsg{Type: keyNUL, Alt: alt}
	}

	// Find the longest sequence of runes that are not control
	// characters from this point.
	var runes []rune
	for rw := 0; i < len(b); i += rw {
		var r rune
		r, rw = utf8.DecodeRune(b[i:])
		if r == utf8.RuneError || r <= rune(keyUS) || r == rune(keyDEL) || r == ' ' {
			// Rune errors are handled below; control characters and spaces will
			// be handled by detectSequence in the next call to detectOneMsg.
			break
		}
		runes = append(runes, r)
		if alt {
			// We only support a single rune after an escape alt modifier.
			i += rw
			break
		}
	}
	if i >= len(b) && canHaveMoreData {
		// We have encountered the end of the input buffer. Alas, we can't
		// be sure whether the data in the remainder of the buffer is
		// complete (maybe there was a short read). Instead of sending anything
		// dumb to the message channel, do a short read. The outer loop will
		// handle this case by extending the buffer as necessary.
		return 0, nil
	}

	// If we found at least one rune, we report the bunch of them as
	// a single KeyRunes or KeySpace event.
	if len(runes) > 0 {
		k := Key{Type: KeyRunes, Runes: runes, Alt: alt}
		if len(runes) == 1 && runes[0] == ' ' {
			k.Type = KeySpace
		}
		return i, KeyMsg(k)
	}

	// We didn't find an escape sequence, nor a valid rune. Was this a
	// lone escape character at the end of the input?
	if alt && len(b) == 1 {
		return 1, KeyMsg(Key{Type: KeyEscape})
	}

	// The character at the current position is neither an escape
	// sequence, a valid rune start or a sole escape character. Report
	// it as an invalid byte.
	return 1, unknownInputByteMsg(b[0])
}
