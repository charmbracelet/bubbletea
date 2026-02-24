package tea

// RawMsg is a message that contains a string to be printed to the terminal
// without any intermediate processing.
type RawMsg struct {
	Msg any
}

// Raw is a command that prints the given string to the terminal without any
// formatting.
//
// This is intended for advanced use cases where you need to query the terminal
// or send escape sequences directly. Don't use this unless you know what
// you're doing :)
//
// Example:
//
//	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	  switch msg := msg.(type) {
//	  case input.PrimaryDeviceAttributesEvent:
//	    for _, attr := range msg {
//	      if attr == 4 {
//	        // We have Sixel graphics support!
//	        break
//	      }
//	    }
//	  }
//
//	  // Request the terminal primary device attributes to detect Sixel graphics
//	  // support.
//	  return m, tea.Raw(ansi.RequestPrimaryDeviceAttributes)
//	}
func Raw(r any) Cmd {
	return func() Msg {
		return RawMsg{r}
	}
}
