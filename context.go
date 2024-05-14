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

	// SetValue sets a value on the context. This is useful for storing values
	// that needs to be accessed across multiple functions.
	// You can access the value later using Value.
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

type contextKey struct{ string }

var (
	// ContextKeyColorProfile is the key used to store the terminal's color
	// profile in the context.
	ContextKeyColorProfile = contextKey{"color-profile"}

	// ContextKeyKittyKeyboardFlags is the key used to store the terminal's Kitty
	// Keyboard Protocol flags in the context.
	ContextKeyKittyKeyboardFlags = contextKey{"kitty-keyboard-flags"}

	// ContextKeyBackgroundColor is the key used to store the terminal's background
	// color in the context.
	ContextKeyBackgroundColor = contextKey{"background-color"}

	// ContextKeyHasLightBackground is the key used to store whether the terminal
	// has a light background in the context.
	ContextKeyHasLightBackground = contextKey{"has-light-background"}
)

type teaContext struct {
	context.Context

	values map[interface{}]interface{}
	mtx    sync.Mutex
}

// newContext returns a new teaContext and a cancel function. It wraps the
// provided context with a new context that can be canceled.
func newContext(ctx context.Context) (*teaContext, context.CancelFunc) {
	c := new(teaContext)
	var cancel context.CancelFunc
	c.Context, cancel = context.WithCancel(ctx)
	c.values = make(map[interface{}]interface{})
	c.SetValue(ContextKeyKittyKeyboardFlags, -1)
	return c, cancel
}

func (c *teaContext) BackgroundColor() color.Color {
	if bg, ok := c.Value(ContextKeyBackgroundColor).(color.Color); ok {
		return bg
	}
	return nil
}

func (c *teaContext) HasLightBackground() bool {
	if v, ok := c.Value(ContextKeyHasLightBackground).(bool); ok {
		return v
	}
	return false
}

func (c *teaContext) SupportsEnhancedKeyboard() bool {
	if v, ok := c.Value(ContextKeyKittyKeyboardFlags).(int); ok {
		return v >= 0
	}
	return false
}

func (c *teaContext) NewStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		ColorProfile(c.ColorProfile()).
		HasLightBackground(c.HasLightBackground())
}

func (c *teaContext) ColorProfile() lipgloss.Profile {
	if v, ok := c.Value(ContextKeyColorProfile).(lipgloss.Profile); ok {
		return v
	}
	return lipgloss.TrueColor
}

func (ctx *teaContext) Value(key interface{}) interface{} {
	ctx.mtx.Lock()
	defer ctx.mtx.Unlock()
	if v, ok := ctx.values[key]; ok {
		return v
	}
	return ctx.Context.Value(key)
}

func (ctx *teaContext) SetValue(key, value interface{}) {
	ctx.mtx.Lock()
	defer ctx.mtx.Unlock()
	ctx.values[key] = value
}
