package tea

type nilRenderer struct{}

func (n nilRenderer) Start()                 {}
func (n nilRenderer) Stop()                  {}
func (n nilRenderer) Kill()                  {}
func (n nilRenderer) Write(v string)         {}
func (n nilRenderer) Repaint()               {}
func (n nilRenderer) AltScreen() bool        { return false }
func (n nilRenderer) SetAltScreen(v bool)    {}
func (n nilRenderer) HandleMessages(msg Msg) {}
