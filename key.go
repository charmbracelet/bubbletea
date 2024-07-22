package tea

import (
	"github.com/charmbracelet/x/input"
)

// Key msgs contains information about a keypress and keyrelease. KeyPressMsgs are
// always sent to the program's update function. There are a couple general
// patterns you could use to check for keypresses:
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
type (
	Key           input.Key
	KeyPressMsg   input.KeyPressEvent
	KeyReleaseMsg input.KeyReleaseEvent

	// Deprecated: Use KeyPressMsg instead.
	KeyMsg KeyPressMsg
)

// String returns a string representation for a key. It's safe (and encouraged)
// for use in key comparison.
//
// Deprecated: Use KeyPressMsg.String instead.
func (k KeyMsg) String() (str string) {
	return input.KeyPressEvent(k).String()
}

// String returns a string representation for a key. It's safe (and encouraged)
// for use in key comparison.
func (k Key) String() (str string) {
	return input.Key(k).String()
}

// String returns a string representation for a key message. It's safe (and
// encouraged) for use in key comparison.
func (k KeyPressMsg) String() (str string) {
	return input.KeyPressEvent(k).String()
}

// String returns a string representation for a key message. It's safe (and
// encouraged) for use in key comparison.
func (k KeyReleaseMsg) String() (str string) {
	return input.KeyReleaseEvent(k).String()
}

// Symbol constants.
const (
	KeyNone             = input.KeyNone
	KeyBackspace        = input.KeyBackspace
	KeyTab              = input.KeyTab
	KeyEnter            = input.KeyEnter
	KeyEscape           = input.KeyEscape
	KeySpace            = input.KeySpace
	KeyDelete           = input.KeyDelete
	KeyUp               = input.KeyUp
	KeyDown             = input.KeyDown
	KeyRight            = input.KeyRight
	KeyLeft             = input.KeyLeft
	KeyBegin            = input.KeyBegin
	KeyFind             = input.KeyFind
	KeyInsert           = input.KeyInsert
	KeySelect           = input.KeySelect
	KeyPgUp             = input.KeyPgUp
	KeyPgDown           = input.KeyPgDown
	KeyHome             = input.KeyHome
	KeyEnd              = input.KeyEnd
	KeyKpEnter          = input.KeyKpEnter
	KeyKpEqual          = input.KeyKpEqual
	KeyKpMultiply       = input.KeyKpMultiply
	KeyKpPlus           = input.KeyKpPlus
	KeyKpComma          = input.KeyKpComma
	KeyKpMinus          = input.KeyKpMinus
	KeyKpDecimal        = input.KeyKpDecimal
	KeyKpDivide         = input.KeyKpDivide
	KeyKp0              = input.KeyKp0
	KeyKp1              = input.KeyKp1
	KeyKp2              = input.KeyKp2
	KeyKp3              = input.KeyKp3
	KeyKp4              = input.KeyKp4
	KeyKp5              = input.KeyKp5
	KeyKp6              = input.KeyKp6
	KeyKp7              = input.KeyKp7
	KeyKp8              = input.KeyKp8
	KeyKp9              = input.KeyKp9
	KeyKpSep            = input.KeyKpSep
	KeyKpUp             = input.KeyKpUp
	KeyKpDown           = input.KeyKpDown
	KeyKpLeft           = input.KeyKpLeft
	KeyKpRight          = input.KeyKpRight
	KeyKpPgUp           = input.KeyKpPgUp
	KeyKpPgDown         = input.KeyKpPgDown
	KeyKpHome           = input.KeyKpHome
	KeyKpEnd            = input.KeyKpEnd
	KeyKpInsert         = input.KeyKpInsert
	KeyKpDelete         = input.KeyKpDelete
	KeyKpBegin          = input.KeyKpBegin
	KeyF1               = input.KeyF1
	KeyF2               = input.KeyF2
	KeyF3               = input.KeyF3
	KeyF4               = input.KeyF4
	KeyF5               = input.KeyF5
	KeyF6               = input.KeyF6
	KeyF7               = input.KeyF7
	KeyF8               = input.KeyF8
	KeyF9               = input.KeyF9
	KeyF10              = input.KeyF10
	KeyF11              = input.KeyF11
	KeyF12              = input.KeyF12
	KeyF13              = input.KeyF13
	KeyF14              = input.KeyF14
	KeyF15              = input.KeyF15
	KeyF16              = input.KeyF16
	KeyF17              = input.KeyF17
	KeyF18              = input.KeyF18
	KeyF19              = input.KeyF19
	KeyF20              = input.KeyF20
	KeyF21              = input.KeyF21
	KeyF22              = input.KeyF22
	KeyF23              = input.KeyF23
	KeyF24              = input.KeyF24
	KeyF25              = input.KeyF25
	KeyF26              = input.KeyF26
	KeyF27              = input.KeyF27
	KeyF28              = input.KeyF28
	KeyF29              = input.KeyF29
	KeyF30              = input.KeyF30
	KeyF31              = input.KeyF31
	KeyF32              = input.KeyF32
	KeyF33              = input.KeyF33
	KeyF34              = input.KeyF34
	KeyF35              = input.KeyF35
	KeyF36              = input.KeyF36
	KeyF37              = input.KeyF37
	KeyF38              = input.KeyF38
	KeyF39              = input.KeyF39
	KeyF40              = input.KeyF40
	KeyF41              = input.KeyF41
	KeyF42              = input.KeyF42
	KeyF43              = input.KeyF43
	KeyF44              = input.KeyF44
	KeyF45              = input.KeyF45
	KeyF46              = input.KeyF46
	KeyF47              = input.KeyF47
	KeyF48              = input.KeyF48
	KeyF49              = input.KeyF49
	KeyF50              = input.KeyF50
	KeyF51              = input.KeyF51
	KeyF52              = input.KeyF52
	KeyF53              = input.KeyF53
	KeyF54              = input.KeyF54
	KeyF55              = input.KeyF55
	KeyF56              = input.KeyF56
	KeyF57              = input.KeyF57
	KeyF58              = input.KeyF58
	KeyF59              = input.KeyF59
	KeyF60              = input.KeyF60
	KeyF61              = input.KeyF61
	KeyF62              = input.KeyF62
	KeyF63              = input.KeyF63
	KeyCapsLock         = input.KeyCapsLock
	KeyScrollLock       = input.KeyScrollLock
	KeyNumLock          = input.KeyNumLock
	KeyPrintScreen      = input.KeyPrintScreen
	KeyPause            = input.KeyPause
	KeyMenu             = input.KeyMenu
	KeyMediaPlay        = input.KeyMediaPlay
	KeyMediaPause       = input.KeyMediaPause
	KeyMediaPlayPause   = input.KeyMediaPlayPause
	KeyMediaReverse     = input.KeyMediaReverse
	KeyMediaStop        = input.KeyMediaStop
	KeyMediaFastForward = input.KeyMediaFastForward
	KeyMediaRewind      = input.KeyMediaRewind
	KeyMediaNext        = input.KeyMediaNext
	KeyMediaPrev        = input.KeyMediaPrev
	KeyMediaRecord      = input.KeyMediaRecord
	KeyLowerVol         = input.KeyLowerVol
	KeyRaiseVol         = input.KeyRaiseVol
	KeyMute             = input.KeyMute
	KeyLeftShift        = input.KeyLeftShift
	KeyLeftAlt          = input.KeyLeftAlt
	KeyLeftCtrl         = input.KeyLeftCtrl
	KeyLeftSuper        = input.KeyLeftSuper
	KeyLeftHyper        = input.KeyLeftHyper
	KeyLeftMeta         = input.KeyLeftMeta
	KeyRightShift       = input.KeyRightShift
	KeyRightAlt         = input.KeyRightAlt
	KeyRightCtrl        = input.KeyRightCtrl
	KeyRightSuper       = input.KeyRightSuper
	KeyRightHyper       = input.KeyRightHyper
	KeyRightMeta        = input.KeyRightMeta
	KeyIsoLevel3Shift   = input.KeyIsoLevel3Shift
	KeyIsoLevel5Shift   = input.KeyIsoLevel5Shift
)
