package tea

import (
	"image/color"

	"github.com/charmbracelet/colorprofile"
)

// nilRenderer is a no-op renderer. It implements the Renderer interface but
// doesn't render anything to the terminal.
type nilRenderer struct{}

var _ renderer = nilRenderer{}

// clearScreen implements renderer.
func (n nilRenderer) clearScreen() {}

// repaint implements renderer.
func (n nilRenderer) repaint() {}

// enterAltScreen implements renderer.
func (n nilRenderer) enterAltScreen() {}

// exitAltScreen implements renderer.
func (n nilRenderer) exitAltScreen() {}

// hideCursor implements renderer.
func (n nilRenderer) hideCursor() {}

// insertAbove implements renderer.
func (n nilRenderer) insertAbove(string) {}

// resize implements renderer.
func (n nilRenderer) resize(int, int) {}

// setColorProfile implements renderer.
func (n nilRenderer) setColorProfile(colorprofile.Profile) {}

// showCursor implements renderer.
func (n nilRenderer) showCursor() {}

// flush implements the Renderer interface.
func (nilRenderer) flush(*Program) error { return nil }

// close implements the Renderer interface.
func (nilRenderer) close() error { return nil }

// render implements the Renderer interface.
func (nilRenderer) render(View) {}

// reset implements the Renderer interface.
func (nilRenderer) reset() {}

// writeString implements the Renderer interface.
func (nilRenderer) writeString(string) (int, error) { return 0, nil }

// setCursorColor implements the Renderer interface.
func (n nilRenderer) setCursorColor(color.Color) {}

// setForegroundColor implements the Renderer interface.
func (n nilRenderer) setForegroundColor(color.Color) {}

// setBackgroundColor implements the Renderer interface.
func (n nilRenderer) setBackgroundColor(color.Color) {}

// setWindowTitle implements the Renderer interface.
func (n nilRenderer) setWindowTitle(string) {}

// hit implements the Renderer interface.
func (n nilRenderer) hit(MouseMsg) []Msg  { return nil }
func (n nilRenderer) resetLinesRendered() {}
