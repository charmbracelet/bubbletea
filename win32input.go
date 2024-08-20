package tea

import (
	"unicode"
)

// enableWin32InputMsg is a message that enables Windows input mode.
type enableWin32InputMsg struct{}

// enableWindowsInputMode is a command that enables Windows input mode
// (win32-input-mode).
//
// See
// https://github.com/microsoft/terminal/blob/main/doc/specs/%234999%20-%20Improved%20keyboard%20handling%20in%20Conpty.md
// for more information.
func enableWindowsInputMode() Msg { //nolint:unused
	return enableWin32InputMsg{}
}

// disableWin32InputMsg is a message that disables Windows input mode.
type disableWin32InputMsg struct{}

// disableWindowsInputMode is a command that disables Windows input mode
// (win32-input-mode).
//
// See
// https://github.com/microsoft/terminal/blob/main/doc/specs/%234999%20-%20Improved%20keyboard%20handling%20in%20Conpty.md
// for more information.
func disableWindowsInputMode() Msg { //nolint:unused
	return disableWin32InputMsg{}
}

func parseWin32InputKeyEvent(vkc uint16, _ uint16, r rune, keyDown bool, cks uint32, repeatCount uint16) Msg {
	var key Key
	isCtrl := cks&(_LEFT_CTRL_PRESSED|_RIGHT_CTRL_PRESSED) != 0
	switch vkc {
	case _VK_SHIFT:
		// We currently ignore these keys when they are pressed alone.
		return nil
	case _VK_MENU:
		if cks&_LEFT_ALT_PRESSED != 0 {
			key.Type = KeyLeftAlt
		} else if cks&_RIGHT_ALT_PRESSED != 0 {
			key.Type = KeyRightAlt
		} else if !keyDown {
			return nil
		}
	case _VK_CONTROL:
		if cks&_LEFT_CTRL_PRESSED != 0 {
			key.Type = KeyLeftCtrl
		} else if cks&_RIGHT_CTRL_PRESSED != 0 {
			key.Type = KeyRightCtrl
		} else if !keyDown {
			return nil
		}
	case _VK_CAPITAL:
		key.Type = KeyCapsLock
	default:
		var ok bool
		key, ok = vkKeyEvent[vkc]
		if !ok {
			if isCtrl {
				key.Runes = []rune{vkCtrlRune(key, r, vkc)}
			} else {
				key.Runes = []rune{r}
			}
		}
	}

	if isCtrl {
		key.Mod |= ModCtrl
	}
	if cks&(_LEFT_ALT_PRESSED|_RIGHT_ALT_PRESSED) != 0 {
		key.Mod |= ModAlt
	}
	if cks&_SHIFT_PRESSED != 0 {
		key.Mod |= ModShift
	}
	if cks&_CAPSLOCK_ON != 0 {
		key.Mod |= ModCapsLock
	}
	if cks&_NUMLOCK_ON != 0 {
		key.Mod |= ModNumLock
	}
	if cks&_SCROLLLOCK_ON != 0 {
		key.Mod |= ModScrollLock
	}

	// Use the unshifted key
	keyRune := key.Rune()
	if cks&(_SHIFT_PRESSED^_CAPSLOCK_ON) != 0 {
		if unicode.IsLower(keyRune) {
			key.altRune = unicode.ToUpper(key.Rune())
		}
	} else {
		if unicode.IsUpper(keyRune) {
			key.altRune = unicode.ToLower(keyRune)
		}
	}

	var e Msg = KeyPressMsg(key)
	key.IsRepeat = repeatCount > 1
	if !keyDown {
		e = KeyReleaseMsg(key)
	}

	if repeatCount <= 1 {
		return e
	}

	var kevents []Msg
	for i := 0; i < int(repeatCount); i++ {
		kevents = append(kevents, e)
	}

	return multiMsg(kevents)
}

var vkKeyEvent = map[uint16]Key{
	_VK_RETURN:    {Type: KeyEnter},
	_VK_BACK:      {Type: KeyBackspace},
	_VK_TAB:       {Type: KeyTab},
	_VK_ESCAPE:    {Type: KeyEscape},
	_VK_SPACE:     {Type: KeySpace, Runes: []rune{' '}},
	_VK_UP:        {Type: KeyUp},
	_VK_DOWN:      {Type: KeyDown},
	_VK_RIGHT:     {Type: KeyRight},
	_VK_LEFT:      {Type: KeyLeft},
	_VK_HOME:      {Type: KeyHome},
	_VK_END:       {Type: KeyEnd},
	_VK_PRIOR:     {Type: KeyPgUp},
	_VK_NEXT:      {Type: KeyPgDown},
	_VK_DELETE:    {Type: KeyDelete},
	_VK_SELECT:    {Type: KeySelect},
	_VK_SNAPSHOT:  {Type: KeyPrintScreen},
	_VK_INSERT:    {Type: KeyInsert},
	_VK_LWIN:      {Type: KeyLeftSuper},
	_VK_RWIN:      {Type: KeyRightSuper},
	_VK_APPS:      {Type: KeyMenu},
	_VK_NUMPAD0:   {Type: KeyKp0},
	_VK_NUMPAD1:   {Type: KeyKp1},
	_VK_NUMPAD2:   {Type: KeyKp2},
	_VK_NUMPAD3:   {Type: KeyKp3},
	_VK_NUMPAD4:   {Type: KeyKp4},
	_VK_NUMPAD5:   {Type: KeyKp5},
	_VK_NUMPAD6:   {Type: KeyKp6},
	_VK_NUMPAD7:   {Type: KeyKp7},
	_VK_NUMPAD8:   {Type: KeyKp8},
	_VK_NUMPAD9:   {Type: KeyKp9},
	_VK_MULTIPLY:  {Type: KeyKpMultiply},
	_VK_ADD:       {Type: KeyKpPlus},
	_VK_SEPARATOR: {Type: KeyKpComma},
	_VK_SUBTRACT:  {Type: KeyKpMinus},
	_VK_DECIMAL:   {Type: KeyKpDecimal},
	_VK_DIVIDE:    {Type: KeyKpDivide},
	_VK_F1:        {Type: KeyF1},
	_VK_F2:        {Type: KeyF2},
	_VK_F3:        {Type: KeyF3},
	_VK_F4:        {Type: KeyF4},
	_VK_F5:        {Type: KeyF5},
	_VK_F6:        {Type: KeyF6},
	_VK_F7:        {Type: KeyF7},
	_VK_F8:        {Type: KeyF8},
	_VK_F9:        {Type: KeyF9},
	_VK_F10:       {Type: KeyF10},
	_VK_F11:       {Type: KeyF11},
	_VK_F12:       {Type: KeyF12},
	_VK_F13:       {Type: KeyF13},
	_VK_F14:       {Type: KeyF14},
	_VK_F15:       {Type: KeyF15},
	_VK_F16:       {Type: KeyF16},
	_VK_F17:       {Type: KeyF17},
	_VK_F18:       {Type: KeyF18},
	_VK_F19:       {Type: KeyF19},
	_VK_F20:       {Type: KeyF20},
	_VK_F21:       {Type: KeyF21},
	_VK_F22:       {Type: KeyF22},
	_VK_F23:       {Type: KeyF23},
	_VK_F24:       {Type: KeyF24},
	_VK_NUMLOCK:   {Type: KeyNumLock},
	_VK_SCROLL:    {Type: KeyScrollLock},
	_VK_LSHIFT:    {Type: KeyLeftShift},
	_VK_RSHIFT:    {Type: KeyRightShift},
	_VK_LCONTROL:  {Type: KeyLeftCtrl},
	_VK_RCONTROL:  {Type: KeyRightCtrl},
	_VK_LMENU:     {Type: KeyLeftAlt},
	_VK_RMENU:     {Type: KeyRightAlt},
	_VK_OEM_4:     {Runes: []rune{'['}},
	// TODO: add more keys
}

func vkCtrlRune(k Key, r rune, kc uint16) rune {
	switch r {
	case 0x01:
		return 'a'
	case 0x02:
		return 'b'
	case 0x03:
		return 'c'
	case 0x04:
		return 'd'
	case 0x05:
		return 'e'
	case 0x06:
		return 'f'
	case '\a':
		return 'g'
	case '\b':
		return 'h'
	case '\t':
		return 'i'
	case '\n':
		return 'j'
	case '\v':
		return 'k'
	case '\f':
		return 'l'
	case '\r':
		return 'm'
	case 0x0e:
		return 'n'
	case 0x0f:
		return 'o'
	case 0x10:
		return 'p'
	case 0x11:
		return 'q'
	case 0x12:
		return 'r'
	case 0x13:
		return 's'
	case 0x14:
		return 't'
	case 0x15:
		return 'u'
	case 0x16:
		return 'v'
	case 0x17:
		return 'w'
	case 0x18:
		return 'x'
	case 0x19:
		return 'y'
	case 0x1a:
		return 'z'
	case 0x1b:
		return ']'
	case 0x1c:
		return '\\'
	case 0x1f:
		return '_'
	}

	switch kc {
	case _VK_OEM_4:
		return '['
	}

	// https://learn.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes
	if len(k.Runes) == 0 &&
		(kc >= 0x30 && kc <= 0x39) ||
		(kc >= 0x41 && kc <= 0x5a) {
		return rune(kc)
	}

	return r
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
