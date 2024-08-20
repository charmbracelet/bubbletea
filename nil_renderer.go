package tea

import "io"

// NilRenderer is a no-op renderer. It implements the Renderer interface but
// doesn't render anything to the terminal.
type NilRenderer struct{}

var _ Renderer = NilRenderer{}

func (NilRenderer) SetOutput(io.Writer)             {}
func (NilRenderer) Flush() error                    { return nil }
func (NilRenderer) Close() error                    { return nil }
func (NilRenderer) Write([]byte) (int, error)       { return 0, nil }
func (NilRenderer) WriteString(string) (int, error) { return 0, nil }
func (NilRenderer) Repaint()                        {}
func (NilRenderer) ClearScreen()                    {}
func (NilRenderer) AltScreen() bool                 { return false }
func (NilRenderer) EnterAltScreen()                 {}
func (NilRenderer) ExitAltScreen()                  {}
func (NilRenderer) CursorVisibility() bool          { return false }
func (NilRenderer) ShowCursor()                     {}
func (NilRenderer) HideCursor()                     {}
func (NilRenderer) Execute(string)                  {}
func (NilRenderer) InsertAbove(string) error        { return nil }
func (NilRenderer) Resize(int, int)                 {}
