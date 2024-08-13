package tea

type nilRenderer struct{}

func (nilRenderer) start()                     {}
func (nilRenderer) stop()                      {}
func (nilRenderer) kill()                      {}
func (nilRenderer) write(string)               {}
func (nilRenderer) repaint()                   {}
func (nilRenderer) clearScreen()               {}
func (nilRenderer) altScreen() bool            { return false }
func (nilRenderer) enterAltScreen()            {}
func (nilRenderer) exitAltScreen()             {}
func (nilRenderer) showCursor()                {}
func (nilRenderer) hideCursor()                {}
func (nilRenderer) execute(string)             {}
func (nilRenderer) bracketedPasteActive() bool { return false }
