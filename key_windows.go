//go:build windows
// +build windows

package tea

import (
	"context"
	"fmt"
	"io"

	"github.com/erikgeiser/coninput"
	localereader "github.com/mattn/go-localereader"
	"github.com/muesli/cancelreader"
)

func readInputs(ctx context.Context, msgs chan<- Msg, input io.Reader) error {
	if coninReader, ok := input.(*conInputReader); ok {
		return readConInputs(ctx, msgs, coninReader)
	}

	return readAnsiInputs(ctx, msgs, localereader.NewReader(input))
}

func readConInputs(ctx context.Context, msgsch chan<- Msg, con *conInputReader) error {
	var ps coninput.ButtonState                 // keep track of previous mouse state
	var ws coninput.WindowBufferSizeEventRecord // keep track of the last window size event
	for {
		events, err := coninput.ReadNConsoleInputs(con.conin, 16)
		if err != nil {
			if con.isCanceled() {
				return cancelreader.ErrCanceled
			}
			return fmt.Errorf("read coninput events: %w", err)
		}

		for _, event := range events {
			var msgs []Msg
			switch e := event.Unwrap().(type) {
			case coninput.KeyEventRecord:
				if !e.KeyDown || e.VirtualKeyCode == coninput.VK_SHIFT {
					continue
				}

				for i := 0; i < int(e.RepeatCount); i++ {
					eventKeyType := keyType(e)
					var runes []rune

					// Add the character only if the key type is an actual character and not a control sequence.
					// This mimics the behavior in readAnsiInputs where the character is also removed.
					// We don't need to handle KeySpace here. See the comment in keyType().
					if eventKeyType == KeyRunes {
						runes = []rune{e.Char}
					}

					msgs = append(msgs, KeyMsg{
						Type:  eventKeyType,
						Runes: runes,
						Alt:   e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED),
					})
				}
			case coninput.WindowBufferSizeEventRecord:
				if e != ws {
					ws = e
					msgs = append(msgs, WindowSizeMsg{
						Width:  int(e.Size.X),
						Height: int(e.Size.Y),
					})
				}
			case coninput.MouseEventRecord:
				event := mouseEvent(ps, e)
				if event.Type != MouseUnknown {
					msgs = append(msgs, event)
				}
				ps = e.ButtonState
			case coninput.FocusEventRecord, coninput.MenuEventRecord:
				// ignore
			default: // unknown event
				continue
			}

			// Send all messages to the channel
			for _, msg := range msgs {
				select {
				case msgsch <- msg:
				case <-ctx.Done():
					err := ctx.Err()
					if err != nil {
						return fmt.Errorf("coninput context error: %w", err)
					}
					return nil
				}
			}
		}
	}
}

func mouseEventButton(p, s coninput.ButtonState) (button MouseButton, action MouseAction) {
	btn := p ^ s
	action = MouseActionPress
	if btn&s == 0 {
		action = MouseActionRelease
	}

	if btn == 0 {
		switch {
		case s&coninput.FROM_LEFT_1ST_BUTTON_PRESSED > 0:
			button = MouseButtonLeft
		case s&coninput.FROM_LEFT_2ND_BUTTON_PRESSED > 0:
			button = MouseButtonMiddle
		case s&coninput.RIGHTMOST_BUTTON_PRESSED > 0:
			button = MouseButtonRight
		case s&coninput.FROM_LEFT_3RD_BUTTON_PRESSED > 0:
			button = MouseButtonBackward
		case s&coninput.FROM_LEFT_4TH_BUTTON_PRESSED > 0:
			button = MouseButtonForward
		}
		return button, action
	}

	switch {
	case btn == coninput.FROM_LEFT_1ST_BUTTON_PRESSED: // left button
		button = MouseButtonLeft
	case btn == coninput.RIGHTMOST_BUTTON_PRESSED: // right button
		button = MouseButtonRight
	case btn == coninput.FROM_LEFT_2ND_BUTTON_PRESSED: // middle button
		button = MouseButtonMiddle
	case btn == coninput.FROM_LEFT_3RD_BUTTON_PRESSED: // unknown (possibly mouse backward)
		button = MouseButtonBackward
	case btn == coninput.FROM_LEFT_4TH_BUTTON_PRESSED: // unknown (possibly mouse forward)
		button = MouseButtonForward
	}

	return button, action
}

func mouseEvent(p coninput.ButtonState, e coninput.MouseEventRecord) MouseMsg {
	ev := MouseMsg{
		X:     int(e.MousePositon.X),
		Y:     int(e.MousePositon.Y),
		Alt:   e.ControlKeyState.Contains(coninput.LEFT_ALT_PRESSED | coninput.RIGHT_ALT_PRESSED),
		Ctrl:  e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED),
		Shift: e.ControlKeyState.Contains(coninput.SHIFT_PRESSED),
	}
	switch e.EventFlags {
	case coninput.CLICK, coninput.DOUBLE_CLICK:
		ev.Button, ev.Action = mouseEventButton(p, e.ButtonState)
		if ev.Action == MouseActionRelease {
			ev.Type = MouseRelease
		}
		switch ev.Button { //nolint:exhaustive
		case MouseButtonLeft:
			ev.Type = MouseLeft
		case MouseButtonMiddle:
			ev.Type = MouseMiddle
		case MouseButtonRight:
			ev.Type = MouseRight
		case MouseButtonBackward:
			ev.Type = MouseBackward
		case MouseButtonForward:
			ev.Type = MouseForward
		}
	case coninput.MOUSE_WHEELED:
		if e.WheelDirection > 0 {
			ev.Button = MouseButtonWheelUp
			ev.Type = MouseWheelUp
		} else {
			ev.Button = MouseButtonWheelDown
			ev.Type = MouseWheelDown
		}
	case coninput.MOUSE_HWHEELED:
		if e.WheelDirection > 0 {
			ev.Button = MouseButtonWheelRight
			ev.Type = MouseWheelRight
		} else {
			ev.Button = MouseButtonWheelLeft
			ev.Type = MouseWheelLeft
		}
	case coninput.MOUSE_MOVED:
		ev.Button, _ = mouseEventButton(p, e.ButtonState)
		ev.Action = MouseActionMotion
		ev.Type = MouseMotion
	}

	return ev
}

func keyType(e coninput.KeyEventRecord) KeyType {
	code := e.VirtualKeyCode

	shiftPressed := e.ControlKeyState.Contains(coninput.SHIFT_PRESSED)
	ctrlPressed := e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED | coninput.RIGHT_CTRL_PRESSED)

	switch code { //nolint:exhaustive
	case coninput.VK_RETURN:
		return KeyEnter
	case coninput.VK_BACK:
		return KeyBackspace
	case coninput.VK_TAB:
		if shiftPressed {
			return KeyShiftTab
		}
		return KeyTab
	case coninput.VK_SPACE:
		return KeyRunes // this could be KeySpace but on unix space also produces KeyRunes
	case coninput.VK_ESCAPE:
		return KeyEscape
	case coninput.VK_UP:
		switch {
		case shiftPressed && ctrlPressed:
			return KeyCtrlShiftUp
		case shiftPressed:
			return KeyShiftUp
		case ctrlPressed:
			return KeyCtrlUp
		default:
			return KeyUp
		}
	case coninput.VK_DOWN:
		switch {
		case shiftPressed && ctrlPressed:
			return KeyCtrlShiftDown
		case shiftPressed:
			return KeyShiftDown
		case ctrlPressed:
			return KeyCtrlDown
		default:
			return KeyDown
		}
	case coninput.VK_RIGHT:
		switch {
		case shiftPressed && ctrlPressed:
			return KeyCtrlShiftRight
		case shiftPressed:
			return KeyShiftRight
		case ctrlPressed:
			return KeyCtrlRight
		default:
			return KeyRight
		}
	case coninput.VK_LEFT:
		switch {
		case shiftPressed && ctrlPressed:
			return KeyCtrlShiftLeft
		case shiftPressed:
			return KeyShiftLeft
		case ctrlPressed:
			return KeyCtrlLeft
		default:
			return KeyLeft
		}
	case coninput.VK_HOME:
		switch {
		case shiftPressed && ctrlPressed:
			return KeyCtrlShiftHome
		case shiftPressed:
			return KeyShiftHome
		case ctrlPressed:
			return KeyCtrlHome
		default:
			return KeyHome
		}
	case coninput.VK_END:
		switch {
		case shiftPressed && ctrlPressed:
			return KeyCtrlShiftEnd
		case shiftPressed:
			return KeyShiftEnd
		case ctrlPressed:
			return KeyCtrlEnd
		default:
			return KeyEnd
		}
	case coninput.VK_PRIOR:
		return KeyPgUp
	case coninput.VK_NEXT:
		return KeyPgDown
	case coninput.VK_DELETE:
		return KeyDelete
	case coninput.VK_F1:
		return KeyF1
	case coninput.VK_F2:
		return KeyF2
	case coninput.VK_F3:
		return KeyF3
	case coninput.VK_F4:
		return KeyF4
	case coninput.VK_F5:
		return KeyF5
	case coninput.VK_F6:
		return KeyF6
	case coninput.VK_F7:
		return KeyF7
	case coninput.VK_F8:
		return KeyF8
	case coninput.VK_F9:
		return KeyF9
	case coninput.VK_F10:
		return KeyF10
	case coninput.VK_F11:
		return KeyF11
	case coninput.VK_F12:
		return KeyF12
	case coninput.VK_F13:
		return KeyF13
	case coninput.VK_F14:
		return KeyF14
	case coninput.VK_F15:
		return KeyF15
	case coninput.VK_F16:
		return KeyF16
	case coninput.VK_F17:
		return KeyF17
	case coninput.VK_F18:
		return KeyF18
	case coninput.VK_F19:
		return KeyF19
	case coninput.VK_F20:
		return KeyF20
	default:
		switch {
		case e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED) && e.ControlKeyState.Contains(coninput.RIGHT_ALT_PRESSED):
			// AltGr is pressed, then it's a rune.
			fallthrough
		case !e.ControlKeyState.Contains(coninput.LEFT_CTRL_PRESSED) && !e.ControlKeyState.Contains(coninput.RIGHT_CTRL_PRESSED):
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
			return KeyCtrlOpenBracket // KeyEscape
		case '\x1c':
			return KeyCtrlBackslash
		case '\x1f':
			return KeyCtrlUnderscore
		}

		switch code { //nolint:exhaustive
		case coninput.VK_OEM_4:
			return KeyCtrlOpenBracket
		case coninput.VK_OEM_6:
			return KeyCtrlCloseBracket
		}

		return KeyRunes
	}
}
