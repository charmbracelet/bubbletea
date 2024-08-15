//go:build windows
// +build windows

package tea

import (
	"errors"
	"fmt"
	"unicode/utf16"

	"github.com/charmbracelet/x/ansi"
	xwindows "github.com/charmbracelet/x/windows"
	"golang.org/x/sys/windows"
)

// ReadEvents reads input events from the terminal.
//
// It reads the events available in the input buffer and returns them.
func (d *driver) ReadEvents() ([]Msg, error) {
	events, err := d.handleConInput(readConsoleInput)
	if errors.Is(err, errNotConInputReader) {
		return d.readEvents()
	}
	return events, err
}

var errNotConInputReader = fmt.Errorf("handleConInput: not a conInputReader")

func (d *driver) handleConInput(
	finput func(windows.Handle, []xwindows.InputRecord) (uint32, error),
) ([]Msg, error) {
	cc, ok := d.rd.(*conInputReader)
	if !ok {
		return nil, errNotConInputReader
	}

	// read up to 256 events, this is to allow for sequences events reported as
	// key events.
	var events [256]xwindows.InputRecord
	_, err := finput(cc.conin, events[:])
	if err != nil {
		return nil, fmt.Errorf("read coninput events: %w", err)
	}

	var evs []Msg
	for _, event := range events {
		if e := parseConInputEvent(event, &d.prevMouseState, &d.lastWinsizeEventX, &d.lastWinsizeEventY); e != nil {
			evs = append(evs, e)
		}
	}

	return d.detectConInputQuerySequences(evs), nil
}

// Using ConInput API, Windows Terminal responds to sequence query events with
// KEY_EVENT_RECORDs so we need to collect them and parse them as a single
// sequence.
// Is this a hack?
func (d *driver) detectConInputQuerySequences(events []Msg) []Msg {
	var newEvents []Msg
	start, end := -1, -1

loop:
	for i, e := range events {
		switch e := e.(type) {
		case KeyPressMsg:
			switch e.Rune() {
			case ansi.ESC, ansi.CSI, ansi.OSC, ansi.DCS, ansi.APC:
				// start of a sequence
				if start == -1 {
					start = i
				}
			}
		default:
			break loop
		}
		end = i
	}

	if start == -1 || end <= start {
		return events
	}

	var seq []byte
	for i := start; i <= end; i++ {
		switch e := events[i].(type) {
		case KeyPressMsg:
			seq = append(seq, byte(e.Rune()))
		}
	}

	n, seqevent := parseSequence(seq)
	switch seqevent.(type) {
	case UnknownMsg:
		// We're not interested in unknown events
	default:
		if start+n > len(events) {
			return events
		}
		newEvents = events[:start]
		newEvents = append(newEvents, seqevent)
		newEvents = append(newEvents, events[start+n:]...)
		return d.detectConInputQuerySequences(newEvents)
	}

	return events
}

func parseConInputEvent(event xwindows.InputRecord, buttonState *uint32, windowSizeX, windowSizeY *int16) Msg {
	switch event.EventType {
	case xwindows.KEY_EVENT:
		kevent := event.KeyEvent()
		event := parseWin32InputKeyEvent(kevent.VirtualKeyCode, kevent.VirtualScanCode,
			kevent.Char, kevent.KeyDown, kevent.ControlKeyState, kevent.RepeatCount)

		var key Key
		switch event := event.(type) {
		case KeyPressMsg:
			key = Key(event)
		case KeyReleaseMsg:
			key = Key(event)
		default:
			return nil
		}

		// If the key is not printable, return the event as is
		// (e.g. function keys, arrows, etc.)
		// Otherwise, try to translate it to a rune based on the active keyboard
		// layout.
		if len(key.Runes) == 0 {
			return event
		}

		// Always use US layout for translation
		// This is to follow the behavior of the Kitty Keyboard base layout
		// feature :eye_roll:
		// https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/windows-language-pack-default-values?view=windows-11
		const usLayout = 0x409

		// Translate key to rune
		var keyState [256]byte
		var utf16Buf [16]uint16
		const dontChangeKernelKeyboardLayout = 0x4
		ret := windows.ToUnicodeEx(
			uint32(kevent.VirtualKeyCode),
			uint32(kevent.VirtualScanCode),
			&keyState[0],
			&utf16Buf[0],
			int32(len(utf16Buf)),
			dontChangeKernelKeyboardLayout,
			usLayout,
		)

		// -1 indicates a dead key
		// 0 indicates no translation for this key
		if ret < 1 {
			return event
		}

		runes := utf16.Decode(utf16Buf[:ret])
		if len(runes) != 1 {
			// Key doesn't translate to a single rune
			return event
		}

		key.baseRune = runes[0]
		if kevent.KeyDown {
			return KeyPressMsg(key)
		}

		return KeyReleaseMsg(key)

	case xwindows.WINDOW_BUFFER_SIZE_EVENT:
		wevent := event.WindowBufferSizeEvent()
		if wevent.Size.X != *windowSizeX || wevent.Size.Y != *windowSizeY {
			*windowSizeX, *windowSizeY = wevent.Size.X, wevent.Size.Y
			return WindowSizeMsg{
				Width:  int(wevent.Size.X),
				Height: int(wevent.Size.Y),
			}
		}
	case xwindows.MOUSE_EVENT:
		mevent := event.MouseEvent()
		msg := mouseEvent(*buttonState, mevent)
		*buttonState = mevent.ButtonState
		return msg
	case xwindows.FOCUS_EVENT:
		fevent := event.FocusEvent()
		if fevent.SetFocus {
			return []Msg{FocusMsg{}}
		}
		return []Msg{BlurMsg{}}
	case xwindows.MENU_EVENT:
		// ignore
	}
	return nil
}

func mouseEventButton(p, s uint32) (button MouseButton, isRelease bool) {
	btn := p ^ s
	if btn&s == 0 {
		isRelease = true
	}

	if btn == 0 {
		switch {
		case s&xwindows.FROM_LEFT_1ST_BUTTON_PRESSED > 0:
			button = MouseLeft
		case s&xwindows.FROM_LEFT_2ND_BUTTON_PRESSED > 0:
			button = MouseMiddle
		case s&xwindows.RIGHTMOST_BUTTON_PRESSED > 0:
			button = MouseRight
		case s&xwindows.FROM_LEFT_3RD_BUTTON_PRESSED > 0:
			button = MouseBackward
		case s&xwindows.FROM_LEFT_4TH_BUTTON_PRESSED > 0:
			button = MouseForward
		}
		return
	}

	switch btn {
	case xwindows.FROM_LEFT_1ST_BUTTON_PRESSED: // left button
		button = MouseLeft
	case xwindows.RIGHTMOST_BUTTON_PRESSED: // right button
		button = MouseRight
	case xwindows.FROM_LEFT_2ND_BUTTON_PRESSED: // middle button
		button = MouseMiddle
	case xwindows.FROM_LEFT_3RD_BUTTON_PRESSED: // unknown (possibly mouse backward)
		button = MouseBackward
	case xwindows.FROM_LEFT_4TH_BUTTON_PRESSED: // unknown (possibly mouse forward)
		button = MouseForward
	}

	return
}

func mouseEvent(p uint32, e xwindows.MouseEventRecord) (ev Msg) {
	var mod KeyMod
	var isRelease bool
	if e.ControlKeyState&(xwindows.LEFT_ALT_PRESSED|xwindows.RIGHT_ALT_PRESSED) != 0 {
		mod |= ModAlt
	}
	if e.ControlKeyState&(xwindows.LEFT_CTRL_PRESSED|xwindows.RIGHT_CTRL_PRESSED) != 0 {
		mod |= ModCtrl
	}
	if e.ControlKeyState&(xwindows.SHIFT_PRESSED) != 0 {
		mod |= ModShift
	}

	m := Mouse{
		X:   int(e.MousePositon.X),
		Y:   int(e.MousePositon.Y),
		Mod: mod,
	}

	wheelDirection := int16(highWord(uint32(e.ButtonState)))
	switch e.EventFlags {
	case xwindows.CLICK, xwindows.DOUBLE_CLICK:
		m.Button, isRelease = mouseEventButton(p, e.ButtonState)
	case xwindows.MOUSE_WHEELED:
		if wheelDirection > 0 {
			m.Button = MouseWheelUp
		} else {
			m.Button = MouseWheelDown
		}
	case xwindows.MOUSE_HWHEELED:
		if wheelDirection > 0 {
			m.Button = MouseWheelRight
		} else {
			m.Button = MouseWheelLeft
		}
	case xwindows.MOUSE_MOVED:
		m.Button, _ = mouseEventButton(p, e.ButtonState)
		return MouseMotionMsg(m)
	}

	if isWheel(m.Button) {
		return MouseWheelMsg(m)
	} else if isRelease {
		return MouseReleaseMsg(m)
	}

	return MouseClickMsg(m)
}

func highWord(data uint32) uint16 {
	return uint16((data & 0xFFFF0000) >> 16)
}

func readConsoleInput(console windows.Handle, inputRecords []xwindows.InputRecord) (uint32, error) {
	if len(inputRecords) == 0 {
		return 0, fmt.Errorf("size of input record buffer cannot be zero")
	}

	var read uint32

	err := xwindows.ReadConsoleInput(console, &inputRecords[0], uint32(len(inputRecords)), &read)

	return read, err
}

func peekConsoleInput(console windows.Handle, inputRecords []xwindows.InputRecord) (uint32, error) {
	if len(inputRecords) == 0 {
		return 0, fmt.Errorf("size of input record buffer cannot be zero")
	}

	var read uint32

	err := xwindows.PeekConsoleInput(console, &inputRecords[0], uint32(len(inputRecords)), &read)

	return read, err
}
