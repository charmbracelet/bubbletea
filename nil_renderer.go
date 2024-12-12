package tea

// nilRenderer is a no-op renderer. It implements the Renderer interface but
// doesn't render anything to the terminal.
type nilRenderer struct{}

var _ renderer = nilRenderer{}

// flush implements the Renderer interface.
func (nilRenderer) flush() error { return nil }

// close implements the Renderer interface.
func (nilRenderer) close() error { return nil }

// render implements the Renderer interface.
func (nilRenderer) render(string) {}

// reset implements the Renderer interface.
func (nilRenderer) reset() {}

// update implements the Renderer interface.
func (nilRenderer) update(Msg) {}
