package tea

import "github.com/charmbracelet/x/ansi"

// PrimaryDeviceAttributesMsg is a message that represents the terminal primary
// device attributes.
type PrimaryDeviceAttributesMsg []int

func parsePrimaryDevAttrs(csi *ansi.CsiSequence) Msg {
	// Primary Device Attributes
	da1 := make(PrimaryDeviceAttributesMsg, len(csi.Params))
	csi.Range(func(i int, p int, hasMore bool) bool {
		if !hasMore {
			da1[i] = p
		}
		return true
	})
	return da1
}
