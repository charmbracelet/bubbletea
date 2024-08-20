package tea

import (
	"strings"
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
//	    switch msg.Type {
//	    case KeyEnter:
//	        fmt.Println("you pressed enter!")
//	    case KeyRunes:
//	        switch string(msg.Runes) {
//	        case "a":
//	            fmt.Println("you pressed a!")
//	        }
//	    }
//	}
//
// Note that Key.Runes will always contain at least one character, so you can
// always safely call Key.Runes[0]. In most cases Key.Runes will only contain
// one character, though certain input method editors (most notably Chinese
// IMEs) can input multiple runes at once.
//
// TODO(v2): Add a KeyMsg interface that incorporates all the key message
// types.
//
// Deprecated: KeyMsg is deprecated in favor of KeyPressMsg and KeyReleaseMsg.
type KeyMsg struct {
	Type  KeyType
	Runes []rune
	Alt   bool
	Paste bool
}

// String returns a friendly string representation for a key. It's safe (and
// encouraged) for use in key comparison.
//
//	k := Key{Type: KeyEnter}
//	fmt.Println(k)
//	// Output: enter
func (k KeyMsg) String() (str string) {
	var buf strings.Builder
	if k.Alt {
		buf.WriteString("alt+")
	}
	if k.Type == KeyRunes {
		if k.Paste {
			// Note: bubbles/keys bindings currently do string compares to
			// recognize shortcuts. Since pasted text should never activate
			// shortcuts, we need to ensure that the binding code doesn't
			// match Key events that result from pastes. We achieve this
			// here by enclosing pastes in '[...]' so that the string
			// comparison in Matches() fails in that case.
			buf.WriteByte('[')
		}
		buf.WriteString(string(k.Runes))
		if k.Paste {
			buf.WriteByte(']')
		}
		return buf.String()
	} else if s, ok := keyNames[k.Type]; ok {
		buf.WriteString(s)
		return buf.String()
	}
	return ""
}

// Control key aliases.
const (
	KeyNull KeyType = -iota - 10
	KeyBreak

	KeyCtrlAt // ctrl+@
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlH
	KeyCtrlI
	KeyCtrlJ
	KeyCtrlK
	KeyCtrlL
	KeyCtrlM
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ
	KeyCtrlOpenBracket  // ctrl+[
	KeyCtrlBackslash    // ctrl+\
	KeyCtrlCloseBracket // ctrl+]
	KeyCtrlCaret        // ctrl+^
	KeyCtrlUnderscore   // ctrl+_
	KeyCtrlQuestionMark // ctrl+?
	KeyCtrlUp
	KeyCtrlDown
	KeyCtrlRight
	KeyCtrlLeft
	KeyCtrlPgUp
	KeyCtrlPgDown
	KeyCtrlHome
	KeyCtrlEnd

	KeyShiftTab
	KeyShiftUp
	KeyShiftDown
	KeyShiftRight
	KeyShiftLeft
	KeyShiftHome
	KeyShiftEnd

	KeyCtrlShiftUp
	KeyCtrlShiftDown
	KeyCtrlShiftLeft
	KeyCtrlShiftRight
	KeyCtrlShiftHome
	KeyCtrlShiftEnd

	// Deprecated: Use KeyEscape instead.
	KeyEsc = KeyEscape
)

// Mappings for control keys and other special keys to friendly consts.
var keyNames = map[KeyType]string{
	// Control keys.
	KeyCtrlAt:           "ctrl+@", // also ctrl+` (that's ctrl+backtick)
	KeyCtrlA:            "ctrl+a",
	KeyCtrlB:            "ctrl+b",
	KeyCtrlC:            "ctrl+c",
	KeyCtrlD:            "ctrl+d",
	KeyCtrlE:            "ctrl+e",
	KeyCtrlF:            "ctrl+f",
	KeyCtrlG:            "ctrl+g",
	KeyCtrlH:            "ctrl+h",
	KeyTab:              "tab", // also ctrl+i
	KeyCtrlJ:            "ctrl+j",
	KeyCtrlK:            "ctrl+k",
	KeyCtrlL:            "ctrl+l",
	KeyEnter:            "enter",
	KeyCtrlN:            "ctrl+n",
	KeyCtrlO:            "ctrl+o",
	KeyCtrlP:            "ctrl+p",
	KeyCtrlQ:            "ctrl+q",
	KeyCtrlR:            "ctrl+r",
	KeyCtrlS:            "ctrl+s",
	KeyCtrlT:            "ctrl+t",
	KeyCtrlU:            "ctrl+u",
	KeyCtrlV:            "ctrl+v",
	KeyCtrlW:            "ctrl+w",
	KeyCtrlX:            "ctrl+x",
	KeyCtrlY:            "ctrl+y",
	KeyCtrlZ:            "ctrl+z",
	KeyEscape:           "esc",
	KeyCtrlOpenBracket:  "ctrl+[",
	KeyCtrlBackslash:    "ctrl+\\",
	KeyCtrlCloseBracket: "ctrl+]",
	KeyCtrlCaret:        "ctrl+^",
	KeyCtrlUnderscore:   "ctrl+_",
	KeyBackspace:        "backspace",

	// Other keys.
	KeyRunes:          "runes",
	KeyUp:             "up",
	KeyDown:           "down",
	KeyRight:          "right",
	KeySpace:          " ", // for backwards compatibility
	KeyLeft:           "left",
	KeyShiftTab:       "shift+tab",
	KeyHome:           "home",
	KeyEnd:            "end",
	KeyCtrlHome:       "ctrl+home",
	KeyCtrlEnd:        "ctrl+end",
	KeyShiftHome:      "shift+home",
	KeyShiftEnd:       "shift+end",
	KeyCtrlShiftHome:  "ctrl+shift+home",
	KeyCtrlShiftEnd:   "ctrl+shift+end",
	KeyPgUp:           "pgup",
	KeyPgDown:         "pgdown",
	KeyCtrlPgUp:       "ctrl+pgup",
	KeyCtrlPgDown:     "ctrl+pgdown",
	KeyDelete:         "delete",
	KeyInsert:         "insert",
	KeyCtrlUp:         "ctrl+up",
	KeyCtrlDown:       "ctrl+down",
	KeyCtrlRight:      "ctrl+right",
	KeyCtrlLeft:       "ctrl+left",
	KeyShiftUp:        "shift+up",
	KeyShiftDown:      "shift+down",
	KeyShiftRight:     "shift+right",
	KeyShiftLeft:      "shift+left",
	KeyCtrlShiftUp:    "ctrl+shift+up",
	KeyCtrlShiftDown:  "ctrl+shift+down",
	KeyCtrlShiftLeft:  "ctrl+shift+left",
	KeyCtrlShiftRight: "ctrl+shift+right",
	KeyF1:             "f1",
	KeyF2:             "f2",
	KeyF3:             "f3",
	KeyF4:             "f4",
	KeyF5:             "f5",
	KeyF6:             "f6",
	KeyF7:             "f7",
	KeyF8:             "f8",
	KeyF9:             "f9",
	KeyF10:            "f10",
	KeyF11:            "f11",
	KeyF12:            "f12",
	KeyF13:            "f13",
	KeyF14:            "f14",
	KeyF15:            "f15",
	KeyF16:            "f16",
	KeyF17:            "f17",
	KeyF18:            "f18",
	KeyF19:            "f19",
	KeyF20:            "f20",
}
