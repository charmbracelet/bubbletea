package tea

import "github.com/charmbracelet/x/ansi"

// PrimaryDeviceAttributesMsg is a message that represents the terminal primary
// device attributes.
type PrimaryDeviceAttributesMsg []uint

// primaryDeviceAttrsMsg is an internal message that queries the terminal for
// its primary device attributes (DA1).
type primaryDeviceAttrsMsg struct{}

// PrimaryDeviceAttributes is a command that queries the terminal for its
// primary device attributes (DA1). This command is used to determine some of
// the terminal capabilities.
func PrimaryDeviceAttributes() Msg {
	return primaryDeviceAttrsMsg{}
}

func parsePrimaryDevAttrs(csi *ansi.CsiSequence) Msg {
	// Primary Device Attributes
	da1 := make(PrimaryDeviceAttributesMsg, len(csi.Params))
	csi.Range(func(i int, p int, hasMore bool) bool {
		if !hasMore {
			da1[i] = uint(p)
		}
		return true
	})
	return da1
}
