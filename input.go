package tea

import (
	"github.com/charmbracelet/tv"
)

// translateInputEvent translates an input event into a Bubble Tea Msg.
func (p *Program) translateInputEvent(e tv.Event) Msg {
	switch e := e.(type) {
	case tv.ClipboardEvent:
		switch e.Selection {
		case tv.SystemClipboard:
			return ClipboardMsg(e.Content)
		case tv.PrimaryClipboard:
			return PrimaryClipboardMsg(e.Content)
		}
	case tv.ForegroundColorEvent:
		return ForegroundColorMsg(e)
	case tv.BackgroundColorEvent:
		return BackgroundColorMsg(e)
	case tv.CursorColorEvent:
		return CursorColorMsg(e)
	case tv.CursorPositionEvent:
		return CursorPositionMsg(e)
	case tv.FocusEvent:
		return FocusMsg(e)
	case tv.BlurEvent:
		return BlurMsg(e)
	case tv.KeyPressEvent:
		return KeyPressMsg(e)
	case tv.KeyReleaseEvent:
		if !isWindows() || p.requestedEnhancements.keyReleases {
			return KeyReleaseMsg(e)
		}
	case tv.MouseClickEvent:
		return MouseClickMsg(e)
	case tv.MouseMotionEvent:
		return MouseMotionMsg(e)
	case tv.MouseReleaseEvent:
		return MouseReleaseMsg(e)
	case tv.MouseWheelEvent:
		return MouseWheelMsg(e)
	case tv.PasteEvent:
		return PasteMsg(e)
	case tv.PasteStartEvent:
		return PasteStartMsg(e)
	case tv.PasteEndEvent:
		return PasteEndMsg(e)
	case tv.WindowSizeEvent:
		return WindowSizeMsg(e)
	case tv.CapabilityEvent:
		return CapabilityMsg(e)
	case tv.TerminalVersionEvent:
		return TerminalVersionMsg(e)
	case tv.KittyEnhancementsEvent:
		return KeyboardEnhancementsMsg{
			kittyFlags:      int(e),
			modifyOtherKeys: p.activeEnhancements.modifyOtherKeys,
		}
	case tv.ModifyOtherKeysEvent:
		return KeyboardEnhancementsMsg{
			modifyOtherKeys: int(e),
			kittyFlags:      p.activeEnhancements.kittyFlags,
		}
	}
	return e
}
