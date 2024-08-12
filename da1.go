package tea

import "github.com/charmbracelet/x/ansi"

// PrimaryDeviceAttributesMsg represents a primary device attributes message.
type PrimaryDeviceAttributesMsg []uint

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
