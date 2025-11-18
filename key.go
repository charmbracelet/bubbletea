package tea

import (
	"fmt"

	uv "github.com/charmbracelet/ultraviolet"
)

const (
	// KeyExtended is a special key code used to signify that a key event
	// contains multiple runes.
	KeyExtended = uv.KeyExtended
)

// Special key symbols.
const (

	// Special keys.

	KeyUp     = uv.KeyUp
	KeyDown   = uv.KeyDown
	KeyRight  = uv.KeyRight
	KeyLeft   = uv.KeyLeft
	KeyBegin  = uv.KeyBegin
	KeyFind   = uv.KeyFind
	KeyInsert = uv.KeyInsert
	KeyDelete = uv.KeyDelete
	KeySelect = uv.KeySelect
	KeyPgUp   = uv.KeyPgUp
	KeyPgDown = uv.KeyPgDown
	KeyHome   = uv.KeyHome
	KeyEnd    = uv.KeyEnd

	// Keypad keys.

	KeyKpEnter    = uv.KeyKpEnter
	KeyKpEqual    = uv.KeyKpEqual
	KeyKpMultiply = uv.KeyKpMultiply
	KeyKpPlus     = uv.KeyKpPlus
	KeyKpComma    = uv.KeyKpComma
	KeyKpMinus    = uv.KeyKpMinus
	KeyKpDecimal  = uv.KeyKpDecimal
	KeyKpDivide   = uv.KeyKpDivide
	KeyKp0        = uv.KeyKp0
	KeyKp1        = uv.KeyKp1
	KeyKp2        = uv.KeyKp2
	KeyKp3        = uv.KeyKp3
	KeyKp4        = uv.KeyKp4
	KeyKp5        = uv.KeyKp5
	KeyKp6        = uv.KeyKp6
	KeyKp7        = uv.KeyKp7
	KeyKp8        = uv.KeyKp8
	KeyKp9        = uv.KeyKp9

	// The following are keys defined in the Kitty keyboard protocol.
	// XXX: Investigate the names of these keys.
	KeyKpSep    = uv.KeyKpSep
	KeyKpUp     = uv.KeyKpUp
	KeyKpDown   = uv.KeyKpDown
	KeyKpLeft   = uv.KeyKpLeft
	KeyKpRight  = uv.KeyKpRight
	KeyKpPgUp   = uv.KeyKpPgUp
	KeyKpPgDown = uv.KeyKpPgDown
	KeyKpHome   = uv.KeyKpHome
	KeyKpEnd    = uv.KeyKpEnd
	KeyKpInsert = uv.KeyKpInsert
	KeyKpDelete = uv.KeyKpDelete
	KeyKpBegin  = uv.KeyKpBegin

	// Function keys.

	KeyF1  = uv.KeyF1
	KeyF2  = uv.KeyF2
	KeyF3  = uv.KeyF3
	KeyF4  = uv.KeyF4
	KeyF5  = uv.KeyF5
	KeyF6  = uv.KeyF6
	KeyF7  = uv.KeyF7
	KeyF8  = uv.KeyF8
	KeyF9  = uv.KeyF9
	KeyF10 = uv.KeyF10
	KeyF11 = uv.KeyF11
	KeyF12 = uv.KeyF12
	KeyF13 = uv.KeyF13
	KeyF14 = uv.KeyF14
	KeyF15 = uv.KeyF15
	KeyF16 = uv.KeyF16
	KeyF17 = uv.KeyF17
	KeyF18 = uv.KeyF18
	KeyF19 = uv.KeyF19
	KeyF20 = uv.KeyF20
	KeyF21 = uv.KeyF21
	KeyF22 = uv.KeyF22
	KeyF23 = uv.KeyF23
	KeyF24 = uv.KeyF24
	KeyF25 = uv.KeyF25
	KeyF26 = uv.KeyF26
	KeyF27 = uv.KeyF27
	KeyF28 = uv.KeyF28
	KeyF29 = uv.KeyF29
	KeyF30 = uv.KeyF30
	KeyF31 = uv.KeyF31
	KeyF32 = uv.KeyF32
	KeyF33 = uv.KeyF33
	KeyF34 = uv.KeyF34
	KeyF35 = uv.KeyF35
	KeyF36 = uv.KeyF36
	KeyF37 = uv.KeyF37
	KeyF38 = uv.KeyF38
	KeyF39 = uv.KeyF39
	KeyF40 = uv.KeyF40
	KeyF41 = uv.KeyF41
	KeyF42 = uv.KeyF42
	KeyF43 = uv.KeyF43
	KeyF44 = uv.KeyF44
	KeyF45 = uv.KeyF45
	KeyF46 = uv.KeyF46
	KeyF47 = uv.KeyF47
	KeyF48 = uv.KeyF48
	KeyF49 = uv.KeyF49
	KeyF50 = uv.KeyF50
	KeyF51 = uv.KeyF51
	KeyF52 = uv.KeyF52
	KeyF53 = uv.KeyF53
	KeyF54 = uv.KeyF54
	KeyF55 = uv.KeyF55
	KeyF56 = uv.KeyF56
	KeyF57 = uv.KeyF57
	KeyF58 = uv.KeyF58
	KeyF59 = uv.KeyF59
	KeyF60 = uv.KeyF60
	KeyF61 = uv.KeyF61
	KeyF62 = uv.KeyF62
	KeyF63 = uv.KeyF63

	// The following are keys defined in the Kitty keyboard protocol.
	// XXX: Investigate the names of these keys.

	KeyCapsLock    = uv.KeyCapsLock
	KeyScrollLock  = uv.KeyScrollLock
	KeyNumLock     = uv.KeyNumLock
	KeyPrintScreen = uv.KeyPrintScreen
	KeyPause       = uv.KeyPause
	KeyMenu        = uv.KeyMenu

	KeyMediaPlay        = uv.KeyMediaPlay
	KeyMediaPause       = uv.KeyMediaPause
	KeyMediaPlayPause   = uv.KeyMediaPlayPause
	KeyMediaReverse     = uv.KeyMediaReverse
	KeyMediaStop        = uv.KeyMediaStop
	KeyMediaFastForward = uv.KeyMediaFastForward
	KeyMediaRewind      = uv.KeyMediaRewind
	KeyMediaNext        = uv.KeyMediaNext
	KeyMediaPrev        = uv.KeyMediaPrev
	KeyMediaRecord

	KeyLowerVol = uv.KeyLowerVol
	KeyRaiseVol = uv.KeyRaiseVol
	KeyMute     = uv.KeyMute

	KeyLeftShift      = uv.KeyLeftShift
	KeyLeftAlt        = uv.KeyLeftAlt
	KeyLeftCtrl       = uv.KeyLeftCtrl
	KeyLeftSuper      = uv.KeyLeftSuper
	KeyLeftHyper      = uv.KeyLeftHyper
	KeyLeftMeta       = uv.KeyLeftMeta
	KeyRightShift     = uv.KeyRightShift
	KeyRightAlt       = uv.KeyRightAlt
	KeyRightCtrl      = uv.KeyRightCtrl
	KeyRightSuper     = uv.KeyRightSuper
	KeyRightHyper     = uv.KeyRightHyper
	KeyRightMeta      = uv.KeyRightMeta
	KeyIsoLevel3Shift = uv.KeyIsoLevel3Shift
	KeyIsoLevel5Shift = uv.KeyIsoLevel5Shift

	// Special names in C0.

	KeyBackspace = uv.KeyBackspace
	KeyTab       = uv.KeyTab
	KeyEnter     = uv.KeyEnter
	KeyReturn    = uv.KeyReturn
	KeyEscape    = uv.KeyEscape
	KeyEsc       = uv.KeyEsc

	// Special names in G0.

	KeySpace = uv.KeySpace
)

// KeyPressMsg represents a key press message.
type KeyPressMsg Key

// String implements [fmt.Stringer] and is quite useful for matching key
// events. For details, on what this returns see [Key.String].
func (k KeyPressMsg) String() string {
	return Key(k).String()
}

// Keystroke returns the keystroke representation of the [Key]. While less type
// safe than looking at the individual fields, it will usually be more
// convenient and readable to use this method when matching against keys.
//
// Note that modifier keys are always printed in the following order:
//   - ctrl
//   - alt
//   - shift
//   - meta
//   - hyper
//   - super
//
// For example, you'll always see "ctrl+shift+alt+a" and never
// "shift+ctrl+alt+a".
func (k KeyPressMsg) Keystroke() string {
	return uv.Key(k).Keystroke()
}

// Key returns the underlying key event. This is a syntactic sugar for casting
// the key event to a [Key].
func (k KeyPressMsg) Key() Key {
	return Key(k)
}

// KeyReleaseMsg represents a key release message.
type KeyReleaseMsg Key

// String implements [fmt.Stringer] and is quite useful for matching key
// events. For details, on what this returns see [Key.String].
func (k KeyReleaseMsg) String() string {
	return Key(k).String()
}

// Keystroke returns the keystroke representation of the [Key]. While less type
// safe than looking at the individual fields, it will usually be more
// convenient and readable to use this method when matching against keys.
//
// Note that modifier keys are always printed in the following order:
//   - ctrl
//   - alt
//   - shift
//   - meta
//   - hyper
//   - super
//
// For example, you'll always see "ctrl+shift+alt+a" and never
// "shift+ctrl+alt+a".
func (k KeyReleaseMsg) Keystroke() string {
	return uv.Key(k).Keystroke()
}

// Key returns the underlying key event. This is a convenience method and
// syntactic sugar to satisfy the [KeyMsg] interface, and cast the key event to
// [Key].
func (k KeyReleaseMsg) Key() Key {
	return Key(k)
}

// KeyMsg represents a key event. This can be either a key press or a key
// release event.
type KeyMsg interface {
	fmt.Stringer

	// Key returns the underlying key event.
	Key() Key
}

// Key represents a Key press or release event. It contains information about
// the Key pressed, like the runes, the type of Key, and the modifiers pressed.
// There are a couple general patterns you could use to check for key presses
// or releases:
//
//	// Switch on the string representation of the key (shorter)
//	switch msg := msg.(type) {
//	case KeyPressMsg:
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
//	    // catch both KeyPressMsg and KeyReleaseMsg
//	    switch key := msg.Key(); key.Code {
//	    case KeyEnter:
//	        fmt.Println("you pressed enter!")
//	    default:
//	        switch key.Text {
//	        case "a":
//	            fmt.Println("you pressed a!")
//	        }
//	    }
//	}
//
// Note that [Key.Text] will be empty for special keys like [KeyEnter],
// [KeyTab], and for keys that don't represent printable characters like key
// combos with modifier keys. In other words, [Key.Text] is populated only for
// keys that represent printable characters shifted or unshifted (like 'a',
// 'A', '1', '!', etc.).
type Key struct {
	// Text contains the actual characters received. This usually the same as
	// [Key.Code]. When [Key.Text] is non-empty, it indicates that the key
	// pressed represents printable character(s).
	Text string

	// Mod represents modifier keys, like [ModCtrl], [ModAlt], and so on.
	Mod KeyMod

	// Code represents the key pressed. This is usually a special key like
	// [KeyTab], [KeyEnter], [KeyF1], or a printable character like 'a'.
	Code rune

	// ShiftedCode is the actual, shifted key pressed by the user. For example,
	// if the user presses shift+a, or caps lock is on, [Key.ShiftedCode] will
	// be 'A' and [Key.Code] will be 'a'.
	//
	// In the case of non-latin keyboards, like Arabic, [Key.ShiftedCode] is the
	// unshifted key on the keyboard.
	//
	// This is only available with the Kitty Keyboard Protocol or the Windows
	// Console API.
	ShiftedCode rune

	// BaseCode is the key pressed according to the standard PC-101 key layout.
	// On international keyboards, this is the key that would be pressed if the
	// keyboard was set to US PC-101 layout.
	//
	// For example, if the user presses 'q' on a French AZERTY keyboard,
	// [Key.BaseCode] will be 'q'.
	//
	// This is only available with the Kitty Keyboard Protocol or the Windows
	// Console API.
	BaseCode rune

	// IsRepeat indicates whether the key is being held down and sending events
	// repeatedly.
	//
	// This is only available with the Kitty Keyboard Protocol or the Windows
	// Console API.
	IsRepeat bool
}

// String implements [fmt.Stringer] and is quite useful for matching key
// events. It will return the textual representation of the [Key] if there is
// one, otherwise, it will fallback to [Key.Keystroke].
//
// For example, you'll always get "?" and instead of "shift+/" on a US ANSI
// keyboard.
func (k Key) String() string {
	return uv.Key(k).String()
}

// Keystroke returns the keystroke representation of the [Key]. While less type
// safe than looking at the individual fields, it will usually be more
// convenient and readable to use this method when matching against keys.
//
// Note that modifier keys are always printed in the following order:
//   - ctrl
//   - alt
//   - shift
//   - meta
//   - hyper
//   - super
//
// For example, you'll always see "ctrl+shift+alt+a" and never
// "shift+ctrl+alt+a".
func (k Key) Keystroke() string {
	return uv.Key(k).Keystroke()
}
