package lipgloss

import (
	"io"

	"github.com/muesli/termenv"
)

// We're manually creating the struct here to avoid initializing the output and
// query the terminal multiple times.
var renderer = &Renderer{
	output: termenv.DefaultOutput(),
}

// Renderer is a lipgloss terminal renderer.
type Renderer struct {
	output            *termenv.Output
	hasDarkBackground *bool
}

// RendererOption is a function that can be used to configure a [Renderer].
type RendererOption func(r *Renderer)

// DefaultRenderer returns the default renderer.
func DefaultRenderer() *Renderer {
	return renderer
}

// SetDefaultRenderer sets the default global renderer.
func SetDefaultRenderer(r *Renderer) {
	renderer = r
}

// NewRenderer creates a new Renderer.
//
// w will be used to determine the terminal's color capabilities.
func NewRenderer(w io.Writer, opts ...termenv.OutputOption) *Renderer {
	r := &Renderer{
		output: termenv.NewOutput(w, opts...),
	}
	return r
}

// Output returns the termenv output.
func (r *Renderer) Output() *termenv.Output {
	return r.output
}

// SetOutput sets the termenv output.
func (r *Renderer) SetOutput(o *termenv.Output) {
	r.output = o
}

// ColorProfile returns the detected termenv color profile.
func (r *Renderer) ColorProfile() termenv.Profile {
	return r.output.Profile
}

// ColorProfile returns the detected termenv color profile.
func ColorProfile() termenv.Profile {
	return renderer.ColorProfile()
}

// SetColorProfile sets the color profile on the renderer. This function exists
// mostly for testing purposes so that you can assure you're testing against
// a specific profile.
//
// Outside of testing you likely won't want to use this function as the color
// profile will detect and cache the terminal's color capabilities and choose
// the best available profile.
//
// Available color profiles are:
//
//	termenv.Ascii     // no color, 1-bit
//	termenv.ANSI      //16 colors, 4-bit
//	termenv.ANSI256   // 256 colors, 8-bit
//	termenv.TrueColor // 16,777,216 colors, 24-bit
//
// This function is thread-safe.
func (r *Renderer) SetColorProfile(p termenv.Profile) {
	r.output.Profile = p
}

// SetColorProfile sets the color profile on the default renderer. This
// function exists mostly for testing purposes so that you can assure you're
// testing against a specific profile.
//
// Outside of testing you likely won't want to use this function as the color
// profile will detect and cache the terminal's color capabilities and choose
// the best available profile.
//
// Available color profiles are:
//
//	termenv.Ascii     // no color, 1-bit
//	termenv.ANSI      //16 colors, 4-bit
//	termenv.ANSI256   // 256 colors, 8-bit
//	termenv.TrueColor // 16,777,216 colors, 24-bit
//
// This function is thread-safe.
func SetColorProfile(p termenv.Profile) {
	renderer.SetColorProfile(p)
}

// HasDarkBackground returns whether or not the terminal has a dark background.
func HasDarkBackground() bool {
	return renderer.HasDarkBackground()
}

// HasDarkBackground returns whether or not the renderer will render to a dark
// background. A dark background can either be auto-detected, or set explicitly
// on the renderer.
func (r *Renderer) HasDarkBackground() bool {
	if r.hasDarkBackground != nil {
		return *r.hasDarkBackground
	}
	return r.output.HasDarkBackground()
}

// SetHasDarkBackground sets the background color detection value for the
// default renderer. This function exists mostly for testing purposes so that
// you can assure you're testing against a specific background color setting.
//
// Outside of testing you likely won't want to use this function as the
// backgrounds value will be automatically detected and cached against the
// terminal's current background color setting.
//
// This function is thread-safe.
func SetHasDarkBackground(b bool) {
	renderer.SetHasDarkBackground(b)
}

// SetHasDarkBackground sets the background color detection value on the
// renderer. This function exists mostly for testing purposes so that you can
// assure you're testing against a specific background color setting.
//
// Outside of testing you likely won't want to use this function as the
// backgrounds value will be automatically detected and cached against the
// terminal's current background color setting.
//
// This function is thread-safe.
func (r *Renderer) SetHasDarkBackground(b bool) {
	r.hasDarkBackground = &b
}
