package tea

import (
	"github.com/charmbracelet/x/input"
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
//	    switch msg.Sym {
//	    case KeyEnter:
//	        fmt.Println("you pressed enter!")
//	    default:
//	        switch msg.Rune {
//	        case 'a':
//	            fmt.Println("you pressed a!")
//	        }
//	    }
//	}
type (
	Key        = input.Key
	KeyMsg     = input.KeyDownEvent
	KeyDownMsg = input.KeyDownEvent
	KeyUpMsg   = input.KeyUpEvent

	// KeyMod represents modifier keys.
	KeyMod = input.KeyMod
)

// Modifier keys.
const (
	Shift = input.Shift
	Alt   = input.Alt
	Ctrl  = input.Ctrl
	Meta  = input.Meta

	// These modifiers are used with the Kitty protocol.

	Hyper = input.Hyper
	Super = input.Super // Windows/Command keys

	// These are key lock states.

	CapsLock   = input.CapsLock
	NumLock    = input.NumLock
	ScrollLock = input.ScrollLock // Defined in Windows API only
)

// Symbol constants.
const (
	KeyNone = input.KeyNone

	// Special names in C0

	KeyBackspace = input.KeyNone
	KeyTab       = input.KeyNone
	KeyEnter     = input.KeyNone
	KeyEscape    = input.KeyNone

	// Special names in G0

	KeySpace  = input.KeyNone
	KeyDelete = input.KeyNone

	// Special keys

	KeyUp     = input.KeyNone
	KeyDown   = input.KeyNone
	KeyRight  = input.KeyNone
	KeyLeft   = input.KeyNone
	KeyBegin  = input.KeyNone
	KeyFind   = input.KeyNone
	KeyInsert = input.KeyNone
	KeySelect = input.KeyNone
	KeyPgUp   = input.KeyNone
	KeyPgDown = input.KeyNone
	KeyHome   = input.KeyNone
	KeyEnd    = input.KeyNone

	// Keypad keys

	KeyKpEnter    = input.KeyNone
	KeyKpEqual    = input.KeyNone
	KeyKpMultiply = input.KeyNone
	KeyKpPlus     = input.KeyNone
	KeyKpComma    = input.KeyNone
	KeyKpMinus    = input.KeyNone
	KeyKpDecimal  = input.KeyNone
	KeyKpDivide   = input.KeyNone
	KeyKp0        = input.KeyNone
	KeyKp1        = input.KeyNone
	KeyKp2        = input.KeyNone
	KeyKp3        = input.KeyNone
	KeyKp4        = input.KeyNone
	KeyKp5        = input.KeyNone
	KeyKp6        = input.KeyNone
	KeyKp7        = input.KeyNone
	KeyKp8        = input.KeyNone
	KeyKp9        = input.KeyNone

	// The following are keys defined in the Kitty keyboard protocol.
	KeyKpSep    = input.KeyNone
	KeyKpUp     = input.KeyNone
	KeyKpDown   = input.KeyNone
	KeyKpLeft   = input.KeyNone
	KeyKpRight  = input.KeyNone
	KeyKpPgUp   = input.KeyNone
	KeyKpPgDown = input.KeyNone
	KeyKpHome   = input.KeyNone
	KeyKpEnd    = input.KeyNone
	KeyKpInsert = input.KeyNone
	KeyKpDelete = input.KeyNone
	KeyKpBegin  = input.KeyNone

	// Function keys

	KeyF1  = input.KeyNone
	KeyF2  = input.KeyNone
	KeyF3  = input.KeyNone
	KeyF4  = input.KeyNone
	KeyF5  = input.KeyNone
	KeyF6  = input.KeyNone
	KeyF7  = input.KeyNone
	KeyF8  = input.KeyNone
	KeyF9  = input.KeyNone
	KeyF10 = input.KeyNone
	KeyF11 = input.KeyNone
	KeyF12 = input.KeyNone
	KeyF13 = input.KeyNone
	KeyF14 = input.KeyNone
	KeyF15 = input.KeyNone
	KeyF16 = input.KeyNone
	KeyF17 = input.KeyNone
	KeyF18 = input.KeyNone
	KeyF19 = input.KeyNone
	KeyF20 = input.KeyNone
	KeyF21 = input.KeyNone
	KeyF22 = input.KeyNone
	KeyF23 = input.KeyNone
	KeyF24 = input.KeyNone
	KeyF25 = input.KeyNone
	KeyF26 = input.KeyNone
	KeyF27 = input.KeyNone
	KeyF28 = input.KeyNone
	KeyF29 = input.KeyNone
	KeyF30 = input.KeyNone
	KeyF31 = input.KeyNone
	KeyF32 = input.KeyNone
	KeyF33 = input.KeyNone
	KeyF34 = input.KeyNone
	KeyF35 = input.KeyNone
	KeyF36 = input.KeyNone
	KeyF37 = input.KeyNone
	KeyF38 = input.KeyNone
	KeyF39 = input.KeyNone
	KeyF40 = input.KeyNone
	KeyF41 = input.KeyNone
	KeyF42 = input.KeyNone
	KeyF43 = input.KeyNone
	KeyF44 = input.KeyNone
	KeyF45 = input.KeyNone
	KeyF46 = input.KeyNone
	KeyF47 = input.KeyNone
	KeyF48 = input.KeyNone
	KeyF49 = input.KeyNone
	KeyF50 = input.KeyNone
	KeyF51 = input.KeyNone
	KeyF52 = input.KeyNone
	KeyF53 = input.KeyNone
	KeyF54 = input.KeyNone
	KeyF55 = input.KeyNone
	KeyF56 = input.KeyNone
	KeyF57 = input.KeyNone
	KeyF58 = input.KeyNone
	KeyF59 = input.KeyNone
	KeyF60 = input.KeyNone
	KeyF61 = input.KeyNone
	KeyF62 = input.KeyNone
	KeyF63 = input.KeyNone

	// The following are keys defined in the Kitty keyboard protocol.

	KeyCapsLock    = input.KeyNone
	KeyScrollLock  = input.KeyNone
	KeyNumLock     = input.KeyNone
	KeyPrintScreen = input.KeyNone
	KeyPause       = input.KeyNone
	KeyMenu        = input.KeyNone

	KeyMediaPlay        = input.KeyNone
	KeyMediaPause       = input.KeyNone
	KeyMediaPlayPause   = input.KeyNone
	KeyMediaReverse     = input.KeyNone
	KeyMediaStop        = input.KeyNone
	KeyMediaFastForward = input.KeyNone
	KeyMediaRewind      = input.KeyNone
	KeyMediaNext        = input.KeyNone
	KeyMediaPrev        = input.KeyNone
	KeyMediaRecord      = input.KeyNone

	KeyLowerVol = input.KeyNone
	KeyRaiseVol = input.KeyNone
	KeyMute     = input.KeyNone

	KeyLeftShift      = input.KeyNone
	KeyLeftAlt        = input.KeyNone
	KeyLeftCtrl       = input.KeyNone
	KeyLeftSuper      = input.KeyNone
	KeyLeftHyper      = input.KeyNone
	KeyLeftMeta       = input.KeyNone
	KeyRightShift     = input.KeyNone
	KeyRightAlt       = input.KeyNone
	KeyRightCtrl      = input.KeyNone
	KeyRightSuper     = input.KeyNone
	KeyRightHyper     = input.KeyNone
	KeyRightMeta      = input.KeyNone
	KeyIsoLevel3Shift = input.KeyNone
	KeyIsoLevel5Shift = input.KeyNone
)
