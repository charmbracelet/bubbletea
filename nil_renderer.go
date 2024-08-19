package tea

type nilRenderer struct{}

var _ Renderer = nilRenderer{}

func (nilRenderer) Flush() error                    { return nil }
func (nilRenderer) Close() error                    { return nil }
func (nilRenderer) Write([]byte) (int, error)       { return 0, nil }
func (nilRenderer) WriteString(string) (int, error) { return 0, nil }
func (nilRenderer) Repaint()                        {}
func (nilRenderer) ClearScreen()                    {}
func (nilRenderer) AltScreen() bool                 { return false }
func (nilRenderer) EnterAltScreen()                 {}
func (nilRenderer) ExitAltScreen()                  {}
func (nilRenderer) CursorVisibility() bool          { return false }
func (nilRenderer) ShowCursor()                     {}
func (nilRenderer) HideCursor()                     {}
func (nilRenderer) Execute(string)                  {}
func (nilRenderer) InsertAbove(string) error        { return nil }
func (nilRenderer) Resize(int, int)                 {}
