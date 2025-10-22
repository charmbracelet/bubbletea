package tea

// TerminalVersionMsg is a message that represents the terminal version.
type TerminalVersionMsg string

// terminalVersion is an internal message that queries the terminal for its
// version using XTVERSION.
type terminalVersion struct{}

// RequestTerminalVersion is a command that queries the terminal for its
// version using XTVERSION. Note that some terminals may not support this
// command.
func RequestTerminalVersion() Msg {
	return terminalVersion{}
}
