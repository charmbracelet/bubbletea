package tea

// KeyMod represents modifier keys.
type KeyMod uint16

// Modifier keys.
const (
	ModShift KeyMod = 1 << iota
	ModAlt
	ModCtrl
	ModMeta

	// These modifiers are used with the Kitty protocol.
	// XXX: Meta and Super are swapped in the Kitty protocol,
	// this is to preserve compatibility with XTerm modifiers.

	ModHyper
	ModSuper // Windows/Command keys

	// These are key lock states.

	ModCapsLock
	ModNumLock
	ModScrollLock // Defined in Windows API only
)

// HasShift reports whether the Shift modifier is set.
func (m KeyMod) HasShift() bool {
	return m&ModShift != 0
}

// HasAlt reports whether the Alt modifier is set.
func (m KeyMod) HasAlt() bool {
	return m&ModAlt != 0
}

// HasCtrl reports whether the Ctrl modifier is set.
func (m KeyMod) HasCtrl() bool {
	return m&ModCtrl != 0
}

// HasMeta reports whether the Meta modifier is set.
func (m KeyMod) HasMeta() bool {
	return m&ModMeta != 0
}

// HasHyper reports whether the Hyper modifier is set.
func (m KeyMod) HasHyper() bool {
	return m&ModHyper != 0
}

// HasSuper reports whether the Super modifier is set.
func (m KeyMod) HasSuper() bool {
	return m&ModSuper != 0
}

// HasCapsLock reports whether the CapsLock key is enabled.
func (m KeyMod) HasCapsLock() bool {
	return m&ModCapsLock != 0
}

// HasNumLock reports whether the NumLock key is enabled.
func (m KeyMod) HasNumLock() bool {
	return m&ModNumLock != 0
}

// HasScrollLock reports whether the ScrollLock key is enabled.
func (m KeyMod) HasScrollLock() bool {
	return m&ModScrollLock != 0
}
