package tea

// RawMsg is a message that contains a string to be printed to the terminal
// without any intermediate processing.
type RawMsg struct {
	Msg any
}

// Raw is a command that prints the given string to the terminal without any
// formatting.
func Raw(r any) Cmd {
	return func() Msg {
		return RawMsg{r}
	}
}
