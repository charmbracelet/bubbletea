package tea

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// ClipboardMsg is a clipboard read message event. This message is emitted when
// a terminal receives an OSC52 clipboard read message event.
type ClipboardMsg struct {
	Content   string
	Selection byte
}

// Clipboard returns the clipboard selection type. This will be one of the
// following values:
//
//   - c: System clipboard.
//   - p: Primary clipboard (X11/Wayland only).
func (e ClipboardMsg) Clipboard() byte {
	return e.Selection
}

// String returns the string representation of the clipboard message.
func (e ClipboardMsg) String() string {
	return e.Content
}

// setClipboardMsg is an internal message used to set the system clipboard
// using OSC52.
type setClipboardMsg string

// SetClipboard produces a command that sets the system clipboard using OSC52.
// Note that OSC52 is not supported in all terminals.
func SetClipboard(s string) Cmd {
	return func() Msg {
		return setClipboardMsg(s)
	}
}

// readClipboardMsg is an internal message used to read the system clipboard
// using OSC52.
type readClipboardMsg struct{}

// ReadClipboard produces a command that reads the system clipboard using OSC52.
// Note that OSC52 is not supported in all terminals.
func ReadClipboard() Msg {
	return readClipboardMsg{}
}

// setPrimaryClipboardMsg is an internal message used to set the primary
// clipboard using OSC52.
type setPrimaryClipboardMsg string

// SetPrimaryClipboard produces a command that sets the primary clipboard using
// OSC52. Primary clipboard selection is a feature present in X11 and Wayland
// only.
// Note that OSC52 is not supported in all terminals.
func SetPrimaryClipboard(s string) Cmd {
	return func() Msg {
		return setPrimaryClipboardMsg(s)
	}
}

// readPrimaryClipboardMsg is an internal message used to read the primary
// clipboard using OSC52.
type readPrimaryClipboardMsg struct{}

// ReadPrimaryClipboard produces a command that reads the primary clipboard
// using OSC52. Primary clipboard selection is a feature present in X11 and
// Wayland only.
// Note that OSC52 is not supported in all terminals.
func ReadPrimaryClipboard() Msg {
	return readPrimaryClipboardMsg{}
}

// handleSetClipboard writes content to the given clipboard selection, either
// through the program's clipboard backend or as an OSC52 sequence.
func (p *Program) handleSetClipboard(selection ClipboardSelection, content string) {
	if p.clipboard == nil {
		switch selection {
		case uv.PrimaryClipboard:
			p.execute(ansi.SetPrimaryClipboard(content))
		default:
			p.execute(ansi.SetSystemClipboard(content))
		}
		return
	}

	go func() {
		if err := p.clipboard.Set(selection, content); err != nil {
			p.logClipboardError("set", err)
		}
	}()
}

// handleReadClipboard reads the given clipboard selection and delivers the
// result as a [ClipboardMsg].
//
// With a backend the message is produced as soon as the clipboard has been
// read. Otherwise the terminal is queried with OSC52 and replies with a
// [ClipboardMsg] of its own.
func (p *Program) handleReadClipboard(selection ClipboardSelection) {
	if p.clipboard == nil {
		switch selection {
		case uv.PrimaryClipboard:
			p.execute(ansi.RequestPrimaryClipboard)
		default:
			p.execute(ansi.RequestSystemClipboard)
		}
		return
	}

	go func() {
		content, err := p.clipboard.Get(selection)
		if err != nil {
			p.logClipboardError("read", err)
		}
		p.Send(ClipboardMsg{Content: content, Selection: selection})
	}()
}

// logClipboardError logs a clipboard backend error when tracing is enabled.
func (p *Program) logClipboardError(op string, err error) {
	if p.logger != nil {
		p.logger.Printf("clipboard: %s: %v", op, err)
	}
}
