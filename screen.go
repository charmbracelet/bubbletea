package tea

import "github.com/charmbracelet/x/ansi"

// WindowSizeMsg is used to report the terminal size. It's sent to Update once
// initially and then on every terminal resize. Note that Windows does not
// have support for reporting when resizes occur as it does not support the
// SIGWINCH signal.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// ClearScreen is a special command that tells the program to clear the screen
// before the next update. This can be used to move the cursor to the top left
// of the screen and clear visual clutter when the alt screen is not in use.
//
// Note that it should never be necessary to call ClearScreen() for regular
// redraws.
func ClearScreen() Msg {
	return clearScreenMsg{}
}

// clearScreenMsg is an internal message that signals to clear the screen.
// You can send a clearScreenMsg with ClearScreen.
type clearScreenMsg struct{}

// ModeReportMsg is a message that represents a mode report event (DECRPM).
//
// This is sent by the terminal in response to a request for a terminal mode
// report (DECRQM). It indicates the current setting of a specific terminal
// mode like cursor visibility, mouse tracking, etc.
//
// Example:
//
//	```go
//	func (m model) Init() tea.Cmd {
//	  // Does my terminal support reporting focus events?
//	  return tea.Raw(ansi.RequestModeFocusEvent)
//	}
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	  switch msg := msg.(type) {
//	  case tea.ModeReportMsg:
//	    if msg.Mode == ansi.ModeFocusEvent && !msg.Value.IsNotRecognized() {
//	      // Terminal supports focus events
//	      m.supportsFocus = true
//	    }
//	  }
//	  return m, nil
//	}
//
//	func (m model) View() tea.View {
//	  var view tea.View
//	  view.ReportFocus = m.supportsFocus
//	  view.SetContent(fmt.Sprintf("Terminal supports focus events: %v", m.supportsFocus))
//	  return view
//	}
//	```
//
// See: https://vt100.net/docs/vt510-rm/DECRPM.html
type ModeReportMsg struct {
	// Mode is the mode number.
	Mode ansi.Mode

	// Value is the mode value.
	Value ansi.ModeSetting
}
