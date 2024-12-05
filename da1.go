package tea

import "github.com/charmbracelet/x/ansi"

// PrimaryDeviceAttributesMsg is a message that represents the terminal primary
// device attributes.
type PrimaryDeviceAttributesMsg []int

func parsePrimaryDevAttrs(csi *ansi.CsiSequence) Msg {
	// Primary Device Attributes
	da1 := make(PrimaryDeviceAttributesMsg, len(csi.Params))
	for i, p := range csi.Params {
		if !p.HasMore() {
			da1[i] = p.Param(0)
		}
	}
	return da1
}
