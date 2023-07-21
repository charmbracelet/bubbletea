// Package key provides some types and functions for generating user-definable
// keymappings useful in Bubble Tea components. There are a few different ways
// you can define a keymapping with this package. Here's one example:
//
//	type KeyMap struct {
//	    Up key.Binding
//	    Down key.Binding
//	}
//
//	var DefaultKeyMap = KeyMap{
//	    Up: key.NewBinding(
//	        key.WithKeys("k", "up"),        // actual keybindings
//	        key.WithHelp("↑/k", "move up"), // corresponding help text
//	    ),
//	    Down: key.NewBinding(
//	        key.WithKeys("j", "down"),
//	        key.WithHelp("↓/j", "move down"),
//	    ),
//	}
//
//	func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    switch msg := msg.(type) {
//	    case tea.MsgKey:
//	        switch {
//	        case key.Matches(msg, DefaultKeyMap.Up):
//	            // The user pressed up
//	        case key.Matches(msg, DefaultKeyMap.Down):
//	            // The user pressed down
//	        }
//	    }
//
//	    // ...
//	}
//
// The help information, which is not used in the example above, can be used
// to render help text for keystrokes in your views.
package key

import (
	tea "github.com/rprtr258/bubbletea"
)

// Binding describes a set of keybindings and, optionally, their associated
// help text.
type Binding struct {
	keys     []string
	help     Help
	disabled bool
}

// BindingOpt is an initialization option for a keybinding. It's used as an
// argument to NewBinding.
type BindingOpt func(*Binding)

// NewBinding returns a new keybinding from a set of BindingOpt options.
func NewBinding(opts ...BindingOpt) Binding {
	b := &Binding{}
	for _, opt := range opts {
		opt(b)
	}
	return *b
}

// WithKeys initializes a keybinding with the given keystrokes.
func WithKeys(keys ...string) BindingOpt {
	return func(b *Binding) {
		b.keys = keys
	}
}

// WithHelp initializes a keybinding with the given help text.
func WithHelp(key, desc string) BindingOpt {
	return func(b *Binding) {
		b.help = Help{Key: key, Desc: desc}
	}
}

// WithDisabled initializes a disabled keybinding.
func WithDisabled() BindingOpt {
	return func(b *Binding) {
		b.disabled = true
	}
}

// SetKeys sets the keys for the keybinding.
func (b *Binding) SetKeys(keys ...string) {
	b.keys = keys
}

// Keys returns the keys for the keybinding.
func (b Binding) Keys() []string {
	return b.keys
}

// SetHelp sets the help text for the keybinding.
func (b *Binding) SetHelp(key, desc string) {
	b.help = Help{Key: key, Desc: desc}
}

// Help returns the Help information for the keybinding.
func (b Binding) Help() Help {
	return b.help
}

// Enabled returns whether or not the keybinding is enabled. Disabled
// keybindings won't be activated and won't show up in help. Keybindings are
// enabled by default.
func (b Binding) Enabled() bool {
	return !b.disabled && b.keys != nil
}

// SetEnabled enables or disables the keybinding.
func (b *Binding) SetEnabled(v bool) {
	b.disabled = !v
}

// Unbind removes the keys and help from this binding, effectively nullifying
// it. This is a step beyond disabling it, since applications can enable
// or disable key bindings based on application state.
func (b *Binding) Unbind() {
	b.keys = nil
	b.help = Help{}
}

// Help is help information for a given keybinding.
type Help struct {
	Key  string
	Desc string
}

// Matches checks if the given MsgKey matches the given bindings.
func Matches(k tea.MsgKey, b ...Binding) bool {
	keys := k.String()
	for _, binding := range b {
		for _, v := range binding.keys {
			if keys == v && binding.Enabled() {
				return true
			}
		}
	}
	return false
}
