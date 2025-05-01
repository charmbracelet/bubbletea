package tea

import (
	"image/color"

	"github.com/charmbracelet/x/input"
)

// backgroundColorMsg is a message that requests the terminal background color.
type backgroundColorMsg struct{}

// RequestBackgroundColor is a command that requests the terminal background color.
func RequestBackgroundColor() Msg {
	return backgroundColorMsg{}
}

// foregroundColorMsg is a message that requests the terminal foreground color.
type foregroundColorMsg struct{}

// RequestForegroundColor is a command that requests the terminal foreground color.
func RequestForegroundColor() Msg {
	return foregroundColorMsg{}
}

// cursorColorMsg is a message that requests the terminal cursor color.
type cursorColorMsg struct{}

// RequestCursorColor is a command that requests the terminal cursor color.
func RequestCursorColor() Msg {
	return cursorColorMsg{}
}

// setBackgroundColorMsg is a message that sets the terminal background color.
type setBackgroundColorMsg struct{ color.Color }

// SetBackgroundColor is a command that sets the terminal background color.
func SetBackgroundColor(c color.Color) Cmd {
	return func() Msg {
		return setBackgroundColorMsg{c}
	}
}

// setForegroundColorMsg is a message that sets the terminal foreground color.
type setForegroundColorMsg struct{ color.Color }

// SetForegroundColor is a command that sets the terminal foreground color.
func SetForegroundColor(c color.Color) Cmd {
	return func() Msg {
		return setForegroundColorMsg{c}
	}
}

// setCursorColorMsg is a message that sets the terminal cursor color.
type setCursorColorMsg struct{ color.Color }

// SetCursorColor is a command that sets the terminal cursor color.
func SetCursorColor(c color.Color) Cmd {
	return func() Msg {
		return setCursorColorMsg{c}
	}
}

// ForegroundColorMsg represents a foreground color message. This message is
// emitted when the program requests the terminal foreground color with the
// [RequestForegroundColor] Cmd.
type ForegroundColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e ForegroundColorMsg) String() string {
	return input.ForegroundColorEvent(e).String()
}

// IsDark returns whether the color is dark.
func (e ForegroundColorMsg) IsDark() bool {
	return input.ForegroundColorEvent(e).IsDark()
}

// BackgroundColorMsg represents a background color message. This message is
// emitted when the program requests the terminal background color with the
// [RequestBackgroundColor] Cmd.
//
// This is commonly used in [Update.Init] to get the terminal background color
// for style definitions. For that you'll want to call
// [BackgroundColorMsg.IsDark] to determine if the color is dark or light. For
// example:
//
//	func (m Model) Init() (Model, Cmd) {
//	  return m, RequestBackgroundColor()
//	}
//
//	func (m Model) Update(msg Msg) (Model, Cmd) {
//	  switch msg := msg.(type) {
//	  case BackgroundColorMsg:
//	      m.styles = newStyles(msg.IsDark())
//	  }
//	}
type BackgroundColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e BackgroundColorMsg) String() string {
	return input.BackgroundColorEvent(e).String()
}

// IsDark returns whether the color is dark.
func (e BackgroundColorMsg) IsDark() bool {
	return input.BackgroundColorEvent(e).IsDark()
}

// CursorColorMsg represents a cursor color change message. This message is
// emitted when the program requests the terminal cursor color.
type CursorColorMsg struct{ color.Color }

// String returns the hex representation of the color.
func (e CursorColorMsg) String() string {
	return input.CursorColorEvent(e).String()
}

// IsDark returns whether the color is dark.
func (e CursorColorMsg) IsDark() bool {
	return input.CursorColorEvent(e).IsDark()
}
