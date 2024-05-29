package tea

import (
	"context"
	"image/color"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Context represents a Bubble Tea program's context. It is passed to the
// program's Init, Update, and View functions to provide information about the
// program's state and to allow them to interact with the terminal.
type Context interface {
	context.Context

	// SetValue sets a value in the context. This value can be retrieved later
	// using Value.
	SetValue(key, value interface{})

	// BackgroundColor returns the current background color of the terminal.
	// It returns nil if the terminal's doesn't support querying the background
	// color.
	BackgroundColor() color.Color

	// HasLightBackground returns true if the terminal's background color is
	// light. This is useful for determining whether to use light or dark colors
	// in the program's UI.
	HasLightBackground() bool

	// SupportsEnhancedKeyboard reports whether the terminal supports enhanced
	// keyboard keys. On Windows, this means it supports virtual keys like and
	// the Windows Console API. On Unix, this means it supports the Kitty
	// Keyboard Protocol.
	SupportsEnhancedKeyboard() bool

	// NewStyle returns a new Lip Gloss style that is suitable for the program's
	// environment.
	NewStyle() lipgloss.Style

	// ColorProfile returns the terminal's color profile.
	ColorProfile() lipgloss.Profile

	// what else?
}

// ContextKey is a key for storing values in a Context.
type ContextKey struct{ string }

var (
	// ContextKeyBackgroundColor is a key for storing the terminal's background
	// color in a Context.
	ContextKeyBackgroundColor = &ContextKey{"background-color"}

	// ContextKeyKittyFlags is a key for storing the terminal's Kitty flags in a
	// Context.
	ContextKeyKittyFlags = &ContextKey{"kitty-flags"}
)

type teaContext struct {
	context.Context

	values map[interface{}]interface{}
	mtx    sync.RWMutex

	profile    lipgloss.Profile
	hasLightBg bool
}

var _ Context = new(teaContext)

func newContext(ctx context.Context) *teaContext {
	c := new(teaContext)
	c.Context = ctx
	c.values = make(map[interface{}]interface{})
	c.values[ContextKeyKittyFlags] = -1 // Assume no Kitty support by default
	return c
}

// Value returns the value associated with this context for key, or nil if no
// value is associated with key. Successive calls to Value with the same key
// returns the same result.
func (c *teaContext) Value(key interface{}) interface{} {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	if v, ok := c.values[key]; ok {
		return v
	}
	return c.Context.Value(key)
}

// SetValue sets a value in the context. This value can be retrieved later using
// Value.
func (c *teaContext) SetValue(key, value interface{}) {
	c.mtx.Lock()
	c.values[key] = value
	c.mtx.Unlock()
}

func (c *teaContext) BackgroundColor() color.Color {
	if c, ok := c.Value(ContextKeyBackgroundColor).(color.Color); ok {
		return c
	}
	return nil
}

func (c *teaContext) HasLightBackground() bool {
	return c.hasLightBg
}

func (c *teaContext) SupportsEnhancedKeyboard() bool {
	if k, ok := c.Value(ContextKeyKittyFlags).(int); ok {
		return k >= 0
	}
	return false
}

func (c *teaContext) NewStyle() (s lipgloss.Style) {
	return s.ColorProfile(c.profile).HasLightBackground(c.hasLightBg)
}

func (c *teaContext) ColorProfile() lipgloss.Profile {
	return c.profile
}
