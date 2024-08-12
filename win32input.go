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
			key = Key{Sym: KeyLeftAlt}
		} else if cks.Contains(coninput.RIGHT_ALT_PRESSED) {
			key = Key{Sym: KeyRightAlt}
		} else if !keyDown {
			return nil
		}
	case coninput.VK_CONTROL:
		if cks.Contains(coninput.LEFT_CTRL_PRESSED) {
			key = Key{Sym: KeyLeftCtrl}
		} else if cks.Contains(coninput.RIGHT_CTRL_PRESSED) {
			key = Key{Sym: KeyRightCtrl}
		} else if !keyDown {
			return nil
		}
	case coninput.VK_CAPITAL:
		key = Key{Sym: KeyCapsLock}
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
	coninput.VK_RETURN:    {Sym: KeyEnter},
	coninput.VK_BACK:      {Sym: KeyBackspace},
	coninput.VK_TAB:       {Sym: KeyTab},
	coninput.VK_ESCAPE:    {Sym: KeyEscape},
	coninput.VK_SPACE:     {Sym: KeySpace, Runes: []rune{' '}},
	coninput.VK_UP:        {Sym: KeyUp},
	coninput.VK_DOWN:      {Sym: KeyDown},
	coninput.VK_RIGHT:     {Sym: KeyRight},
	coninput.VK_LEFT:      {Sym: KeyLeft},
	coninput.VK_HOME:      {Sym: KeyHome},
	coninput.VK_END:       {Sym: KeyEnd},
	coninput.VK_PRIOR:     {Sym: KeyPgUp},
	coninput.VK_NEXT:      {Sym: KeyPgDown},
	coninput.VK_DELETE:    {Sym: KeyDelete},
	coninput.VK_SELECT:    {Sym: KeySelect},
	coninput.VK_SNAPSHOT:  {Sym: KeyPrintScreen},
	coninput.VK_INSERT:    {Sym: KeyInsert},
	coninput.VK_LWIN:      {Sym: KeyLeftSuper},
	coninput.VK_RWIN:      {Sym: KeyRightSuper},
	coninput.VK_APPS:      {Sym: KeyMenu},
	coninput.VK_NUMPAD0:   {Sym: KeyKp0},
	coninput.VK_NUMPAD1:   {Sym: KeyKp1},
	coninput.VK_NUMPAD2:   {Sym: KeyKp2},
	coninput.VK_NUMPAD3:   {Sym: KeyKp3},
	coninput.VK_NUMPAD4:   {Sym: KeyKp4},
	coninput.VK_NUMPAD5:   {Sym: KeyKp5},
	coninput.VK_NUMPAD6:   {Sym: KeyKp6},
	coninput.VK_NUMPAD7:   {Sym: KeyKp7},
	coninput.VK_NUMPAD8:   {Sym: KeyKp8},
	coninput.VK_NUMPAD9:   {Sym: KeyKp9},
	coninput.VK_MULTIPLY:  {Sym: KeyKpMultiply},
	coninput.VK_ADD:       {Sym: KeyKpPlus},
	coninput.VK_SEPARATOR: {Sym: KeyKpComma},
	coninput.VK_SUBTRACT:  {Sym: KeyKpMinus},
	coninput.VK_DECIMAL:   {Sym: KeyKpDecimal},
	coninput.VK_DIVIDE:    {Sym: KeyKpDivide},
	coninput.VK_F1:        {Sym: KeyF1},
	coninput.VK_F2:        {Sym: KeyF2},
	coninput.VK_F3:        {Sym: KeyF3},
	coninput.VK_F4:        {Sym: KeyF4},
	coninput.VK_F5:        {Sym: KeyF5},
	coninput.VK_F6:        {Sym: KeyF6},
	coninput.VK_F7:        {Sym: KeyF7},
	coninput.VK_F8:        {Sym: KeyF8},
	coninput.VK_F9:        {Sym: KeyF9},
	coninput.VK_F10:       {Sym: KeyF10},
	coninput.VK_F11:       {Sym: KeyF11},
	coninput.VK_F12:       {Sym: KeyF12},
	coninput.VK_F13:       {Sym: KeyF13},
	coninput.VK_F14:       {Sym: KeyF14},
	coninput.VK_F15:       {Sym: KeyF15},
	coninput.VK_F16:       {Sym: KeyF16},
	coninput.VK_F17:       {Sym: KeyF17},
	coninput.VK_F18:       {Sym: KeyF18},
	coninput.VK_F19:       {Sym: KeyF19},
	coninput.VK_F20:       {Sym: KeyF20},
	coninput.VK_F21:       {Sym: KeyF21},
	coninput.VK_F22:       {Sym: KeyF22},
	coninput.VK_F23:       {Sym: KeyF23},
	coninput.VK_F24:       {Sym: KeyF24},
	coninput.VK_NUMLOCK:   {Sym: KeyNumLock},
	coninput.VK_SCROLL:    {Sym: KeyScrollLock},
	coninput.VK_LSHIFT:    {Sym: KeyLeftShift},
	coninput.VK_RSHIFT:    {Sym: KeyRightShift},
	coninput.VK_LCONTROL:  {Sym: KeyLeftCtrl},
	coninput.VK_RCONTROL:  {Sym: KeyRightCtrl},
	coninput.VK_LMENU:     {Sym: KeyLeftAlt},
	coninput.VK_RMENU:     {Sym: KeyRightAlt},
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
