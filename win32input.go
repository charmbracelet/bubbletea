package tea

import (
	"unicode"

	"github.com/erikgeiser/coninput"
)

func parseWin32InputKeyEvent(vkc coninput.VirtualKeyCode, _ coninput.VirtualKeyCode, r rune, keyDown bool, cks coninput.ControlKeyState, repeatCount uint16) Msg {
	var key Key
	isCtrl := cks.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED)
	switch vkc {
	case coninput.VK_SHIFT:
		// We currently ignore these keys when they are pressed alone.
		return nil
	case coninput.VK_MENU:
		if cks.Contains(coninput.LEFT_ALT_PRESSED) {
			key = Key{Type: KeyLeftAlt}
		} else if cks.Contains(coninput.RIGHT_ALT_PRESSED) {
			key = Key{Type: KeyRightAlt}
		} else if !keyDown {
			return nil
		}
	case coninput.VK_CONTROL:
		if cks.Contains(coninput.LEFT_CTRL_PRESSED) {
			key = Key{Type: KeyLeftCtrl}
		} else if cks.Contains(coninput.RIGHT_CTRL_PRESSED) {
			key = Key{Type: KeyRightCtrl}
		} else if !keyDown {
			return nil
		}
	case coninput.VK_CAPITAL:
		key = Key{Type: KeyCapsLock}
	default:
		var ok bool
		key, ok = vkKeyEvent[vkc]
		if !ok {
			if isCtrl {
				key = vkCtrlRune(key, r, vkc)
			} else {
				key = Key{Runes: []rune{r}}
			}
		}
	}

	if isCtrl {
		key.Mod |= ModCtrl
	}
	if cks.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED) {
		key.Mod |= ModAlt
	}
	if cks.Contains(coninput.SHIFT_PRESSED) {
		key.Mod |= ModShift
	}
	if cks.Contains(coninput.CAPSLOCK_ON) {
		key.Mod |= ModCapsLock
	}
	if cks.Contains(coninput.NUMLOCK_ON) {
		key.Mod |= ModNumLock
	}
	if cks.Contains(coninput.SCROLLLOCK_ON) {
		key.Mod |= ModScrollLock
	}

	// Use the unshifted key
	if cks.Contains(coninput.SHIFT_PRESSED ^ coninput.CAPSLOCK_ON) {
		key.altRune = unicode.ToUpper(key.Rune())
	} else {
		key.altRune = unicode.ToLower(key.Rune())
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

var vkKeyEvent = map[coninput.VirtualKeyCode]Key{
	coninput.VK_RETURN:    {Type: KeyEnter},
	coninput.VK_BACK:      {Type: KeyBackspace},
	coninput.VK_TAB:       {Type: KeyTab},
	coninput.VK_ESCAPE:    {Type: KeyEscape},
	coninput.VK_SPACE:     {Type: KeySpace, Runes: []rune{' '}},
	coninput.VK_UP:        {Type: KeyUp},
	coninput.VK_DOWN:      {Type: KeyDown},
	coninput.VK_RIGHT:     {Type: KeyRight},
	coninput.VK_LEFT:      {Type: KeyLeft},
	coninput.VK_HOME:      {Type: KeyHome},
	coninput.VK_END:       {Type: KeyEnd},
	coninput.VK_PRIOR:     {Type: KeyPgUp},
	coninput.VK_NEXT:      {Type: KeyPgDown},
	coninput.VK_DELETE:    {Type: KeyDelete},
	coninput.VK_SELECT:    {Type: KeySelect},
	coninput.VK_SNAPSHOT:  {Type: KeyPrintScreen},
	coninput.VK_INSERT:    {Type: KeyInsert},
	coninput.VK_LWIN:      {Type: KeyLeftSuper},
	coninput.VK_RWIN:      {Type: KeyRightSuper},
	coninput.VK_APPS:      {Type: KeyMenu},
	coninput.VK_NUMPAD0:   {Type: KeyKp0},
	coninput.VK_NUMPAD1:   {Type: KeyKp1},
	coninput.VK_NUMPAD2:   {Type: KeyKp2},
	coninput.VK_NUMPAD3:   {Type: KeyKp3},
	coninput.VK_NUMPAD4:   {Type: KeyKp4},
	coninput.VK_NUMPAD5:   {Type: KeyKp5},
	coninput.VK_NUMPAD6:   {Type: KeyKp6},
	coninput.VK_NUMPAD7:   {Type: KeyKp7},
	coninput.VK_NUMPAD8:   {Type: KeyKp8},
	coninput.VK_NUMPAD9:   {Type: KeyKp9},
	coninput.VK_MULTIPLY:  {Type: KeyKpMultiply},
	coninput.VK_ADD:       {Type: KeyKpPlus},
	coninput.VK_SEPARATOR: {Type: KeyKpComma},
	coninput.VK_SUBTRACT:  {Type: KeyKpMinus},
	coninput.VK_DECIMAL:   {Type: KeyKpDecimal},
	coninput.VK_DIVIDE:    {Type: KeyKpDivide},
	coninput.VK_F1:        {Type: KeyF1},
	coninput.VK_F2:        {Type: KeyF2},
	coninput.VK_F3:        {Type: KeyF3},
	coninput.VK_F4:        {Type: KeyF4},
	coninput.VK_F5:        {Type: KeyF5},
	coninput.VK_F6:        {Type: KeyF6},
	coninput.VK_F7:        {Type: KeyF7},
	coninput.VK_F8:        {Type: KeyF8},
	coninput.VK_F9:        {Type: KeyF9},
	coninput.VK_F10:       {Type: KeyF10},
	coninput.VK_F11:       {Type: KeyF11},
	coninput.VK_F12:       {Type: KeyF12},
	coninput.VK_F13:       {Type: KeyF13},
	coninput.VK_F14:       {Type: KeyF14},
	coninput.VK_F15:       {Type: KeyF15},
	coninput.VK_F16:       {Type: KeyF16},
	coninput.VK_F17:       {Type: KeyF17},
	coninput.VK_F18:       {Type: KeyF18},
	coninput.VK_F19:       {Type: KeyF19},
	coninput.VK_F20:       {Type: KeyF20},
	coninput.VK_F21:       {Type: KeyF21},
	coninput.VK_F22:       {Type: KeyF22},
	coninput.VK_F23:       {Type: KeyF23},
	coninput.VK_F24:       {Type: KeyF24},
	coninput.VK_NUMLOCK:   {Type: KeyNumLock},
	coninput.VK_SCROLL:    {Type: KeyScrollLock},
	coninput.VK_LSHIFT:    {Type: KeyLeftShift},
	coninput.VK_RSHIFT:    {Type: KeyRightShift},
	coninput.VK_LCONTROL:  {Type: KeyLeftCtrl},
	coninput.VK_RCONTROL:  {Type: KeyRightCtrl},
	coninput.VK_LMENU:     {Type: KeyLeftAlt},
	coninput.VK_RMENU:     {Type: KeyRightAlt},
	coninput.VK_OEM_4:     {Runes: []rune{'['}},
	// TODO: add more keys
}

func vkCtrlRune(k Key, r rune, kc coninput.VirtualKeyCode) Key {
	switch r {
	case '@':
		k.Runes = []rune{'@'}
	case '\x01':
		k.Runes = []rune{'a'}
	case '\x02':
		k.Runes = []rune{'b'}
	case '\x03':
		k.Runes = []rune{'c'}
	case '\x04':
		k.Runes = []rune{'d'}
	case '\x05':
		k.Runes = []rune{'e'}
	case '\x06':
		k.Runes = []rune{'f'}
	case '\a':
		k.Runes = []rune{'g'}
	case '\b':
		k.Runes = []rune{'h'}
	case '\t':
		k.Runes = []rune{'i'}
	case '\n':
		k.Runes = []rune{'j'}
	case '\v':
		k.Runes = []rune{'k'}
	case '\f':
		k.Runes = []rune{'l'}
	case '\r':
		k.Runes = []rune{'m'}
	case '\x0e':
		k.Runes = []rune{'n'}
	case '\x0f':
		k.Runes = []rune{'o'}
	case '\x10':
		k.Runes = []rune{'p'}
	case '\x11':
		k.Runes = []rune{'q'}
	case '\x12':
		k.Runes = []rune{'r'}
	case '\x13':
		k.Runes = []rune{'s'}
	case '\x14':
		k.Runes = []rune{'t'}
	case '\x15':
		k.Runes = []rune{'u'}
	case '\x16':
		k.Runes = []rune{'v'}
	case '\x17':
		k.Runes = []rune{'w'}
	case '\x18':
		k.Runes = []rune{'x'}
	case '\x19':
		k.Runes = []rune{'y'}
	case '\x1a':
		k.Runes = []rune{'z'}
	case '\x1b':
		k.Runes = []rune{']'}
	case '\x1c':
		k.Runes = []rune{'\\'}
	case '\x1f':
		k.Runes = []rune{'_'}
	}

	switch kc {
	case coninput.VK_OEM_4:
		k.Runes = []rune{'['}
	}

	// https://learn.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes
	if len(k.Runes) == 0 &&
		(kc >= 0x30 && kc <= 0x39) ||
		(kc >= 0x41 && kc <= 0x5a) {
		k.Runes = []rune{rune(kc)}
	}

	return k
}
