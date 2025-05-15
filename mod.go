package tea

import "github.com/charmbracelet/tv"

// KeyMod represents modifier keys.
type KeyMod = tv.KeyMod

// Modifier keys.
const (
	ModShift = tv.ModShift
	ModAlt   = tv.ModAlt
	ModCtrl  = tv.ModCtrl
	ModMeta  = tv.ModMeta

	// These modifiers are used with the Kitty protocol.
	// XXX: Meta and Super are swapped in the Kitty protocol,
	// this is to preserve compatibility with XTerm modifiers.

	ModHyper = tv.ModHyper
	ModSuper = tv.ModSuper // Windows/Command keys

	// These are key lock states.

	ModCapsLock   = tv.ModCapsLock
	ModNumLock    = tv.ModNumLock
	ModScrollLock = tv.ModScrollLock // Defined in Windows API only
)
