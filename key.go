package tea

import (
	"fmt"

	"github.com/charmbracelet/tv"
)

const (
	// KeyExtended is a special key code used to signify that a key event
	// contains multiple runes.
	KeyExtended = tv.KeyExtended
)

// Special key symbols.
const (

	// Special keys.

	KeyUp     = tv.KeyUp
	KeyDown   = tv.KeyDown
	KeyRight  = tv.KeyRight
	KeyLeft   = tv.KeyLeft
	KeyBegin  = tv.KeyBegin
	KeyFind   = tv.KeyFind
	KeyInsert = tv.KeyInsert
	KeyDelete = tv.KeyDelete
	KeySelect = tv.KeySelect
	KeyPgUp   = tv.KeyPgUp
	KeyPgDown = tv.KeyPgDown
	KeyHome   = tv.KeyHome
	KeyEnd    = tv.KeyEnd

	// Keypad keys.

	KeyKpEnter    = tv.KeyKpEnter
	KeyKpEqual    = tv.KeyKpEqual
	KeyKpMultiply = tv.KeyKpMultiply
	KeyKpPlus     = tv.KeyKpPlus
	KeyKpComma    = tv.KeyKpComma
	KeyKpMinus    = tv.KeyKpMinus
	KeyKpDecimal  = tv.KeyKpDecimal
	KeyKpDivide   = tv.KeyKpDivide
	KeyKp0        = tv.KeyKp0
	KeyKp1        = tv.KeyKp1
	KeyKp2        = tv.KeyKp2
	KeyKp3        = tv.KeyKp3
	KeyKp4        = tv.KeyKp4
	KeyKp5        = tv.KeyKp5
	KeyKp6        = tv.KeyKp6
	KeyKp7        = tv.KeyKp7
	KeyKp8        = tv.KeyKp8
	KeyKp9        = tv.KeyKp9

	// The following are keys defined in the Kitty keyboard protocol.
	// XXX: Investigate the names of these keys.
	KeyKpSep    = tv.KeyKpSep
	KeyKpUp     = tv.KeyKpUp
	KeyKpDown   = tv.KeyKpDown
	KeyKpLeft   = tv.KeyKpLeft
	KeyKpRight  = tv.KeyKpRight
	KeyKpPgUp   = tv.KeyKpPgUp
	KeyKpPgDown = tv.KeyKpPgDown
	KeyKpHome   = tv.KeyKpHome
	KeyKpEnd    = tv.KeyKpEnd
	KeyKpInsert = tv.KeyKpInsert
	KeyKpDelete = tv.KeyKpDelete
	KeyKpBegin  = tv.KeyKpBegin

	// Function keys.

	KeyF1  = tv.KeyF1
	KeyF2  = tv.KeyF2
	KeyF3  = tv.KeyF3
	KeyF4  = tv.KeyF4
	KeyF5  = tv.KeyF5
	KeyF6  = tv.KeyF6
	KeyF7  = tv.KeyF7
	KeyF8  = tv.KeyF8
	KeyF9  = tv.KeyF9
	KeyF10 = tv.KeyF10
	KeyF11 = tv.KeyF11
	KeyF12 = tv.KeyF12
	KeyF13 = tv.KeyF13
	KeyF14 = tv.KeyF14
	KeyF15 = tv.KeyF15
	KeyF16 = tv.KeyF16
	KeyF17 = tv.KeyF17
	KeyF18 = tv.KeyF18
	KeyF19 = tv.KeyF19
	KeyF20 = tv.KeyF20
	KeyF21 = tv.KeyF21
	KeyF22 = tv.KeyF22
	KeyF23 = tv.KeyF23
	KeyF24 = tv.KeyF24
	KeyF25 = tv.KeyF25
	KeyF26 = tv.KeyF26
	KeyF27 = tv.KeyF27
	KeyF28 = tv.KeyF28
	KeyF29 = tv.KeyF29
	KeyF30 = tv.KeyF30
	KeyF31 = tv.KeyF31
	KeyF32 = tv.KeyF32
	KeyF33 = tv.KeyF33
	KeyF34 = tv.KeyF34
	KeyF35 = tv.KeyF35
	KeyF36 = tv.KeyF36
	KeyF37 = tv.KeyF37
	KeyF38 = tv.KeyF38
	KeyF39 = tv.KeyF39
	KeyF40 = tv.KeyF40
	KeyF41 = tv.KeyF41
	KeyF42 = tv.KeyF42
	KeyF43 = tv.KeyF43
	KeyF44 = tv.KeyF44
	KeyF45 = tv.KeyF45
	KeyF46 = tv.KeyF46
	KeyF47 = tv.KeyF47
	KeyF48 = tv.KeyF48
	KeyF49 = tv.KeyF49
	KeyF50 = tv.KeyF50
	KeyF51 = tv.KeyF51
	KeyF52 = tv.KeyF52
	KeyF53 = tv.KeyF53
	KeyF54 = tv.KeyF54
	KeyF55 = tv.KeyF55
	KeyF56 = tv.KeyF56
	KeyF57 = tv.KeyF57
	KeyF58 = tv.KeyF58
	KeyF59 = tv.KeyF59
	KeyF60 = tv.KeyF60
	KeyF61 = tv.KeyF61
	KeyF62 = tv.KeyF62
	KeyF63 = tv.KeyF63

	// The following are keys defined in the Kitty keyboard protocol.
	// XXX: Investigate the names of these keys.

	KeyCapsLock    = tv.KeyCapsLock
	KeyScrollLock  = tv.KeyScrollLock
	KeyNumLock     = tv.KeyNumLock
	KeyPrintScreen = tv.KeyPrintScreen
	KeyPause       = tv.KeyPause
	KeyMenu        = tv.KeyMenu

	KeyMediaPlay        = tv.KeyMediaPlay
	KeyMediaPause       = tv.KeyMediaPause
	KeyMediaPlayPause   = tv.KeyMediaPlayPause
	KeyMediaReverse     = tv.KeyMediaReverse
	KeyMediaStop        = tv.KeyMediaStop
	KeyMediaFastForward = tv.KeyMediaFastForward
	KeyMediaRewind      = tv.KeyMediaRewind
	KeyMediaNext        = tv.KeyMediaNext
	KeyMediaPrev        = tv.KeyMediaPrev
	KeyMediaRecord

	KeyLowerVol = tv.KeyLowerVol
	KeyRaiseVol = tv.KeyRaiseVol
	KeyMute     = tv.KeyMute

	KeyLeftShift      = tv.KeyLeftShift
	KeyLeftAlt        = tv.KeyLeftAlt
	KeyLeftCtrl       = tv.KeyLeftCtrl
	KeyLeftSuper      = tv.KeyLeftSuper
	KeyLeftHyper      = tv.KeyLeftHyper
	KeyLeftMeta       = tv.KeyLeftMeta
	KeyRightShift     = tv.KeyRightShift
	KeyRightAlt       = tv.KeyRightAlt
	KeyRightCtrl      = tv.KeyRightCtrl
	KeyRightSuper     = tv.KeyRightSuper
	KeyRightHyper     = tv.KeyRightHyper
	KeyRightMeta      = tv.KeyRightMeta
	KeyIsoLevel3Shift = tv.KeyIsoLevel3Shift
	KeyIsoLevel5Shift = tv.KeyIsoLevel5Shift

	// Special names in C0.

	KeyBackspace = tv.KeyBackspace
	KeyTab       = tv.KeyTab
	KeyEnter     = tv.KeyEnter
	KeyReturn    = tv.KeyReturn
	KeyEscape    = tv.KeyEscape
	KeyEsc       = tv.KeyEsc

	// Special names in G0.

	KeySpace = tv.KeySpace
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
	return tv.Key(k).Keystroke()
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
	return tv.Key(k).Keystroke()
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
	return tv.Key(k).String()
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
	return tv.Key(k).Keystroke()
}
