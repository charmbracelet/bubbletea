package tea

import "github.com/charmbracelet/x/input"

// KeyMod represents modifier keys.
type KeyMod = input.KeyMod

// Modifier keys.
const (
	ModShift = input.ModShift
	ModAlt   = input.ModAlt
	ModCtrl  = input.ModCtrl
	ModMeta  = input.ModMeta

	// These modifiers are used with the Kitty protocol.
	// XXX: Meta and Super are swapped in the Kitty protocol,
	// this is to preserve compatibility with XTerm modifiers.

	ModHyper = input.ModHyper
	ModSuper = input.ModSuper // Windows/Command keys

	// These are key lock states.

	ModCapsLock   = input.ModCapsLock
	ModNumLock    = input.ModNumLock
	ModScrollLock = input.ModScrollLock // Defined in Windows API only
)
