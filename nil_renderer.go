package tea

// NilRenderer is a no-op renderer. It implements the Renderer interface but
// doesn't render anything to the terminal.
type NilRenderer struct{}

var _ renderer = NilRenderer{}

// flush implements the Renderer interface.
func (NilRenderer) flush() error { return nil }

// close implements the Renderer interface.
func (NilRenderer) close() error { return nil }

// render implements the Renderer interface.
func (NilRenderer) render(string) {}

// reset implements the Renderer interface.
func (NilRenderer) reset() {}

// update implements the Renderer interface.
func (NilRenderer) update(Msg) Cmd { return nil }
