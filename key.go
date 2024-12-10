package tea

import (
	"fmt"

	"github.com/charmbracelet/x/input"
)

const (
	// KeyExtended is a special key code used to signify that a key event
	// contains multiple runes.
	KeyExtended = input.KeyExtended
)

// Special key symbols.
const (

	// Special keys

	KeyUp     = input.KeyUp
	KeyDown   = input.KeyDown
	KeyRight  = input.KeyRight
	KeyLeft   = input.KeyLeft
	KeyBegin  = input.KeyBegin
	KeyFind   = input.KeyFind
	KeyInsert = input.KeyInsert
	KeyDelete = input.KeyDelete
	KeySelect = input.KeySelect
	KeyPgUp   = input.KeyPgUp
	KeyPgDown = input.KeyPgDown
	KeyHome   = input.KeyHome
	KeyEnd    = input.KeyEnd

	// Keypad keys

	KeyKpEnter    = input.KeyKpEnter
	KeyKpEqual    = input.KeyKpEqual
	KeyKpMultiply = input.KeyKpMultiply
	KeyKpPlus     = input.KeyKpPlus
	KeyKpComma    = input.KeyKpComma
	KeyKpMinus    = input.KeyKpMinus
	KeyKpDecimal  = input.KeyKpDecimal
	KeyKpDivide   = input.KeyKpDivide
	KeyKp0        = input.KeyKp0
	KeyKp1        = input.KeyKp1
	KeyKp2        = input.KeyKp2
	KeyKp3        = input.KeyKp3
	KeyKp4        = input.KeyKp4
	KeyKp5        = input.KeyKp5
	KeyKp6        = input.KeyKp6
	KeyKp7        = input.KeyKp7
	KeyKp8        = input.KeyKp8
	KeyKp9        = input.KeyKp9

	// The following are keys defined in the Kitty keyboard protocol.
	// TODO: Investigate the names of these keys
	KeyKpSep    = input.KeyKpSep
	KeyKpUp     = input.KeyKpUp
	KeyKpDown   = input.KeyKpDown
	KeyKpLeft   = input.KeyKpLeft
	KeyKpRight  = input.KeyKpRight
	KeyKpPgUp   = input.KeyKpPgUp
	KeyKpPgDown = input.KeyKpPgDown
	KeyKpHome   = input.KeyKpHome
	KeyKpEnd    = input.KeyKpEnd
	KeyKpInsert = input.KeyKpInsert
	KeyKpDelete = input.KeyKpDelete
	KeyKpBegin  = input.KeyKpBegin

	// Function keys

	KeyF1  = input.KeyF1
	KeyF2  = input.KeyF2
	KeyF3  = input.KeyF3
	KeyF4  = input.KeyF4
	KeyF5  = input.KeyF5
	KeyF6  = input.KeyF6
	KeyF7  = input.KeyF7
	KeyF8  = input.KeyF8
	KeyF9  = input.KeyF9
	KeyF10 = input.KeyF10
	KeyF11 = input.KeyF11
	KeyF12 = input.KeyF12
	KeyF13 = input.KeyF13
	KeyF14 = input.KeyF14
	KeyF15 = input.KeyF15
	KeyF16 = input.KeyF16
	KeyF17 = input.KeyF17
	KeyF18 = input.KeyF18
	KeyF19 = input.KeyF19
	KeyF20 = input.KeyF20
	KeyF21 = input.KeyF21
	KeyF22 = input.KeyF22
	KeyF23 = input.KeyF23
	KeyF24 = input.KeyF24
	KeyF25 = input.KeyF25
	KeyF26 = input.KeyF26
	KeyF27 = input.KeyF27
	KeyF28 = input.KeyF28
	KeyF29 = input.KeyF29
	KeyF30 = input.KeyF30
	KeyF31 = input.KeyF31
	KeyF32 = input.KeyF32
	KeyF33 = input.KeyF33
	KeyF34 = input.KeyF34
	KeyF35 = input.KeyF35
	KeyF36 = input.KeyF36
	KeyF37 = input.KeyF37
	KeyF38 = input.KeyF38
	KeyF39 = input.KeyF39
	KeyF40 = input.KeyF40
	KeyF41 = input.KeyF41
	KeyF42 = input.KeyF42
	KeyF43 = input.KeyF43
	KeyF44 = input.KeyF44
	KeyF45 = input.KeyF45
	KeyF46 = input.KeyF46
	KeyF47 = input.KeyF47
	KeyF48 = input.KeyF48
	KeyF49 = input.KeyF49
	KeyF50 = input.KeyF50
	KeyF51 = input.KeyF51
	KeyF52 = input.KeyF52
	KeyF53 = input.KeyF53
	KeyF54 = input.KeyF54
	KeyF55 = input.KeyF55
	KeyF56 = input.KeyF56
	KeyF57 = input.KeyF57
	KeyF58 = input.KeyF58
	KeyF59 = input.KeyF59
	KeyF60 = input.KeyF60
	KeyF61 = input.KeyF61
	KeyF62 = input.KeyF62
	KeyF63 = input.KeyF63

	// The following are keys defined in the Kitty keyboard protocol.
	// TODO: Investigate the names of these keys

	KeyCapsLock    = input.KeyCapsLock
	KeyScrollLock  = input.KeyScrollLock
	KeyNumLock     = input.KeyNumLock
	KeyPrintScreen = input.KeyPrintScreen
	KeyPause       = input.KeyPause
	KeyMenu        = input.KeyMenu

	KeyMediaPlay        = input.KeyMediaPlay
	KeyMediaPause       = input.KeyMediaPause
	KeyMediaPlayPause   = input.KeyMediaPlayPause
	KeyMediaReverse     = input.KeyMediaReverse
	KeyMediaStop        = input.KeyMediaStop
	KeyMediaFastForward = input.KeyMediaFastForward
	KeyMediaRewind      = input.KeyMediaRewind
	KeyMediaNext        = input.KeyMediaNext
	KeyMediaPrev        = input.KeyMediaPrev
	KeyMediaRecord

	KeyLowerVol = input.KeyLowerVol
	KeyRaiseVol = input.KeyRaiseVol
	KeyMute     = input.KeyMute

	KeyLeftShift      = input.KeyLeftShift
	KeyLeftAlt        = input.KeyLeftAlt
	KeyLeftCtrl       = input.KeyLeftCtrl
	KeyLeftSuper      = input.KeyLeftSuper
	KeyLeftHyper      = input.KeyLeftHyper
	KeyLeftMeta       = input.KeyLeftMeta
	KeyRightShift     = input.KeyRightShift
	KeyRightAlt       = input.KeyRightAlt
	KeyRightCtrl      = input.KeyRightCtrl
	KeyRightSuper     = input.KeyRightSuper
	KeyRightHyper     = input.KeyRightHyper
	KeyRightMeta      = input.KeyRightMeta
	KeyIsoLevel3Shift = input.KeyIsoLevel3Shift
	KeyIsoLevel5Shift = input.KeyIsoLevel5Shift

	// Special names in C0

	KeyBackspace = input.KeyBackspace
	KeyTab       = input.KeyTab
	KeyEnter     = input.KeyEnter
	KeyReturn    = input.KeyReturn
	KeyEscape    = input.KeyEscape
	KeyEsc       = input.KeyEsc

	// Special names in G0

	KeySpace = input.KeySpace
)

// KeyPressMsg represents a key press message.
type KeyPressMsg Key

// String implements [fmt.Stringer] and is quite useful for matching key
// events. For details, on what this returns see [Key.String].
func (k KeyPressMsg) String() string {
	return Key(k).String()
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

// String implements [fmt.Stringer] and is used to convert a key to a string.
// While less type safe than looking at the individual fields, it will usually
// be more convenient and readable to use this method when matching against
// keys.
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
func (k Key) String() string {
	return input.Key(k).String()
}
