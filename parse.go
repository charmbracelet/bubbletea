package tea

import (
	"encoding/base64"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// Flags to control the behavior of the parser.
// TODO: Should these be exported?
const (
	// When this flag is set, the driver will treat both Ctrl+Space and Ctrl+@
	// as the same key sequence.
	//
	// Historically, the ANSI specs generate NUL (0x00) on both the Ctrl+Space
	// and Ctrl+@ key sequences. This flag allows the driver to treat both as
	// the same key sequence.
	_FlagCtrlAt = 1 << iota

	// When this flag is set, the driver will treat the Tab key and Ctrl+I as
	// the same key sequence.
	//
	// Historically, the ANSI specs generate HT (0x09) on both the Tab key and
	// Ctrl+I. This flag allows the driver to treat both as the same key
	// sequence.
	_FlagCtrlI

	// When this flag is set, the driver will treat the Enter key and Ctrl+M as
	// the same key sequence.
	//
	// Historically, the ANSI specs generate CR (0x0D) on both the Enter key
	// and Ctrl+M. This flag allows the driver to treat both as the same key
	_FlagCtrlM

	// When this flag is set, the driver will treat Escape and Ctrl+[ as
	// the same key sequence.
	//
	// Historically, the ANSI specs generate ESC (0x1B) on both the Escape key
	// and Ctrl+[. This flag allows the driver to treat both as the same key
	// sequence.
	_FlagCtrlOpenBracket

	// When this flag is set, the driver will send a BS (0x08 byte) character
	// instead of a DEL (0x7F byte) character when the Backspace key is
	// pressed.
	//
	// The VT100 terminal has both a Backspace and a Delete key. The VT220
	// terminal dropped the Backspace key and replaced it with the Delete key.
	// Both terminals send a DEL character when the Delete key is pressed.
	// Modern terminals and PCs later readded the Delete key but used a
	// different key sequence, and the Backspace key was standardized to send a
	// DEL character.
	_FlagBackspace

	// When this flag is set, the driver will recognize the Find key instead of
	// treating it as a Home key.
	//
	// The Find key was part of the VT220 keyboard, and is no longer used in
	// modern day PCs.
	_FlagFind

	// When this flag is set, the driver will recognize the Select key instead
	// of treating it as a End key.
	//
	// The Symbol key was part of the VT220 keyboard, and is no longer used in
	// modern day PCs.
	_FlagSelect

	// When this flag is set, the driver will use Terminfo databases to
	// overwrite the default key sequences.
	_FlagTerminfo

	// When this flag is set, the driver will preserve function keys (F13-F63)
	// as symbols.
	//
	// Since these keys are not part of today's standard 20th century keyboard,
	// we treat them as F1-F12 modifier keys i.e. ctrl/shift/alt + Fn combos.
	// Key definitions come from Terminfo, this flag is only useful when
	// FlagTerminfo is not set.
	_FlagFKeys
)

// inputParser is a parser for input escape sequences.
// TODO: Use [ansi.DecodeSequence] instead of this parser.
type inputParser struct {
	flags int
}

// parseSequence finds the first recognized event sequence and returns it along
// with its length.
//
// It will return zero and nil no sequence is recognized or when the buffer is
// empty. If a sequence is not supported, an UnknownEvent is returned.
func (p *inputParser) parseSequence(buf []byte) (n int, msg Msg) {
	if len(buf) == 0 {
		return 0, nil
	}

	switch b := buf[0]; b {
	case ansi.ESC:
		if len(buf) == 1 {
			// Escape key
			return 1, KeyPressMsg{Code: KeyEscape}
		}

		switch bPrime := buf[1]; bPrime {
		case 'O': // Esc-prefixed SS3
			return p.parseSs3(buf)
		case 'P': // Esc-prefixed DCS
			return p.parseDcs(buf)
		case '[': // Esc-prefixed CSI
			return p.parseCsi(buf)
		case ']': // Esc-prefixed OSC
			return p.parseOsc(buf)
		case '_': // Esc-prefixed APC
			return p.parseApc(buf)
		default:
			n, e := p.parseSequence(buf[1:])
			if k, ok := e.(KeyPressMsg); ok {
				k.Text = ""
				k.Mod |= ModAlt
				return n + 1, k
			}

			// Not a key sequence, nor an alt modified key sequence. In that
			// case, just report a single escape key.
			return 1, KeyPressMsg{Code: KeyEscape}
		}
	case ansi.SS3:
		return p.parseSs3(buf)
	case ansi.DCS:
		return p.parseDcs(buf)
	case ansi.CSI:
		return p.parseCsi(buf)
	case ansi.OSC:
		return p.parseOsc(buf)
	case ansi.APC:
		return p.parseApc(buf)
	default:
		if b <= ansi.US || b == ansi.DEL || b == ansi.SP {
			return 1, p.parseControl(b)
		} else if b >= ansi.PAD && b <= ansi.APC {
			// C1 control code
			// UTF-8 never starts with a C1 control code
			// Encode these as Ctrl+Alt+<code - 0x40>
			code := rune(b) - 0x40
			return 1, KeyPressMsg{Code: code, Mod: ModCtrl | ModAlt}
		}
		return p.parseUtf8(buf)
	}
}

func (p *inputParser) parseCsi(b []byte) (int, Msg) {
	if len(b) == 2 && b[0] == ansi.ESC {
		// short cut if this is an alt+[ key
		return 2, KeyPressMsg{Text: string(rune(b[1])), Mod: ModAlt}
	}

	var csi ansi.CsiSequence
	var params [parser.MaxParamsSize]int
	var paramsLen int

	var i int
	if b[i] == ansi.CSI || b[i] == ansi.ESC {
		i++
	}
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == '[' {
		i++
	}

	// Initial CSI byte
	if i < len(b) && b[i] >= '<' && b[i] <= '?' {
		csi.Cmd |= int(b[i]) << parser.MarkerShift
	}

	// Scan parameter bytes in the range 0x30-0x3F
	var j int
	for j = 0; i < len(b) && paramsLen < len(params) && b[i] >= 0x30 && b[i] <= 0x3F; i, j = i+1, j+1 {
		if b[i] >= '0' && b[i] <= '9' {
			if params[paramsLen] == parser.MissingParam {
				params[paramsLen] = 0
			}
			params[paramsLen] *= 10
			params[paramsLen] += int(b[i]) - '0'
		}
		if b[i] == ':' {
			params[paramsLen] |= parser.HasMoreFlag
		}
		if b[i] == ';' || b[i] == ':' {
			paramsLen++
			if paramsLen < len(params) {
				// Don't overflow the params slice
				params[paramsLen] = parser.MissingParam
			}
		}
	}

	if j > 0 && paramsLen < len(params) {
		// has parameters
		paramsLen++
	}

	// Scan intermediate bytes in the range 0x20-0x2F
	var intermed byte
	for ; i < len(b) && b[i] >= 0x20 && b[i] <= 0x2F; i++ {
		intermed = b[i]
	}

	// Set the intermediate byte
	csi.Cmd |= int(intermed) << parser.IntermedShift

	// Scan final byte in the range 0x40-0x7E
	if i >= len(b) || b[i] < 0x40 || b[i] > 0x7E {
		// Special case for URxvt keys
		// CSI <number> $ is an invalid sequence, but URxvt uses it for
		// shift modified keys.
		if b[i-1] == '$' {
			n, ev := p.parseCsi(append(b[:i-1], '~'))
			if k, ok := ev.(KeyPressMsg); ok {
				k.Mod |= ModShift
				return n, k
			}
		}
		return i, UnknownMsg(b[:i-1])
	}

	// Add the final byte
	csi.Cmd |= int(b[i])
	i++

	csi.Params = params[:paramsLen]
	switch cmd := csi.Cmd; cmd {
	case 'y' | '?'<<parser.MarkerShift | '$'<<parser.IntermedShift:
		// Report Mode (DECRPM)
		if paramsLen == 2 && csi.Param(0) != -1 && csi.Param(1) != -1 {
			return i, modeReportMsg{Mode: csi.Param(0), Value: csi.Param(1)}
		}
	case 'c' | '?'<<parser.MarkerShift:
		// Primary Device Attributes
		return i, parsePrimaryDevAttrs(&csi)
	case 'u' | '?'<<parser.MarkerShift:
		// Kitty keyboard flags
		if param := csi.Param(0); param != -1 {
			return i, KeyboardEnhancementsMsg{kittyFlags: param}
		}
	case 'R' | '?'<<parser.MarkerShift:
		// This report may return a third parameter representing the page
		// number, but we don't really need it.
		if paramsLen >= 2 && csi.Param(0) != -1 && csi.Param(1) != -1 {
			return i, CursorPositionMsg{Y: csi.Param(0) - 1, X: csi.Param(1) - 1}
		}
	case 'm' | '<'<<parser.MarkerShift, 'M' | '<'<<parser.MarkerShift:
		// Handle SGR mouse
		if paramsLen == 3 {
			return i, parseSGRMouseEvent(&csi)
		}
	case 'm' | '>'<<parser.MarkerShift:
		// XTerm modifyOtherKeys
		if paramsLen == 2 && csi.Param(0) == 4 && csi.Param(1) != -1 {
			return i, KeyboardEnhancementsMsg{modifyOtherKeys: csi.Param(1)}
		}
	case 'I':
		return i, FocusMsg{}
	case 'O':
		return i, BlurMsg{}
	case 'R':
		// Cursor position report OR modified F3
		if paramsLen == 2 && csi.Param(0) != -1 && csi.Param(1) != -1 {
			m := CursorPositionMsg{Y: csi.Param(0) - 1, X: csi.Param(1) - 1}
			if csi.Param(0) == 1 && csi.Param(1)-1 <= int(ModMeta|ModShift|ModAlt|ModCtrl) {
				// XXX: We cannot differentiate between cursor position report and
				// CSI 1 ; <mod> R (which is modified F3) when the cursor is at the
				// row 1. In this case, we report both messages.
				//
				// For a non ambiguous cursor position report, use
				// [ansi.RequestExtendedCursorPosition] (DECXCPR) instead.
				return i, multiMsg{KeyPressMsg{Code: KeyF3, Mod: KeyMod(csi.Param(1) - 1)}, m}
			}

			return i, m
		}

		if paramsLen != 0 {
			break
		}

		// Unmodified key F3 (CSI R)
		fallthrough
	case 'a', 'b', 'c', 'd', 'A', 'B', 'C', 'D', 'E', 'F', 'H', 'P', 'Q', 'S', 'Z':
		var k KeyPressMsg
		switch cmd {
		case 'a', 'b', 'c', 'd':
			k = KeyPressMsg{Code: KeyUp + rune(cmd-'a'), Mod: ModShift}
		case 'A', 'B', 'C', 'D':
			k = KeyPressMsg{Code: KeyUp + rune(cmd-'A')}
		case 'E':
			k = KeyPressMsg{Code: KeyBegin}
		case 'F':
			k = KeyPressMsg{Code: KeyEnd}
		case 'H':
			k = KeyPressMsg{Code: KeyHome}
		case 'P', 'Q', 'R', 'S':
			k = KeyPressMsg{Code: KeyF1 + rune(cmd-'P')}
		case 'Z':
			k = KeyPressMsg{Code: KeyTab, Mod: ModShift}
		}
		if paramsLen > 1 && csi.Param(0) == 1 && csi.Param(1) != -1 {
			// CSI 1 ; <modifiers> A
			k.Mod |= KeyMod(csi.Param(1) - 1)
		}
		// Don't forget to handle Kitty keyboard protocol
		return i, parseKittyKeyboardExt(&csi, k)
	case 'M':
		// Handle X10 mouse
		if i+3 > len(b) {
			return i, UnknownMsg(b[:i])
		}
		return i + 3, parseX10MouseEvent(append(b[:i], b[i:i+3]...))
	case 'y':
		// Report Mode (DECRPM)
		if paramsLen != 2 && csi.Param(0) != -1 && csi.Param(0) != -1 {
			return i, UnknownMsg(b[:i])
		}
		return i, modeReportMsg{Mode: csi.Param(0), Value: csi.Param(1)}
	case 'u':
		// Kitty keyboard protocol & CSI u (fixterms)
		if paramsLen == 0 {
			return i, UnknownMsg(b[:i])
		}
		return i, parseKittyKeyboard(&csi)
	case '_':
		// Win32 Input Mode
		if paramsLen != 6 {
			return i, UnknownMsg(b[:i])
		}

		rc := uint16(csi.Param(5)) //nolint:gosec
		if rc == 0 {
			rc = 1
		}

		event := p.parseWin32InputKeyEvent(
			nil,
			uint16(csi.Param(0)), //nolint:gosec // Vk wVirtualKeyCode
			uint16(csi.Param(1)), //nolint:gosec // Sc wVirtualScanCode
			rune(csi.Param(2)),   // Uc UnicodeChar
			csi.Param(3) == 1,    // Kd bKeyDown
			uint32(csi.Param(4)), //nolint:gosec // Cs dwControlKeyState
			rc,                   // Rc wRepeatCount
		)

		if event == nil {
			return i, UnknownMsg(b[:])
		}

		return i, event
	case '@', '^', '~':
		if paramsLen == 0 {
			return i, UnknownMsg(b[:i])
		}

		param := csi.Param(0)
		switch cmd {
		case '~':
			switch param {
			case 27:
				// XTerm modifyOtherKeys 2
				if paramsLen != 3 {
					return i, UnknownMsg(b[:i])
				}
				return i, parseXTermModifyOtherKeys(&csi)
			case 200:
				// bracketed-paste start
				return i, PasteStartMsg{}
			case 201:
				// bracketed-paste end
				return i, PasteEndMsg{}
			}
		}

		switch param {
		case 1, 2, 3, 4, 5, 6, 7, 8,
			11, 12, 13, 14, 15,
			17, 18, 19, 20, 21,
			23, 24, 25, 26,
			28, 29, 31, 32, 33, 34:
			var k KeyPressMsg
			switch param {
			case 1:
				if p.flags&_FlagFind != 0 {
					k = KeyPressMsg{Code: KeyFind}
				} else {
					k = KeyPressMsg{Code: KeyHome}
				}
			case 2:
				k = KeyPressMsg{Code: KeyInsert}
			case 3:
				k = KeyPressMsg{Code: KeyDelete}
			case 4:
				if p.flags&_FlagSelect != 0 {
					k = KeyPressMsg{Code: KeySelect}
				} else {
					k = KeyPressMsg{Code: KeyEnd}
				}
			case 5:
				k = KeyPressMsg{Code: KeyPgUp}
			case 6:
				k = KeyPressMsg{Code: KeyPgDown}
			case 7:
				k = KeyPressMsg{Code: KeyHome}
			case 8:
				k = KeyPressMsg{Code: KeyEnd}
			case 11, 12, 13, 14, 15:
				k = KeyPressMsg{Code: KeyF1 + rune(param-11)}
			case 17, 18, 19, 20, 21:
				k = KeyPressMsg{Code: KeyF6 + rune(param-17)}
			case 23, 24, 25, 26:
				k = KeyPressMsg{Code: KeyF11 + rune(param-23)}
			case 28, 29:
				k = KeyPressMsg{Code: KeyF15 + rune(param-28)}
			case 31, 32, 33, 34:
				k = KeyPressMsg{Code: KeyF17 + rune(param-31)}
			}

			// modifiers
			if paramsLen > 1 && csi.Param(1) != -1 {
				k.Mod |= KeyMod(csi.Param(1) - 1)
			}

			// Handle URxvt weird keys
			switch cmd {
			case '~':
				// Don't forget to handle Kitty keyboard protocol
				return i, parseKittyKeyboardExt(&csi, k)
			case '^':
				k.Mod |= ModCtrl
			case '@':
				k.Mod |= ModCtrl | ModShift
			}

			return i, k
		}
	}
	return i, UnknownMsg(b[:i])
}

// parseSs3 parses a SS3 sequence.
// See https://vt100.net/docs/vt220-rm/chapter4.html#S4.4.4.2
func (p *inputParser) parseSs3(b []byte) (int, Msg) {
	if len(b) == 2 && b[0] == ansi.ESC {
		// short cut if this is an alt+O key
		return 2, KeyPressMsg{Code: rune(b[1]), Mod: ModAlt}
	}

	var i int
	if b[i] == ansi.SS3 || b[i] == ansi.ESC {
		i++
	}
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == 'O' {
		i++
	}

	// Scan numbers from 0-9
	var mod int
	for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
		mod *= 10
		mod += int(b[i]) - '0'
	}

	// Scan a GL character
	// A GL character is a single byte in the range 0x21-0x7E
	// See https://vt100.net/docs/vt220-rm/chapter2.html#S2.3.2
	if i >= len(b) || b[i] < 0x21 || b[i] > 0x7E {
		return i, UnknownMsg(b[:i])
	}

	// GL character(s)
	gl := b[i]
	i++

	var k KeyPressMsg
	switch gl {
	case 'a', 'b', 'c', 'd':
		k = KeyPressMsg{Code: KeyUp + rune(gl-'a'), Mod: ModCtrl}
	case 'A', 'B', 'C', 'D':
		k = KeyPressMsg{Code: KeyUp + rune(gl-'A')}
	case 'E':
		k = KeyPressMsg{Code: KeyBegin}
	case 'F':
		k = KeyPressMsg{Code: KeyEnd}
	case 'H':
		k = KeyPressMsg{Code: KeyHome}
	case 'P', 'Q', 'R', 'S':
		k = KeyPressMsg{Code: KeyF1 + rune(gl-'P')}
	case 'M':
		k = KeyPressMsg{Code: KeyKpEnter}
	case 'X':
		k = KeyPressMsg{Code: KeyKpEqual}
	case 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y':
		k = KeyPressMsg{Code: KeyKpMultiply + rune(gl-'j')}
	default:
		return i, UnknownMsg(b[:i])
	}

	// Handle weird SS3 <modifier> Func
	if mod > 0 {
		k.Mod |= KeyMod(mod - 1)
	}

	return i, k
}

func (p *inputParser) parseOsc(b []byte) (int, Msg) {
	if len(b) == 2 && b[0] == ansi.ESC {
		// short cut if this is an alt+] key
		return 2, KeyPressMsg{Code: rune(b[1]), Mod: ModAlt}
	}

	var i int
	if b[i] == ansi.OSC || b[i] == ansi.ESC {
		i++
	}
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == ']' {
		i++
	}

	// Parse OSC command
	// An OSC sequence is terminated by a BEL, ESC, or ST character
	var start, end int
	cmd := -1
	for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
		if cmd == -1 {
			cmd = 0
		} else {
			cmd *= 10
		}
		cmd += int(b[i]) - '0'
	}

	if i < len(b) && b[i] == ';' {
		// mark the start of the sequence data
		i++
		start = i
	}

	for ; i < len(b); i++ {
		// advance to the end of the sequence
		if b[i] == ansi.BEL || b[i] == ansi.ESC || b[i] == ansi.ST {
			break
		}
	}

	if i >= len(b) {
		return i, UnknownMsg(b[:i])
	}

	end = i // end of the sequence data
	i++

	// Check 7-bit ST (string terminator) character
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == '\\' {
		i++
	}

	if end <= start {
		return i, UnknownMsg(b[:i])
	}

	data := string(b[start:end])
	switch cmd {
	case 10:
		return i, ForegroundColorMsg{xParseColor(data)}
	case 11:
		return i, BackgroundColorMsg{xParseColor(data)}
	case 12:
		return i, CursorColorMsg{xParseColor(data)}
	case 52:
		parts := strings.Split(data, ";")
		if len(parts) == 0 {
			return i, ClipboardMsg("")
		}
		if len(parts) != 2 {
			break
		}

		b64 := parts[1]
		bts, err := base64.StdEncoding.DecodeString(b64)

		switch parts[0] {
		case "c":
			if err != nil {
				return i, ClipboardMsg("")
			}
			return i, ClipboardMsg(string(bts))
		case "p":
			if err != nil {
				return i, PrimaryClipboardMsg("")
			}
			return i, PrimaryClipboardMsg(string(bts))
		}
	}

	return i, UnknownMsg(b[:i])
}

// parseStTerminated parses a control sequence that gets terminated by a ST character.
func (p *inputParser) parseStTerminated(intro8, intro7 byte) func([]byte) (int, Msg) {
	return func(b []byte) (int, Msg) {
		var i int
		if b[i] == intro8 || b[i] == ansi.ESC {
			i++
		}
		if i < len(b) && b[i-1] == ansi.ESC && b[i] == intro7 {
			i++
		}

		// Scan control sequence
		// Most common control sequence is terminated by a ST character
		// ST is a 7-bit string terminator character is (ESC \)
		// nolint: revive
		for ; i < len(b) && b[i] != ansi.ST && b[i] != ansi.ESC; i++ {
		}

		if i >= len(b) {
			return i, UnknownMsg(b[:i])
		}
		i++

		// Check 7-bit ST (string terminator) character
		if i < len(b) && b[i-1] == ansi.ESC && b[i] == '\\' {
			i++
		}

		return i, UnknownMsg(b[:i])
	}
}

func (p *inputParser) parseDcs(b []byte) (int, Msg) {
	if len(b) == 2 && b[0] == ansi.ESC {
		// short cut if this is an alt+P key
		return 2, KeyPressMsg{Code: rune(b[1]), Mod: ModAlt}
	}

	var params [16]int
	var paramsLen int
	var dcs ansi.DcsSequence

	// DCS sequences are introduced by DCS (0x90) or ESC P (0x1b 0x50)
	var i int
	if b[i] == ansi.DCS || b[i] == ansi.ESC {
		i++
	}
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == 'P' {
		i++
	}

	// initial DCS byte
	if i < len(b) && b[i] >= '<' && b[i] <= '?' {
		dcs.Cmd |= int(b[i]) << parser.MarkerShift
	}

	// Scan parameter bytes in the range 0x30-0x3F
	var j int
	for j = 0; i < len(b) && paramsLen < len(params) && b[i] >= 0x30 && b[i] <= 0x3F; i, j = i+1, j+1 {
		if b[i] >= '0' && b[i] <= '9' {
			if params[paramsLen] == parser.MissingParam {
				params[paramsLen] = 0
			}
			params[paramsLen] *= 10
			params[paramsLen] += int(b[i]) - '0'
		}
		if b[i] == ':' {
			params[paramsLen] |= parser.HasMoreFlag
		}
		if b[i] == ';' || b[i] == ':' {
			paramsLen++
			if paramsLen < len(params) {
				// Don't overflow the params slice
				params[paramsLen] = parser.MissingParam
			}
		}
	}

	if j > 0 && paramsLen < len(params) {
		// has parameters
		paramsLen++
	}

	// Scan intermediate bytes in the range 0x20-0x2F
	var intermed byte
	for j := 0; i < len(b) && b[i] >= 0x20 && b[i] <= 0x2F; i, j = i+1, j+1 {
		intermed = b[i]
	}

	// set intermediate byte
	dcs.Cmd |= int(intermed) << parser.IntermedShift

	// Scan final byte in the range 0x40-0x7E
	if i >= len(b) || b[i] < 0x40 || b[i] > 0x7E {
		return i, UnknownMsg(b[:i])
	}

	// Add the final byte
	dcs.Cmd |= int(b[i])
	i++

	start := i // start of the sequence data
	for ; i < len(b); i++ {
		if b[i] == ansi.ST || b[i] == ansi.ESC {
			break
		}
	}

	if i >= len(b) {
		return i, UnknownMsg(b[:i])
	}

	end := i // end of the sequence data
	i++

	// Check 7-bit ST (string terminator) character
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == '\\' {
		i++
	}

	dcs.Params = params[:paramsLen]
	switch cmd := dcs.Cmd; cmd {
	case 'r' | '+'<<parser.IntermedShift:
		// XTGETTCAP responses
		switch param := dcs.Param(0); param {
		case 1: // 1 means valid response, 0 means invalid response
			tc := parseTermcap(b[start:end])
			// XXX: some terminals like KiTTY report invalid responses with
			// their queries i.e. sending a query for "Tc" using "\x1bP+q5463\x1b\\"
			// returns "\x1bP0+r5463\x1b\\".
			// The specs says that invalid responses should be in the form of
			// DCS 0 + r ST "\x1bP0+r\x1b\\"
			// We ignore invalid responses and only send valid ones to the program.
			//
			// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
			return i, tc
		}
	case '|' | '>'<<parser.MarkerShift:
		// XTVersion response
		return i, TerminalVersionMsg(b[start:end])
	}

	return i, UnknownMsg(b[:i])
}

func (p *inputParser) parseApc(b []byte) (int, Msg) {
	if len(b) == 2 && b[0] == ansi.ESC {
		// short cut if this is an alt+_ key
		return 2, KeyPressMsg{Code: rune(b[1]), Mod: ModAlt}
	}

	// APC sequences are introduced by APC (0x9f) or ESC _ (0x1b 0x5f)
	return p.parseStTerminated(ansi.APC, '_')(b)
}

func (p *inputParser) parseUtf8(b []byte) (int, Msg) {
	if len(b) == 0 {
		return 0, nil
	}

	c := b[0]
	if c <= ansi.US || c == ansi.DEL || c == ansi.SP {
		// Control codes get handled by parseControl
		return 1, p.parseControl(c)
	} else if c > ansi.US && c < ansi.DEL {
		// ASCII printable characters
		code := rune(c)
		k := KeyPressMsg{Code: code, Text: string(code)}
		if unicode.IsUpper(code) {
			// Convert upper case letters to lower case + shift modifier
			k.Code = unicode.ToLower(code)
			k.ShiftedCode = code
			k.Mod |= ModShift
		}

		return 1, k
	}

	code, _ := utf8.DecodeRune(b)
	if code == utf8.RuneError {
		return 1, UnknownMsg(b[0])
	}

	cluster, _, _, _ := uniseg.FirstGraphemeCluster(b, -1)
	text := string(cluster)
	for i := range text {
		if i > 0 {
			// Use [KeyExtended] for multi-rune graphemes
			code = KeyExtended
			break
		}
	}

	return len(cluster), KeyPressMsg{Code: code, Text: text}
}

func (p *inputParser) parseControl(b byte) Msg {
	switch b {
	case ansi.NUL:
		if p.flags&_FlagCtrlAt != 0 {
			return KeyPressMsg{Code: '@', Mod: ModCtrl}
		}
		return KeyPressMsg{Code: KeySpace, Mod: ModCtrl}
	case ansi.BS:
		return KeyPressMsg{Code: 'h', Mod: ModCtrl}
	case ansi.HT:
		if p.flags&_FlagCtrlI != 0 {
			return KeyPressMsg{Code: 'i', Mod: ModCtrl}
		}
		return KeyPressMsg{Code: KeyTab}
	case ansi.CR:
		if p.flags&_FlagCtrlM != 0 {
			return KeyPressMsg{Code: 'm', Mod: ModCtrl}
		}
		return KeyPressMsg{Code: KeyEnter}
	case ansi.ESC:
		if p.flags&_FlagCtrlOpenBracket != 0 {
			return KeyPressMsg{Code: '[', Mod: ModCtrl}
		}
		return KeyPressMsg{Code: KeyEscape}
	case ansi.DEL:
		if p.flags&_FlagBackspace != 0 {
			return KeyPressMsg{Code: KeyDelete}
		}
		return KeyPressMsg{Code: KeyBackspace}
	case ansi.SP:
		return KeyPressMsg{Code: KeySpace, Text: " "}
	default:
		if b >= ansi.SOH && b <= ansi.SUB {
			// Use lower case letters for control codes
			code := rune(b + 0x60)
			return KeyPressMsg{Code: code, Mod: ModCtrl}
		} else if b >= ansi.FS && b <= ansi.US {
			code := rune(b + 0x40)
			return KeyPressMsg{Code: code, Mod: ModCtrl}
		}
		return UnknownMsg(b)
	}
}
