package tea

import (
	"github.com/charmbracelet/x/input"
)

// translateInputEvent translates an input event into a Bubble Tea Msg.
func translateInputEvent(e input.Event) Msg {
	switch e := e.(type) {
	case input.ClipboardEvent:
		switch e.Selection {
		case input.SystemClipboard:
			return ClipboardMsg(e.Content)
		case input.PrimaryClipboard:
			return PrimaryClipboardMsg(e.Content)
		}
	case input.ForegroundColorEvent:
		return ForegroundColorMsg(e)
	case input.BackgroundColorEvent:
		return BackgroundColorMsg(e)
	case input.CursorColorEvent:
		return CursorColorMsg(e)
	case input.CursorPositionEvent:
		return CursorPositionMsg(e)
	case input.FocusEvent:
		return FocusMsg(e)
	case input.BlurEvent:
		return BlurMsg(e)
	case input.KeyPressEvent:
		return KeyPressMsg(e)
	case input.KeyReleaseEvent:
		return KeyReleaseMsg(e)
	case input.MouseClickEvent:
		return MouseClickMsg(e)
	case input.MouseMotionEvent:
		return MouseMotionMsg(e)
	case input.MouseReleaseEvent:
		return MouseReleaseMsg(e)
	case input.MouseWheelEvent:
		return MouseWheelMsg(e)
	case input.PasteEvent:
		return PasteMsg(e)
	case input.PasteStartEvent:
		return PasteStartMsg(e)
	case input.PasteEndEvent:
		return PasteEndMsg(e)
	case input.WindowSizeEvent:
		return WindowSizeMsg(e)
	case input.CapabilityEvent:
		return CapabilityMsg(e)
	case input.TerminalVersionEvent:
		return TerminalVersionMsg(e)
	}
	return nil
}
