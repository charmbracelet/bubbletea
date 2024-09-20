package tea

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// numEvents is the number of events to read from the Windows Console API at a
// time.
const numEvents = 256

// win32KeyState is a state machine for parsing key events from the Windows
// Console API into escape sequences and utf8 runes.
type win32KeyState struct {
	ansiBuf   [numEvents]byte
	ansiIdx   int
	utf16Buf  [2]rune
	utf16Half bool
	lastCks   uint32 // the last control key state for the previous event
}

// parseWin32InputKeyEvent parses a single key event from either the Windows
// Console API or win32-input-mode events. When state is nil, it means this is
// an event from win32-input-mode. Otherwise, it's a key event from the Windows
// Console API and needs a state to decode ANSI escape sequences and utf16
// runes.
func parseWin32InputKeyEvent(state *win32KeyState, vkc uint16, _ uint16, r rune, keyDown bool, cks uint32, repeatCount uint16) (msg Msg) {
	defer func() {
		// Respect the repeat count.
		if repeatCount > 1 {
			var multi multiMsg
			for i := 0; i < int(repeatCount); i++ {
				multi = append(multi, msg)
			}
			msg = multi
		}
	}()
	if state != nil {
		defer func() {
			state.lastCks = cks
		}()
	}

	var utf8Buf [utf8.UTFMax]byte
	var key Key
	if state != nil && state.utf16Half {
		state.utf16Half = false
		state.utf16Buf[1] = r
		codepoint := utf16.DecodeRune(state.utf16Buf[0], state.utf16Buf[1])
		rw := utf8.EncodeRune(utf8Buf[:], codepoint)
		r, _ = utf8.DecodeRune(utf8Buf[:rw])
		key.Code = r
		key.Text = string(r)
		key.Mod = translateControlKeyState(cks)
		key = ensureKeyCase(key, cks)
		if keyDown {
			return KeyPressMsg(key)
		}
		return KeyReleaseMsg(key)
	}

	var baseCode rune
	switch {
	case vkc == 0:
		// Zero means this event is either an escape code or a unicode
		// codepoint.
		if state != nil && state.ansiIdx == 0 && r != ansi.ESC {
			// This is a unicode codepoint.
			baseCode = r
			break
		}

		if state != nil {
			// Collect ANSI escape code.
			state.ansiBuf[state.ansiIdx] = byte(r)
			state.ansiIdx++
			if state.ansiIdx <= 2 {
				// We haven't received enough bytes to determine if this is an
				// ANSI escape code.
				return nil
			}

			n, msg := parseSequence(state.ansiBuf[:state.ansiIdx])
			if n == 0 {
				return nil
			}

			if _, ok := msg.(UnknownMsg); ok {
				return nil
			}

			state.ansiIdx = 0
			return msg
		}
	case vkc == _VK_BACK:
		baseCode = KeyBackspace
	case vkc == _VK_TAB:
		baseCode = KeyTab
	case vkc == _VK_RETURN:
		baseCode = KeyEnter
	case vkc == _VK_SHIFT:
		if cks&_SHIFT_PRESSED != 0 {
			if cks&_ENHANCED_KEY != 0 {
				baseCode = KeyRightShift
			} else {
				baseCode = KeyLeftShift
			}
		} else if state != nil {
			if state.lastCks&_SHIFT_PRESSED != 0 {
				if state.lastCks&_ENHANCED_KEY != 0 {
					baseCode = KeyRightShift
				} else {
					baseCode = KeyLeftShift
				}
			}
		}
	case vkc == _VK_CONTROL:
		if cks&_LEFT_CTRL_PRESSED != 0 {
			baseCode = KeyLeftCtrl
		} else if cks&_RIGHT_CTRL_PRESSED != 0 {
			baseCode = KeyRightCtrl
		} else if state != nil {
			if state.lastCks&_LEFT_CTRL_PRESSED != 0 {
				baseCode = KeyLeftCtrl
			} else if state.lastCks&_RIGHT_CTRL_PRESSED != 0 {
				baseCode = KeyRightCtrl
			}
		}
	case vkc == _VK_MENU:
		if cks&_LEFT_ALT_PRESSED != 0 {
			baseCode = KeyLeftAlt
		} else if cks&_RIGHT_ALT_PRESSED != 0 {
			baseCode = KeyRightAlt
		} else if state != nil {
			if state.lastCks&_LEFT_ALT_PRESSED != 0 {
				baseCode = KeyLeftAlt
			} else if state.lastCks&_RIGHT_ALT_PRESSED != 0 {
				baseCode = KeyRightAlt
			}
		}
	case vkc == _VK_PAUSE:
		baseCode = KeyPause
	case vkc == _VK_CAPITAL:
		baseCode = KeyCapsLock
	case vkc == _VK_ESCAPE:
		baseCode = KeyEscape
	case vkc == _VK_SPACE:
		baseCode = KeySpace
	case vkc == _VK_PRIOR:
		baseCode = KeyPgUp
	case vkc == _VK_NEXT:
		baseCode = KeyPgDown
	case vkc == _VK_END:
		baseCode = KeyEnd
	case vkc == _VK_HOME:
		baseCode = KeyHome
	case vkc == _VK_LEFT:
		baseCode = KeyLeft
	case vkc == _VK_UP:
		baseCode = KeyUp
	case vkc == _VK_RIGHT:
		baseCode = KeyRight
	case vkc == _VK_DOWN:
		baseCode = KeyDown
	case vkc == _VK_SELECT:
		baseCode = KeySelect
	case vkc == _VK_SNAPSHOT:
		baseCode = KeyPrintScreen
	case vkc == _VK_INSERT:
		baseCode = KeyInsert
	case vkc == _VK_DELETE:
		baseCode = KeyDelete
	case vkc >= '0' && vkc <= '9':
		baseCode = rune(vkc)
	case vkc >= 'A' && vkc <= 'Z':
		// Convert to lowercase.
		baseCode = rune(vkc) + 32
	case vkc == _VK_LWIN:
		baseCode = KeyLeftSuper
	case vkc == _VK_RWIN:
		baseCode = KeyRightSuper
	case vkc == _VK_APPS:
		baseCode = KeyMenu
	case vkc >= _VK_NUMPAD0 && vkc <= _VK_NUMPAD9:
		baseCode = rune(vkc-_VK_NUMPAD0) + KeyKp0
	case vkc == _VK_MULTIPLY:
		baseCode = KeyKpMultiply
	case vkc == _VK_ADD:
		baseCode = KeyKpPlus
	case vkc == _VK_SEPARATOR:
		baseCode = KeyKpComma
	case vkc == _VK_SUBTRACT:
		baseCode = KeyKpMinus
	case vkc == _VK_DECIMAL:
		baseCode = KeyKpDecimal
	case vkc == _VK_DIVIDE:
		baseCode = KeyKpDivide
	case vkc >= _VK_F1 && vkc <= _VK_F24:
		baseCode = rune(vkc-_VK_F1) + KeyF1
	case vkc == _VK_NUMLOCK:
		baseCode = KeyNumLock
	case vkc == _VK_SCROLL:
		baseCode = KeyScrollLock
	case vkc == _VK_LSHIFT:
		baseCode = KeyLeftShift
	case vkc == _VK_RSHIFT:
		baseCode = KeyRightShift
	case vkc == _VK_LCONTROL:
		baseCode = KeyLeftCtrl
	case vkc == _VK_RCONTROL:
		baseCode = KeyRightCtrl
	case vkc == _VK_LMENU:
		baseCode = KeyLeftAlt
	case vkc == _VK_RMENU:
		baseCode = KeyRightAlt
	case vkc == _VK_VOLUME_MUTE:
		baseCode = KeyMute
	case vkc == _VK_VOLUME_DOWN:
		baseCode = KeyLowerVol
	case vkc == _VK_VOLUME_UP:
		baseCode = KeyRaiseVol
	case vkc == _VK_MEDIA_NEXT_TRACK:
		baseCode = KeyMediaNext
	case vkc == _VK_MEDIA_PREV_TRACK:
		baseCode = KeyMediaPrev
	case vkc == _VK_MEDIA_STOP:
		baseCode = KeyMediaStop
	case vkc == _VK_MEDIA_PLAY_PAUSE:
		baseCode = KeyMediaPlayPause
	case vkc == _VK_OEM_1:
		baseCode = ';'
	case vkc == _VK_OEM_PLUS:
		baseCode = '+'
	case vkc == _VK_OEM_COMMA:
		baseCode = ','
	case vkc == _VK_OEM_MINUS:
		baseCode = '-'
	case vkc == _VK_OEM_PERIOD:
		baseCode = '.'
	case vkc == _VK_OEM_2:
		baseCode = '/'
	case vkc == _VK_OEM_3:
		baseCode = '`'
	case vkc == _VK_OEM_4:
		baseCode = '['
	case vkc == _VK_OEM_5:
		baseCode = '\\'
	case vkc == _VK_OEM_6:
		baseCode = ']'
	case vkc == _VK_OEM_7:
		baseCode = '\''
	}

	if utf16.IsSurrogate(r) {
		if state != nil {
			state.utf16Buf[0] = r
			state.utf16Half = true
		}
		return nil
	}

	// AltGr is left ctrl + right alt. On non-US keyboards, this is used to type
	// special characters and produce printable events.
	// XXX: Should this be a KeyMod?
	altGr := cks&(_LEFT_CTRL_PRESSED|_RIGHT_ALT_PRESSED) == _LEFT_CTRL_PRESSED|_RIGHT_ALT_PRESSED

	var text string
	keyCode := baseCode
	if r >= ansi.NUL && r <= ansi.US {
		// Control characters.
	} else {
		rw := utf8.EncodeRune(utf8Buf[:], r)
		keyCode, _ = utf8.DecodeRune(utf8Buf[:rw])
		if cks == _NO_CONTROL_KEY ||
			cks == _SHIFT_PRESSED ||
			cks == _CAPSLOCK_ON ||
			altGr {
			// If the control key state is 0, shift is pressed, or caps lock
			// then the key event is a printable event i.e. [text] is not empty.
			text = string(keyCode)
		}
	}

	key.Code = keyCode
	key.Text = text
	key.Mod = translateControlKeyState(cks)
	key.BaseCode = baseCode
	key = ensureKeyCase(key, cks)
	if keyDown {
		return KeyPressMsg(key)
	}

	return KeyReleaseMsg(key)
}

// ensureKeyCase ensures that the key's text is in the correct case based on the
// control key state.
func ensureKeyCase(key Key, cks uint32) Key {
	if len(key.Text) == 0 {
		return key
	}

	hasShift := cks&_SHIFT_PRESSED != 0
	hasCaps := cks&_CAPSLOCK_ON != 0
	if hasShift || hasCaps {
		if unicode.IsLower(key.Code) {
			key.ShiftedCode = unicode.ToUpper(key.Code)
			key.Text = string(key.ShiftedCode)
		}
	} else {
		if unicode.IsUpper(key.Code) {
			key.ShiftedCode = unicode.ToLower(key.Code)
			key.Text = string(key.ShiftedCode)
		}
	}

	return key
}

// translateControlKeyState translates the control key state from the Windows
// Console API into a Mod bitmask.
func translateControlKeyState(cks uint32) (m KeyMod) {
	if cks&_LEFT_CTRL_PRESSED != 0 || cks&_RIGHT_CTRL_PRESSED != 0 {
		m |= ModCtrl
	}
	if cks&_LEFT_ALT_PRESSED != 0 || cks&_RIGHT_ALT_PRESSED != 0 {
		m |= ModAlt
	}
	if cks&_SHIFT_PRESSED != 0 {
		m |= ModShift
	}
	if cks&_CAPSLOCK_ON != 0 {
		m |= ModCapsLock
	}
	if cks&_NUMLOCK_ON != 0 {
		m |= ModNumLock
	}
	if cks&_SCROLLLOCK_ON != 0 {
		m |= ModScrollLock
	}
	return
}

//nolint:unused
func keyEventString(vkc, sc uint16, r rune, keyDown bool, cks uint32, repeatCount uint16) string {
	var s strings.Builder
	s.WriteString("vkc: ")
	s.WriteString(fmt.Sprintf("%d, 0x%02x", vkc, vkc))
	s.WriteString(", sc: ")
	s.WriteString(fmt.Sprintf("%d, 0x%02x", sc, sc))
	s.WriteString(", r: ")
	s.WriteString(fmt.Sprintf("%q", r))
	s.WriteString(", down: ")
	s.WriteString(fmt.Sprintf("%v", keyDown))
	s.WriteString(", cks: [")
	if cks&_LEFT_ALT_PRESSED != 0 {
		s.WriteString("left alt, ")
	}
	if cks&_RIGHT_ALT_PRESSED != 0 {
		s.WriteString("right alt, ")
	}
	if cks&_LEFT_CTRL_PRESSED != 0 {
		s.WriteString("left ctrl, ")
	}
	if cks&_RIGHT_CTRL_PRESSED != 0 {
		s.WriteString("right ctrl, ")
	}
	if cks&_SHIFT_PRESSED != 0 {
		s.WriteString("shift, ")
	}
	if cks&_CAPSLOCK_ON != 0 {
		s.WriteString("caps lock, ")
	}
	if cks&_NUMLOCK_ON != 0 {
		s.WriteString("num lock, ")
	}
	if cks&_SCROLLLOCK_ON != 0 {
		s.WriteString("scroll lock, ")
	}
	if cks&_ENHANCED_KEY != 0 {
		s.WriteString("enhanced key, ")
	}
	s.WriteString("], repeat count: ")
	s.WriteString(fmt.Sprintf("%d", repeatCount))
	return s.String()
}

//nolint:revive
const (
	_VK_LBUTTON             = 0x01
	_VK_RBUTTON             = 0x02
	_VK_CANCEL              = 0x03
	_VK_MBUTTON             = 0x04
	_VK_XBUTTON1            = 0x05
	_VK_XBUTTON2            = 0x06
	_VK_BACK                = 0x08
	_VK_TAB                 = 0x09
	_VK_CLEAR               = 0x0C
	_VK_RETURN              = 0x0D
	_VK_SHIFT               = 0x10
	_VK_CONTROL             = 0x11
	_VK_MENU                = 0x12
	_VK_PAUSE               = 0x13
	_VK_CAPITAL             = 0x14
	_VK_KANA                = 0x15
	_VK_HANGEUL             = 0x15
	_VK_HANGUL              = 0x15
	_VK_IME_ON              = 0x16
	_VK_JUNJA               = 0x17
	_VK_FINAL               = 0x18
	_VK_HANJA               = 0x19
	_VK_KANJI               = 0x19
	_VK_IME_OFF             = 0x1A
	_VK_ESCAPE              = 0x1B
	_VK_CONVERT             = 0x1C
	_VK_NONCONVERT          = 0x1D
	_VK_ACCEPT              = 0x1E
	_VK_MODECHANGE          = 0x1F
	_VK_SPACE               = 0x20
	_VK_PRIOR               = 0x21
	_VK_NEXT                = 0x22
	_VK_END                 = 0x23
	_VK_HOME                = 0x24
	_VK_LEFT                = 0x25
	_VK_UP                  = 0x26
	_VK_RIGHT               = 0x27
	_VK_DOWN                = 0x28
	_VK_SELECT              = 0x29
	_VK_PRINT               = 0x2A
	_VK_EXECUTE             = 0x2B
	_VK_SNAPSHOT            = 0x2C
	_VK_INSERT              = 0x2D
	_VK_DELETE              = 0x2E
	_VK_HELP                = 0x2F
	_VK_0                   = 0x30
	_VK_1                   = 0x31
	_VK_2                   = 0x32
	_VK_3                   = 0x33
	_VK_4                   = 0x34
	_VK_5                   = 0x35
	_VK_6                   = 0x36
	_VK_7                   = 0x37
	_VK_8                   = 0x38
	_VK_9                   = 0x39
	_VK_A                   = 0x41
	_VK_B                   = 0x42
	_VK_C                   = 0x43
	_VK_D                   = 0x44
	_VK_E                   = 0x45
	_VK_F                   = 0x46
	_VK_G                   = 0x47
	_VK_H                   = 0x48
	_VK_I                   = 0x49
	_VK_J                   = 0x4A
	_VK_K                   = 0x4B
	_VK_L                   = 0x4C
	_VK_M                   = 0x4D
	_VK_N                   = 0x4E
	_VK_O                   = 0x4F
	_VK_P                   = 0x50
	_VK_Q                   = 0x51
	_VK_R                   = 0x52
	_VK_S                   = 0x53
	_VK_T                   = 0x54
	_VK_U                   = 0x55
	_VK_V                   = 0x56
	_VK_W                   = 0x57
	_VK_X                   = 0x58
	_VK_Y                   = 0x59
	_VK_Z                   = 0x5A
	_VK_LWIN                = 0x5B
	_VK_RWIN                = 0x5C
	_VK_APPS                = 0x5D
	_VK_SLEEP               = 0x5F
	_VK_NUMPAD0             = 0x60
	_VK_NUMPAD1             = 0x61
	_VK_NUMPAD2             = 0x62
	_VK_NUMPAD3             = 0x63
	_VK_NUMPAD4             = 0x64
	_VK_NUMPAD5             = 0x65
	_VK_NUMPAD6             = 0x66
	_VK_NUMPAD7             = 0x67
	_VK_NUMPAD8             = 0x68
	_VK_NUMPAD9             = 0x69
	_VK_MULTIPLY            = 0x6A
	_VK_ADD                 = 0x6B
	_VK_SEPARATOR           = 0x6C
	_VK_SUBTRACT            = 0x6D
	_VK_DECIMAL             = 0x6E
	_VK_DIVIDE              = 0x6F
	_VK_F1                  = 0x70
	_VK_F2                  = 0x71
	_VK_F3                  = 0x72
	_VK_F4                  = 0x73
	_VK_F5                  = 0x74
	_VK_F6                  = 0x75
	_VK_F7                  = 0x76
	_VK_F8                  = 0x77
	_VK_F9                  = 0x78
	_VK_F10                 = 0x79
	_VK_F11                 = 0x7A
	_VK_F12                 = 0x7B
	_VK_F13                 = 0x7C
	_VK_F14                 = 0x7D
	_VK_F15                 = 0x7E
	_VK_F16                 = 0x7F
	_VK_F17                 = 0x80
	_VK_F18                 = 0x81
	_VK_F19                 = 0x82
	_VK_F20                 = 0x83
	_VK_F21                 = 0x84
	_VK_F22                 = 0x85
	_VK_F23                 = 0x86
	_VK_F24                 = 0x87
	_VK_NUMLOCK             = 0x90
	_VK_SCROLL              = 0x91
	_VK_OEM_NEC_EQUAL       = 0x92
	_VK_OEM_FJ_JISHO        = 0x92
	_VK_OEM_FJ_MASSHOU      = 0x93
	_VK_OEM_FJ_TOUROKU      = 0x94
	_VK_OEM_FJ_LOYA         = 0x95
	_VK_OEM_FJ_ROYA         = 0x96
	_VK_LSHIFT              = 0xA0
	_VK_RSHIFT              = 0xA1
	_VK_LCONTROL            = 0xA2
	_VK_RCONTROL            = 0xA3
	_VK_LMENU               = 0xA4
	_VK_RMENU               = 0xA5
	_VK_BROWSER_BACK        = 0xA6
	_VK_BROWSER_FORWARD     = 0xA7
	_VK_BROWSER_REFRESH     = 0xA8
	_VK_BROWSER_STOP        = 0xA9
	_VK_BROWSER_SEARCH      = 0xAA
	_VK_BROWSER_FAVORITES   = 0xAB
	_VK_BROWSER_HOME        = 0xAC
	_VK_VOLUME_MUTE         = 0xAD
	_VK_VOLUME_DOWN         = 0xAE
	_VK_VOLUME_UP           = 0xAF
	_VK_MEDIA_NEXT_TRACK    = 0xB0
	_VK_MEDIA_PREV_TRACK    = 0xB1
	_VK_MEDIA_STOP          = 0xB2
	_VK_MEDIA_PLAY_PAUSE    = 0xB3
	_VK_LAUNCH_MAIL         = 0xB4
	_VK_LAUNCH_MEDIA_SELECT = 0xB5
	_VK_LAUNCH_APP1         = 0xB6
	_VK_LAUNCH_APP2         = 0xB7
	_VK_OEM_1               = 0xBA
	_VK_OEM_PLUS            = 0xBB
	_VK_OEM_COMMA           = 0xBC
	_VK_OEM_MINUS           = 0xBD
	_VK_OEM_PERIOD          = 0xBE
	_VK_OEM_2               = 0xBF
	_VK_OEM_3               = 0xC0
	_VK_OEM_4               = 0xDB
	_VK_OEM_5               = 0xDC
	_VK_OEM_6               = 0xDD
	_VK_OEM_7               = 0xDE
	_VK_OEM_8               = 0xDF
	_VK_OEM_AX              = 0xE1
	_VK_OEM_102             = 0xE2
	_VK_ICO_HELP            = 0xE3
	_VK_ICO_00              = 0xE4
	_VK_PROCESSKEY          = 0xE5
	_VK_ICO_CLEAR           = 0xE6
	_VK_OEM_RESET           = 0xE9
	_VK_OEM_JUMP            = 0xEA
	_VK_OEM_PA1             = 0xEB
	_VK_OEM_PA2             = 0xEC
	_VK_OEM_PA3             = 0xED
	_VK_OEM_WSCTRL          = 0xEE
	_VK_OEM_CUSEL           = 0xEF
	_VK_OEM_ATTN            = 0xF0
	_VK_OEM_FINISH          = 0xF1
	_VK_OEM_COPY            = 0xF2
	_VK_OEM_AUTO            = 0xF3
	_VK_OEM_ENLW            = 0xF4
	_VK_OEM_BACKTAB         = 0xF5
	_VK_ATTN                = 0xF6
	_VK_CRSEL               = 0xF7
	_VK_EXSEL               = 0xF8
	_VK_EREOF               = 0xF9
	_VK_PLAY                = 0xFA
	_VK_ZOOM                = 0xFB
	_VK_NONAME              = 0xFC
	_VK_PA1                 = 0xFD
	_VK_OEM_CLEAR           = 0xFE
)

//nolint:revive
const (
	_CAPSLOCK_ON        = 0x0080
	_ENHANCED_KEY       = 0x0100
	_LEFT_ALT_PRESSED   = 0x0002
	_LEFT_CTRL_PRESSED  = 0x0008
	_NUMLOCK_ON         = 0x0020
	_RIGHT_ALT_PRESSED  = 0x0001
	_RIGHT_CTRL_PRESSED = 0x0004
	_SCROLLLOCK_ON      = 0x0040
	_SHIFT_PRESSED      = 0x0010
	_NO_CONTROL_KEY     = 0x0000
)
