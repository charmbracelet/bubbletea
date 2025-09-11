package tea

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// translateInputEvent translates an input event into a Bubble Tea Msg.
func (p *Program) translateInputEvent(e uv.Event) Msg {
	switch e := e.(type) {
	case uv.ClipboardEvent:
		switch e.Selection {
		case uv.SystemClipboard:
			return ClipboardMsg(e.Content)
		case uv.PrimaryClipboard:
			return PrimaryClipboardMsg(e.Content)
		}
	case uv.ForegroundColorEvent:
		return ForegroundColorMsg(e)
	case uv.BackgroundColorEvent:
		return BackgroundColorMsg(e)
	case uv.CursorColorEvent:
		return CursorColorMsg(e)
	case uv.CursorPositionEvent:
		return CursorPositionMsg(e)
	case uv.FocusEvent:
		return FocusMsg(e)
	case uv.BlurEvent:
		return BlurMsg(e)
	case uv.KeyPressEvent:
		return KeyPressMsg(e)
	case uv.KeyReleaseEvent:
		if !isWindows() || p.requestedEnhancements.keyReleases {
			return KeyReleaseMsg(e)
		}
	case uv.MouseClickEvent:
		return MouseClickMsg(e)
	case uv.MouseMotionEvent:
		return MouseMotionMsg(e)
	case uv.MouseReleaseEvent:
		return MouseReleaseMsg(e)
	case uv.MouseWheelEvent:
		return MouseWheelMsg(e)
	case uv.PasteEvent:
		return PasteMsg(e)
	case uv.PasteStartEvent:
		return PasteStartMsg(e)
	case uv.PasteEndEvent:
		return PasteEndMsg(e)
	case uv.WindowSizeEvent:
		return WindowSizeMsg(e)
	case uv.CapabilityEvent:
		return CapabilityMsg(e)
	case uv.TerminalVersionEvent:
		return TerminalVersionMsg(e)
	case uv.KittyEnhancementsEvent:
		return KeyboardEnhancementsMsg{
			kittyFlags:      int(e),
			modifyOtherKeys: p.activeEnhancements.modifyOtherKeys,
		}
	case uv.ModifyOtherKeysEvent:
		return KeyboardEnhancementsMsg{
			modifyOtherKeys: int(e),
			kittyFlags:      p.activeEnhancements.kittyFlags,
		}
	}
	return e
}
