//go:build windows
// +build windows

package tea

import (
	"fmt"

	"github.com/erikgeiser/coninput"
)

func parseInputMsgsFromInputRecords(events []coninput.InputRecord) ([]Msg, error) {
	allMessages := make([]Msg, 0, len(events))

	for _, event := range events {
		msgs, err := parseInputMsgsFromInputRecord(event)
		if err != nil {
			return msgs, err
		}

		allMessages = append(allMessages, msgs...)
	}

	return allMessages, nil
}

func parseInputMsgsFromInputRecord(event coninput.InputRecord) ([]Msg, error) {
	var msgs []Msg

	switch e := event.Unwrap().(type) {
	case coninput.KeyEventRecord:
		if !e.KeyDown || e.VirtualKeyCode == coninput.VK_SHIFT {
			return nil, nil
		}

		msgs := make([]Msg, 0, e.RepeatCount)

		for i := 0; i < int(e.RepeatCount); i++ {
			msgs = append(msgs, KeyMsg{
				Type:              keyType(e),
				Runes:             []rune{e.Char},
				Alt:               e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED),
				WinKeyEventRecord: &e,
			})
		}

		return msgs, nil
	case coninput.WindowBufferSizeEventRecord:
		return []Msg{WindowSizeMsg{
			Width:  int(e.Size.X),
			Height: int(e.Size.Y),
		}}, nil
	case coninput.MouseEventRecord:
		event := MouseMsg{
			X:                   int(e.MousePositon.X),
			Y:                   int(e.MousePositon.Y),
			Type:                mouseEventType(e),
			Alt:                 e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED),
			Ctrl:                e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED),
			WinMouseEventRecord: &e,
		}

		if event.Type == MouseUnknown {
			return nil, nil
		} else if e.EventFlags&coninput.DOUBLE_CLICK > 0 {
			return []Msg{event, event}, nil
		} else {
			return []Msg{event}, nil
		}
	case coninput.FocusEventRecord, coninput.MenuEventRecord:
		// ignore
	default:
		return nil, fmt.Errorf("unknown record type: %T", e)
	}

	return msgs, nil
}

func mouseEventType(e coninput.MouseEventRecord) MouseEventType {
	switch e.EventFlags {
	case coninput.CLICK, coninput.DOUBLE_CLICK:
		switch {
		case e.ButtonState&coninput.FROM_LEFT_1ST_BUTTON_PRESSED > 0:
			return MouseLeft
		case e.ButtonState&coninput.FROM_LEFT_2ND_BUTTON_PRESSED > 0:
			return MouseMiddle
		case e.ButtonState&coninput.RIGHTMOST_BUTTON_PRESSED > 0:
			return MouseRight
		}
	case coninput.MOUSE_WHEELED:
		if e.WheelDirection > 0 {
			return MouseWheelUp
		} else {
			return MouseWheelDown
		}
	case coninput.MOUSE_HWHEELED:
		return MouseUnknown
	case coninput.MOUSE_MOVED:
		return MouseMotion
	}

	return MouseUnknown
}

func keyType(e coninput.KeyEventRecord) KeyType {
	code := e.VirtualKeyCode

	switch code {
	case coninput.VK_RETURN:
		return KeyEnter
	case coninput.VK_BACK:
		return KeyBackspace
	case coninput.VK_TAB:
		return KeyTab
	case coninput.VK_SPACE:
		return KeyRunes // this could be KeySpace but on unix space also produces KeyRunes
	case coninput.VK_ESCAPE:
		return KeyEscape
	case coninput.VK_UP:
		return KeyUp
	case coninput.VK_DOWN:
		return KeyDown
	case coninput.VK_RIGHT:
		return KeyRight
	case coninput.VK_LEFT:
		return KeyLeft
	case coninput.VK_HOME:
		return KeyHome
	case coninput.VK_END:
		return KeyEnd
	case coninput.VK_PRIOR:
		return KeyPgUp
	case coninput.VK_NEXT:
		return KeyPgDown
	case coninput.VK_DELETE:
		return KeyDelete
	default:
		if e.ControlKeyState&(coninput.LEFT_CTRL_PRESSED|coninput.RIGHT_CTRL_PRESSED) == 0 {
			return KeyRunes
		}

		switch e.Char {
		case '@':
			return KeyCtrlAt
		case '\x01':
			return KeyCtrlA
		case '\x02':
			return KeyCtrlB
		case '\x03':
			return KeyCtrlC
		case '\x04':
			return KeyCtrlD
		case '\x05':
			return KeyCtrlE
		case '\x06':
			return KeyCtrlF
		case '\a':
			return KeyCtrlG
		case '\b':
			return KeyCtrlH
		case '\t':
			return KeyCtrlI
		case '\n':
			return KeyCtrlJ
		case '\v':
			return KeyCtrlK
		case '\f':
			return KeyCtrlL
		case '\r':
			return KeyCtrlM
		case '\x0e':
			return KeyCtrlN
		case '\x0f':
			return KeyCtrlO
		case '\x10':
			return KeyCtrlP
		case '\x11':
			return KeyCtrlQ
		case '\x12':
			return KeyCtrlR
		case '\x13':
			return KeyCtrlS
		case '\x14':
			return KeyCtrlT
		case '\x15':
			return KeyCtrlU
		case '\x16':
			return KeyCtrlV
		case '\x17':
			return KeyCtrlW
		case '\x18':
			return KeyCtrlX
		case '\x19':
			return KeyCtrlY
		case '\x1a':
			return KeyCtrlZ
		case '\x1b':
			return KeyCtrlCloseBracket
		case '\x1c':
			return KeyCtrlBackslash
		case '\x1f':
			return KeyCtrlUnderscore
		}

		switch code {
		case coninput.VK_OEM_4:
			return KeyCtrlOpenBracket
		}

		return KeyRunes
	}
}
