package tea

type nilRenderer struct{}

func (n nilRenderer) start()                     {}
func (n nilRenderer) stop()                      {}
func (n nilRenderer) kill()                      {}
func (n nilRenderer) write(_ string)             {}
func (n nilRenderer) repaint()                   {}
func (n nilRenderer) clearScreen()               {}
func (n nilRenderer) altScreen() bool            { return false }
func (n nilRenderer) enterAltScreen()            {}
func (n nilRenderer) exitAltScreen()             {}
func (n nilRenderer) showCursor()                {}
func (n nilRenderer) hideCursor()                {}
func (n nilRenderer) execute(_ string)           {}
func (n nilRenderer) bracketedPasteActive() bool { return false }
