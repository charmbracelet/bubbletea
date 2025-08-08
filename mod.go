package tea

import uv "github.com/charmbracelet/ultraviolet"

// KeyMod represents modifier keys.
type KeyMod = uv.KeyMod

// Modifier keys.
const (
	ModShift = uv.ModShift
	ModAlt   = uv.ModAlt
	ModCtrl  = uv.ModCtrl
	ModMeta  = uv.ModMeta

	// These modifiers are used with the Kitty protocol.
	// XXX: Meta and Super are swapped in the Kitty protocol,
	// this is to preserve compatibility with XTerm modifiers.

	ModHyper = uv.ModHyper
	ModSuper = uv.ModSuper // Windows/Command keys

	// These are key lock states.

	ModCapsLock   = uv.ModCapsLock
	ModNumLock    = uv.ModNumLock
	ModScrollLock = uv.ModScrollLock // Defined in Windows API only
)
